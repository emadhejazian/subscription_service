package repository

import "github.com/emadhejazian/subscription_service/internal/domain/entity"

type VoucherRepository interface {
	GetByCode(code string) (*entity.Voucher, error)
	Save(voucher *entity.Voucher) error
}
