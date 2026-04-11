package repository

import "github.com/emadhejazian/subscription_service/internal/domain/entity"

type VoucherRepository interface {
	GetByCode(code string) (*entity.Voucher, error)
	// GetByCodeForUpdate fetches the voucher and holds a row-level lock (SELECT FOR UPDATE)
	// for the duration of the enclosing transaction. Use this inside a transaction when
	// you need to check-then-write atomically (e.g. validate IsValid then increment UsedCount).
	GetByCodeForUpdate(code string) (*entity.Voucher, error)
	Save(voucher *entity.Voucher) error
}
