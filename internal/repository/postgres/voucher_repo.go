package postgres

import (
	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	"gorm.io/gorm"
)

type voucherRepo struct {
	db *gorm.DB
}

func NewVoucherRepo(db *gorm.DB) *voucherRepo {
	return &voucherRepo{db: db}
}

func (r *voucherRepo) GetByCode(code string) (*entity.Voucher, error) {
	var voucher entity.Voucher
	if err := r.db.Where("code = ?", code).First(&voucher).Error; err != nil {
		return nil, err
	}
	return &voucher, nil
}

func (r *voucherRepo) Save(voucher *entity.Voucher) error {
	return r.db.Save(voucher).Error
}
