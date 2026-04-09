package usecase

import "github.com/emadhejazian/subscription_service/internal/domain/entity"

type ProductUsecase interface {
	GetAll() ([]entity.Product, error)
	GetByID(id uint) (*entity.Product, error)
}
