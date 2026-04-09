package entity

import "time"

type SubscriptionStatus string

const (
	StatusTrialing  SubscriptionStatus = "trialing"
	StatusActive    SubscriptionStatus = "active"
	StatusPaused    SubscriptionStatus = "paused"
	StatusCancelled SubscriptionStatus = "cancelled"
)

type Subscription struct {
	ID             uint               `gorm:"primaryKey;autoIncrement"`
	UserID         string             `gorm:"not null;index"`
	ProductID      uint               `gorm:"not null"`
	Product        Product            `gorm:"foreignKey:ProductID"`
	Status         SubscriptionStatus `gorm:"not null;default:'trialing'"`
	OriginalPrice  float64            `gorm:"not null"`
	DiscountAmount float64            `gorm:"not null;default:0"`
	FinalPrice     float64            `gorm:"not null"`
	TaxAmount      float64            `gorm:"not null;default:0"`
	TrialStart     *time.Time         `gorm:"default:null"`
	TrialEnd       *time.Time         `gorm:"default:null"`
	StartDate      *time.Time         `gorm:"default:null"`
	EndDate        *time.Time         `gorm:"default:null"`
	PausedAt       *time.Time         `gorm:"default:null"`
	PausedDays     int                `gorm:"not null;default:0"`
	VoucherID      *uint              `gorm:"default:null"`
	CreatedAt      time.Time          `gorm:"autoCreateTime"`
	UpdatedAt      time.Time          `gorm:"autoUpdateTime"`
}

func (s *Subscription) CanPause() bool {
	return s.Status == StatusActive
}

func (s *Subscription) CanCancel() bool {
	return s.Status == StatusActive || s.Status == StatusPaused || s.Status == StatusTrialing
}

func (s *Subscription) CanUnpause() bool {
	return s.Status == StatusPaused
}

func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive
}
