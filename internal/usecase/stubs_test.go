package usecase

import (
	"errors"

	"github.com/emadhejazian/subscription_service/internal/domain/entity"
)

// --- product repo stub ---

type mockProductRepo struct {
	products map[uint]*entity.Product
}

func newMockProductRepo(products ...*entity.Product) *mockProductRepo {
	m := &mockProductRepo{products: make(map[uint]*entity.Product)}
	for _, p := range products {
		m.products[p.ID] = p
	}
	return m
}

func (m *mockProductRepo) GetAll() ([]entity.Product, error) {
	var out []entity.Product
	for _, p := range m.products {
		out = append(out, *p)
	}
	return out, nil
}

func (m *mockProductRepo) GetByID(id uint) (*entity.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return p, nil
}

// --- plan repo stub ---

type mockPlanRepo struct {
	plans map[uint]*entity.Plan
}

func newMockPlanRepo(plans ...*entity.Plan) *mockPlanRepo {
	m := &mockPlanRepo{plans: make(map[uint]*entity.Plan)}
	for _, p := range plans {
		m.plans[p.ID] = p
	}
	return m
}

func (m *mockPlanRepo) GetByProductID(productID uint) ([]entity.Plan, error) {
	var out []entity.Plan
	for _, p := range m.plans {
		if p.ProductID == productID {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (m *mockPlanRepo) GetByID(id uint) (*entity.Plan, error) {
	p, ok := m.plans[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return p, nil
}

// --- voucher repo stub ---

type mockVoucherRepo struct {
	byCode map[string]*entity.Voucher
}

func newMockVoucherRepo(vouchers ...*entity.Voucher) *mockVoucherRepo {
	m := &mockVoucherRepo{byCode: make(map[string]*entity.Voucher)}
	for _, v := range vouchers {
		m.byCode[v.Code] = v
	}
	return m
}

func (m *mockVoucherRepo) GetByCode(code string) (*entity.Voucher, error) {
	v, ok := m.byCode[code]
	if !ok {
		return nil, errors.New("record not found")
	}
	return v, nil
}

func (m *mockVoucherRepo) Save(v *entity.Voucher) error { return nil }
