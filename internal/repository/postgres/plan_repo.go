package postgres

import (
	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	"gorm.io/gorm"
)

type planRepo struct {
	db *gorm.DB
}

func NewPlanRepo(db *gorm.DB) *planRepo {
	return &planRepo{db: db}
}

func (r *planRepo) GetByProductID(productID uint) ([]entity.Plan, error) {
	var plans []entity.Plan
	if err := r.db.Where("product_id = ?", productID).Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *planRepo) GetByID(id uint) (*entity.Plan, error) {
	var plan entity.Plan
	if err := r.db.First(&plan, id).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}
