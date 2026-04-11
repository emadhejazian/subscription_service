package postgres

import (
	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	"gorm.io/gorm"
)

type subscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepo(db *gorm.DB) *subscriptionRepo {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(subscription *entity.Subscription) error {
	return r.db.Create(subscription).Error
}

func (r *subscriptionRepo) GetByID(id uint) (*entity.Subscription, error) {
	var subscription entity.Subscription
	if err := r.db.Preload("Product").First(&subscription, id).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepo) GetByUserID(userID string) ([]entity.Subscription, error) {
	var subscriptions []entity.Subscription
	if err := r.db.Preload("Product").Where("user_id = ?", userID).Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (r *subscriptionRepo) GetActiveByUserID(userID string) ([]entity.Subscription, error) {
	var subscriptions []entity.Subscription
	err := r.db.Preload("Product").
		Where("user_id = ? AND status IN ?", userID, []string{"active", "paused", "trialing"}).
		Order("created_at DESC").
		Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (r *subscriptionRepo) Save(subscription *entity.Subscription) error {
	return r.db.Save(subscription).Error
}
