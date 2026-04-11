package usecase

import (
	"errors"
	"math"
	"time"

	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	domainrepo "github.com/emadhejazian/subscription_service/internal/domain/repository"
	domainusecase "github.com/emadhejazian/subscription_service/internal/domain/usecase"
)

type subscriptionUsecase struct {
	subscriptionRepo domainrepo.SubscriptionRepository
	productRepo      domainrepo.ProductRepository
	voucherRepo      domainrepo.VoucherRepository
}

func NewSubscriptionUsecase(
	subscriptionRepo domainrepo.SubscriptionRepository,
	productRepo domainrepo.ProductRepository,
	voucherRepo domainrepo.VoucherRepository,
) *subscriptionUsecase {
	return &subscriptionUsecase{
		subscriptionRepo: subscriptionRepo,
		productRepo:      productRepo,
		voucherRepo:      voucherRepo,
	}
}

func (u *subscriptionUsecase) Buy(req domainusecase.BuyRequest) (*entity.Subscription, error) {
	userID := req.UserID
	productID := req.ProductID
	voucherCode := req.VoucherCode
	withTrial := req.WithTrial

	existing, err := u.subscriptionRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	for _, s := range existing {
		if s.ProductID == productID {
			switch s.Status {
			case entity.StatusActive, entity.StatusTrialing, entity.StatusPaused:
				return nil, errors.New("user already has an active subscription for this product")
			}
		}
	}

	product, err := u.productRepo.GetByID(productID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	originalPrice := product.Price
	discountAmount := 0.0
	finalPrice := originalPrice
	var voucherID *uint

	if voucherCode != nil && *voucherCode != "" {
		voucher, err := u.voucherRepo.GetByCode(*voucherCode)
		if err != nil {
			return nil, errors.New("voucher not found")
		}
		if !voucher.IsValid() {
			return nil, errors.New("voucher is expired or has reached its usage limit")
		}
		if voucher.ProductID != nil && *voucher.ProductID != productID {
			return nil, errors.New("voucher is not valid for this product")
		}

		discounted := voucher.ApplyDiscount(originalPrice)
		discountAmount = originalPrice - discounted
		finalPrice = discounted

		voucher.UsedCount++
		if err := u.voucherRepo.Save(voucher); err != nil {
			return nil, err
		}
		voucherID = &voucher.ID
	}

	taxAmount := round2(finalPrice * product.TaxRate)

	now := time.Now()
	sub := &entity.Subscription{
		UserID:         userID,
		ProductID:      productID,
		OriginalPrice:  round2(originalPrice),
		DiscountAmount: round2(discountAmount),
		FinalPrice:     round2(finalPrice),
		TaxAmount:      taxAmount,
		VoucherID:      voucherID,
	}

	if withTrial {
		trialStart := now
		trialEnd := now.AddDate(0, 1, 0)
		startDate := trialEnd
		endDate := trialEnd.AddDate(0, product.DurationMonths, 0)

		sub.Status = entity.StatusTrialing
		sub.TrialStart = &trialStart
		sub.TrialEnd = &trialEnd
		sub.StartDate = &startDate
		sub.EndDate = &endDate
	} else {
		startDate := now
		endDate := now.AddDate(0, product.DurationMonths, 0)

		sub.Status = entity.StatusActive
		sub.StartDate = &startDate
		sub.EndDate = &endDate
	}

	if err := u.subscriptionRepo.Create(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (u *subscriptionUsecase) GetByID(id uint) (*entity.Subscription, error) {
	return u.subscriptionRepo.GetByID(id)
}

func (u *subscriptionUsecase) GetActiveByUserID(userID string) (*entity.Subscription, error) {
	sub, err := u.subscriptionRepo.GetActiveByUserID(userID)
	if err != nil {
		return nil, errors.New("no active subscription found")
	}
	return sub, nil
}

func (u *subscriptionUsecase) Pause(id uint) (*entity.Subscription, error) {
	sub, err := u.subscriptionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if !sub.CanPause() {
		return nil, errors.New("subscription cannot be paused in current status")
	}

	now := time.Now()
	sub.Status = entity.StatusPaused
	sub.PausedAt = &now

	if err := u.subscriptionRepo.Save(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (u *subscriptionUsecase) Unpause(id uint) (*entity.Subscription, error) {
	sub, err := u.subscriptionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if !sub.CanUnpause() {
		return nil, errors.New("subscription cannot be unpaused in current status")
	}

	now := time.Now()
	daysPaused := int(math.Round(now.Sub(*sub.PausedAt).Hours() / 24))

	sub.PausedDays += daysPaused
	if sub.EndDate != nil {
		extended := sub.EndDate.AddDate(0, 0, daysPaused)
		sub.EndDate = &extended
	}

	sub.Status = entity.StatusActive
	sub.PausedAt = nil

	if err := u.subscriptionRepo.Save(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (u *subscriptionUsecase) Cancel(id uint) (*entity.Subscription, error) {
	sub, err := u.subscriptionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if !sub.CanCancel() {
		return nil, errors.New("subscription cannot be cancelled in current status")
	}

	sub.Status = entity.StatusCancelled

	if err := u.subscriptionRepo.Save(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
