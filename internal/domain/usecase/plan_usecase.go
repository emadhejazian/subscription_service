package usecase

import "github.com/emadhejazian/subscription_service/internal/domain/entity"

type PlanUsecase interface {
	GetByProductID(productID uint) ([]entity.Plan, error)
	GetByID(id uint) (*entity.Plan, error)
}
