package usecase

import (
	"errors"

	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	domainrepo "github.com/emadhejazian/subscription_service/internal/domain/repository"
)

type voucherUsecase struct {
	voucherRepo domainrepo.VoucherRepository
}

func NewVoucherUsecase(voucherRepo domainrepo.VoucherRepository) *voucherUsecase {
	return &voucherUsecase{voucherRepo: voucherRepo}
}

func (u *voucherUsecase) Validate(code string, productID uint) (*entity.Voucher, error) {
	voucher, err := u.voucherRepo.GetByCode(code)
	if err != nil {
		return nil, errors.New("voucher not found")
	}
	if !voucher.IsValid() {
		return nil, errors.New("voucher is expired or has reached its usage limit")
	}
	if voucher.ProductID != nil && *voucher.ProductID != productID {
		return nil, errors.New("voucher is not valid for this product")
	}
	return voucher, nil
}

func (u *voucherUsecase) Apply(voucher *entity.Voucher, price float64) float64 {
	return voucher.ApplyDiscount(price)
}
