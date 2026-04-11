package postgres

import (
	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	"gorm.io/gorm"
)

type productRepo struct {
	db *gorm.DB
}

func NewProductRepo(db *gorm.DB) *productRepo {
	return &productRepo{db: db}
}

func (r *productRepo) GetAll() ([]entity.Product, error) {
	var products []entity.Product
	if err := r.db.Preload("Plans").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepo) GetByID(id uint) (*entity.Product, error) {
	var product entity.Product
	if err := r.db.Preload("Plans").First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
