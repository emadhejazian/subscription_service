package repository

import "github.com/emadhejazian/subscription_service/internal/domain/entity"

type SubscriptionRepository interface {
	Create(subscription *entity.Subscription) error
	GetByID(id uint) (*entity.Subscription, error)
	GetByUserID(userID string) ([]entity.Subscription, error)
	GetActiveByUserID(userID string) ([]entity.Subscription, error)
	Save(subscription *entity.Subscription) error
}
