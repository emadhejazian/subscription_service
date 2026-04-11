package usecase

import "github.com/emadhejazian/subscription_service/internal/domain/entity"

type BuyRequest struct {
	UserID      string
	ProductID   uint
	VoucherCode *string
	WithTrial   bool
}

type SubscriptionUsecase interface {
	Buy(req BuyRequest) (*entity.Subscription, error)
	GetByID(id uint) (*entity.Subscription, error)
	GetActiveByUserID(userID string) (*entity.Subscription, error)
	Pause(id uint) (*entity.Subscription, error)
	Unpause(id uint) (*entity.Subscription, error)
	Cancel(id uint) (*entity.Subscription, error)
}
