package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	domainusecase "github.com/emadhejazian/subscription_service/internal/domain/usecase"
)

// ---------------------------------------------------------------------------
// Mock: SubscriptionRepository
// ---------------------------------------------------------------------------

type mockSubRepo struct {
	mock.Mock
}

func (m *mockSubRepo) Create(s *entity.Subscription) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *mockSubRepo) GetByID(id uint) (*entity.Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Subscription), args.Error(1)
}

func (m *mockSubRepo) GetByUserID(userID string) ([]entity.Subscription, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.Subscription), args.Error(1)
}

func (m *mockSubRepo) GetActiveByUserID(userID string) ([]entity.Subscription, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.Subscription), args.Error(1)
}

func (m *mockSubRepo) Save(s *entity.Subscription) error {
	args := m.Called(s)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func defaultProduct() *entity.Product {
	return &entity.Product{ID: 1, Name: "Yoga", Description: "Mindful yoga sessions"}
}

func defaultPlan() *entity.Plan {
	return &entity.Plan{ID: 1, ProductID: 1, Name: "Monthly", DurationMonths: 1, Price: 9.99, TaxRate: 0.19}
}

func activeSubscription() *entity.Subscription {
	now := time.Now()
	end := now.AddDate(0, 1, 0)
	return &entity.Subscription{
		ID:        1,
		UserID:    "user-1",
		ProductID: 1,
		PlanID:    1,
		Status:    entity.StatusActive,
		StartDate: &now,
		EndDate:   &end,
	}
}

func newUC(subRepo *mockSubRepo) *subscriptionUsecase {
	voucherRepo := newMockVoucherRepo()
	txr := &mockTransactor{subscriptionRepo: subRepo, voucherRepo: voucherRepo}
	return NewSubscriptionUsecase(
		subRepo,
		newMockProductRepo(defaultProduct()),
		newMockPlanRepo(defaultPlan()),
		voucherRepo,
		txr,
	)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestBuySubscription_Success(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{}, nil)
	repo.On("Create", mock.AnythingOfType("*entity.Subscription")).
		Return(nil).
		Run(func(args mock.Arguments) {
			// simulate DB auto-increment so GetByID is called with a known ID
			args.Get(0).(*entity.Subscription).ID = 1
		})

	now := time.Now()
	end := now.AddDate(0, 1, 0)
	repo.On("GetByID", uint(1)).Return(&entity.Subscription{
		ID:            1,
		UserID:        "user-1",
		ProductID:     1,
		PlanID:        1,
		Status:        entity.StatusActive,
		OriginalPrice: 9.99,
		TaxAmount:     round2(9.99 * 0.19),
		FinalPrice:    9.99,
		StartDate:     &now,
		EndDate:       &end,
	}, nil)

	uc := newUC(repo)

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 1})

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, entity.StatusActive, sub.Status)
	assert.Equal(t, "user-1", sub.UserID)
	assert.Equal(t, uint(1), sub.ProductID)
	assert.Equal(t, uint(1), sub.PlanID)
	assert.Equal(t, 9.99, sub.OriginalPrice)
	assert.Equal(t, round2(9.99*0.19), sub.TaxAmount)
	assert.NotNil(t, sub.StartDate)
	assert.NotNil(t, sub.EndDate)
	assert.Nil(t, sub.TrialStart)
	repo.AssertExpectations(t)
}

func TestBuySubscription_AlreadyActive(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{*activeSubscription()}, nil)

	uc := newUC(repo)

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 1})

	assert.Nil(t, sub)
	assert.EqualError(t, err, "user already has an active subscription for this product")
	repo.AssertNotCalled(t, "Create", mock.Anything)
	repo.AssertExpectations(t)
}

func TestPause_Success(t *testing.T) {
	sub := activeSubscription()
	repo := new(mockSubRepo)
	repo.On("GetByID", uint(1)).Return(sub, nil)
	repo.On("Save", mock.AnythingOfType("*entity.Subscription")).Return(nil)

	uc := newUC(repo)

	result, err := uc.Pause(1)

	assert.NoError(t, err)
	assert.Equal(t, entity.StatusPaused, result.Status)
	assert.NotNil(t, result.PausedAt)
	repo.AssertExpectations(t)
}

func TestPause_WhenAlreadyPaused(t *testing.T) {
	sub := activeSubscription()
	sub.Status = entity.StatusPaused
	repo := new(mockSubRepo)
	repo.On("GetByID", uint(1)).Return(sub, nil)

	uc := newUC(repo)

	result, err := uc.Pause(1)

	assert.Nil(t, result)
	assert.EqualError(t, err, "subscription cannot be paused in current status")
	repo.AssertNotCalled(t, "Save", mock.Anything)
	repo.AssertExpectations(t)
}

func TestPause_WhenTrialing(t *testing.T) {
	sub := activeSubscription()
	sub.Status = entity.StatusTrialing
	repo := new(mockSubRepo)
	repo.On("GetByID", uint(1)).Return(sub, nil)

	uc := newUC(repo)

	result, err := uc.Pause(1)

	assert.Nil(t, result)
	assert.EqualError(t, err, "subscription cannot be paused in current status")
	repo.AssertNotCalled(t, "Save", mock.Anything)
	repo.AssertExpectations(t)
}

func TestUnpause_ExtendsEndDate(t *testing.T) {
	pausedAt := time.Now().Add(-5 * 24 * time.Hour)
	endDate := time.Now().Add(20 * 24 * time.Hour)
	sub := activeSubscription()
	sub.Status = entity.StatusPaused
	sub.PausedAt = &pausedAt
	sub.EndDate = &endDate

	repo := new(mockSubRepo)
	repo.On("GetByID", uint(1)).Return(sub, nil)
	repo.On("Save", mock.AnythingOfType("*entity.Subscription")).Return(nil)

	uc := newUC(repo)

	result, err := uc.Unpause(1)

	assert.NoError(t, err)
	assert.Equal(t, entity.StatusActive, result.Status)
	assert.Nil(t, result.PausedAt)
	assert.Equal(t, 5, result.PausedDays)

	expectedEnd := endDate.Add(5 * 24 * time.Hour)
	diff := result.EndDate.Sub(expectedEnd).Abs()
	assert.Less(t, diff, time.Minute, "EndDate should be extended by the number of days paused")
	repo.AssertExpectations(t)
}

func TestCancel_Success(t *testing.T) {
	sub := activeSubscription()
	repo := new(mockSubRepo)
	repo.On("GetByID", uint(1)).Return(sub, nil)
	repo.On("Save", mock.AnythingOfType("*entity.Subscription")).Return(nil)

	uc := newUC(repo)

	result, err := uc.Cancel(1)

	assert.NoError(t, err)
	assert.Equal(t, entity.StatusCancelled, result.Status)
	repo.AssertExpectations(t)
}

func TestCancel_WhenAlreadyCancelled(t *testing.T) {
	sub := activeSubscription()
	sub.Status = entity.StatusCancelled
	repo := new(mockSubRepo)
	repo.On("GetByID", uint(1)).Return(sub, nil)

	uc := newUC(repo)

	result, err := uc.Cancel(1)

	assert.Nil(t, result)
	assert.EqualError(t, err, "subscription cannot be cancelled in current status")
	repo.AssertNotCalled(t, "Save", mock.Anything)
	repo.AssertExpectations(t)
}

func TestCancel_NotFound(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByID", uint(99)).Return(nil, errors.New("record not found"))

	uc := newUC(repo)

	result, err := uc.Cancel(99)

	assert.Nil(t, result)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}
