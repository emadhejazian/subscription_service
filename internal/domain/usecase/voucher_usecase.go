package usecase

import "github.com/emadhejazian/subscription_service/internal/domain/entity"

type VoucherUsecase interface {
	Validate(code string, productID uint) (*entity.Voucher, error)
	Apply(voucher *entity.Voucher, price float64) float64
}
