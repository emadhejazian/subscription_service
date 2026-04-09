package usecase

import (
	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	domainrepo "github.com/emadhejazian/subscription_service/internal/domain/repository"
)

type productUsecase struct {
	productRepo domainrepo.ProductRepository
}

func NewProductUsecase(productRepo domainrepo.ProductRepository) *productUsecase {
	return &productUsecase{productRepo: productRepo}
}

func (u *productUsecase) GetAll() ([]entity.Product, error) {
	return u.productRepo.GetAll()
}

func (u *productUsecase) GetByID(id uint) (*entity.Product, error) {
	return u.productRepo.GetByID(id)
}
