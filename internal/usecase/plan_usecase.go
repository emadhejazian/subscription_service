package usecase

import (
	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	domainrepo "github.com/emadhejazian/subscription_service/internal/domain/repository"
)

type planUsecase struct {
	planRepo domainrepo.PlanRepository
}

func NewPlanUsecase(planRepo domainrepo.PlanRepository) *planUsecase {
	return &planUsecase{planRepo: planRepo}
}

func (u *planUsecase) GetByProductID(productID uint) ([]entity.Plan, error) {
	return u.planRepo.GetByProductID(productID)
}

func (u *planUsecase) GetByID(id uint) (*entity.Plan, error) {
	return u.planRepo.GetByID(id)
}
