package entity

import "time"

type DiscountType string

const (
	DiscountFixed   DiscountType = "fixed"
	DiscountPercent DiscountType = "percent"
)

type Voucher struct {
	ID         uint         `gorm:"primaryKey;autoIncrement"`
	Code       string       `gorm:"not null;uniqueIndex"`
	Type       DiscountType `gorm:"not null"`
	Value      float64      `gorm:"not null"`
	ValidFrom  time.Time    `gorm:"not null"`
	ValidUntil time.Time    `gorm:"not null"`
	MaxUses    int          `gorm:"not null"`
	UsedCount  int          `gorm:"not null;default:0"`
	ProductID  *uint        `gorm:"default:null"`
}

func (v *Voucher) ApplyDiscount(price float64) float64 {
	var discounted float64
	switch v.Type {
	case DiscountFixed:
		discounted = price - v.Value
	case DiscountPercent:
		discounted = price * (1 - v.Value/100)
	default:
		return price
	}
	if discounted < 0 {
		return 0
	}
	return discounted
}

func (v *Voucher) IsValid() bool {
	now := time.Now()
	return !now.Before(v.ValidFrom) && !now.After(v.ValidUntil) && v.UsedCount < v.MaxUses
}
