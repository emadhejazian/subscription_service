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
	planRepo         domainrepo.PlanRepository
	voucherRepo      domainrepo.VoucherRepository
	transactor       domainrepo.Transactor
}

func NewSubscriptionUsecase(
	subscriptionRepo domainrepo.SubscriptionRepository,
	productRepo domainrepo.ProductRepository,
	planRepo domainrepo.PlanRepository,
	voucherRepo domainrepo.VoucherRepository,
	transactor domainrepo.Transactor,
) *subscriptionUsecase {
	return &subscriptionUsecase{
		subscriptionRepo: subscriptionRepo,
		productRepo:      productRepo,
		planRepo:         planRepo,
		voucherRepo:      voucherRepo,
		transactor:       transactor,
	}
}

func (u *subscriptionUsecase) Buy(req domainusecase.BuyRequest) (*entity.Subscription, error) {
	userID := req.UserID
	productID := req.ProductID
	planID := req.PlanID
	voucherCode := req.VoucherCode
	withTrial := req.WithTrial

	// idempotency: one active subscription per product per user
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

	// validate product exists
	if _, err := u.productRepo.GetByID(productID); err != nil {
		return nil, errors.New("product not found")
	}

	// fetch plan for pricing
	plan, err := u.planRepo.GetByID(planID)
	if err != nil {
		return nil, errors.New("plan not found")
	}
	if plan.ProductID != productID {
		return nil, errors.New("plan does not belong to this product")
	}

	originalPrice := plan.Price
	discountAmount := 0.0
	finalPrice := originalPrice
	var voucherID *uint
	var voucher *entity.Voucher

	// validate voucher outside the transaction (read-only checks)
	if voucherCode != nil && *voucherCode != "" {
		v, err := u.voucherRepo.GetByCode(*voucherCode)
		if err != nil {
			return nil, errors.New("voucher not found")
		}
		if !v.IsValid() {
			return nil, errors.New("voucher is expired or has reached its usage limit")
		}
		if v.ProductID != nil && *v.ProductID != productID {
			return nil, errors.New("voucher is not valid for this product")
		}

		discounted := v.ApplyDiscount(originalPrice)
		discountAmount = originalPrice - discounted
		finalPrice = discounted
		voucher = v
		voucherID = &v.ID
	}

	taxAmount := round2(finalPrice * plan.TaxRate)

	now := time.Now()
	sub := &entity.Subscription{
		UserID:         userID,
		ProductID:      productID,
		PlanID:         planID,
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
		endDate := trialEnd.AddDate(0, plan.DurationMonths, 0)

		sub.Status = entity.StatusTrialing
		sub.TrialStart = &trialStart
		sub.TrialEnd = &trialEnd
		sub.StartDate = &startDate
		sub.EndDate = &endDate
	} else {
		startDate := now
		endDate := now.AddDate(0, plan.DurationMonths, 0)

		sub.Status = entity.StatusActive
		sub.StartDate = &startDate
		sub.EndDate = &endDate
	}

	// atomic: lock the voucher row, re-validate, increment usage, and create the subscription
	// in one transaction. SELECT FOR UPDATE blocks any concurrent transaction that also tries
	// to lock the same voucher row, eliminating the TOCTOU race on UsedCount.
	if err := u.transactor.WithinTransaction(func(
		txSubRepo domainrepo.SubscriptionRepository,
		txVoucherRepo domainrepo.VoucherRepository,
	) error {
		if voucher != nil {
			// re-fetch with a row-level lock — reads the latest UsedCount under the lock
			locked, err := txVoucherRepo.GetByCodeForUpdate(voucher.Code)
			if err != nil {
				return err
			}
			// re-validate: another request may have consumed the last use between our
			// outside check and now
			if !locked.IsValid() {
				return errors.New("voucher is expired or has reached its usage limit")
			}
			locked.UsedCount++
			if err := txVoucherRepo.Save(locked); err != nil {
				return err
			}
		}
		return txSubRepo.Create(sub)
	}); err != nil {
		return nil, err
	}

	return u.subscriptionRepo.GetByID(sub.ID)
}

func (u *subscriptionUsecase) GetByID(id uint) (*entity.Subscription, error) {
	return u.subscriptionRepo.GetByID(id)
}

func (u *subscriptionUsecase) GetActiveByUserID(userID string) ([]entity.Subscription, error) {
	subs, err := u.subscriptionRepo.GetActiveByUserID(userID)
	if err != nil {
		return nil, errors.New("no active subscriptions found")
	}
	return subs, nil
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
