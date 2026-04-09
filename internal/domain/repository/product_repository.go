package repository

import "github.com/emadhejazian/subscription_service/internal/domain/entity"

type ProductRepository interface {
	GetAll() ([]entity.Product, error)
	GetByID(id uint) (*entity.Product, error)
}
