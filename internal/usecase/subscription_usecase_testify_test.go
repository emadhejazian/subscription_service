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

// newUCWith builds a usecase with a custom voucher repo, wiring the same instance
// into both the usecase and the transactor so mock expectations are shared.
func newUCWith(subRepo *mockSubRepo, voucherRepo *mockVoucherRepo) *subscriptionUsecase {
	txr := &mockTransactor{subscriptionRepo: subRepo, voucherRepo: voucherRepo}
	return NewSubscriptionUsecase(
		subRepo,
		newMockProductRepo(defaultProduct()),
		newMockPlanRepo(defaultPlan()),
		voucherRepo,
		txr,
	)
}

func newUC(subRepo *mockSubRepo) *subscriptionUsecase {
	return newUCWith(subRepo, newMockVoucherRepo())
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

// ---------------------------------------------------------------------------
// Voucher failure tests
// ---------------------------------------------------------------------------

func TestBuySubscription_VoucherNotFound(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{}, nil)

	code := "GHOST"
	uc := newUCWith(repo, newMockVoucherRepo()) // no vouchers registered

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 1, VoucherCode: &code})

	assert.Nil(t, sub)
	assert.EqualError(t, err, "voucher not found")
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestBuySubscription_VoucherExpired(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{}, nil)

	expired := &entity.Voucher{
		Code:       "OLD10",
		Type:       entity.DiscountPercent,
		Value:      10,
		ValidFrom:  time.Now().AddDate(-2, 0, 0),
		ValidUntil: time.Now().AddDate(-1, 0, 0), // expired a year ago
		MaxUses:    100,
	}
	code := "OLD10"
	uc := newUCWith(repo, newMockVoucherRepo(expired))

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 1, VoucherCode: &code})

	assert.Nil(t, sub)
	assert.EqualError(t, err, "voucher is expired or has reached its usage limit")
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestBuySubscription_VoucherMaxUsesReached(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{}, nil)

	exhausted := &entity.Voucher{
		Code:       "FULL",
		Type:       entity.DiscountPercent,
		Value:      10,
		ValidFrom:  time.Now().AddDate(-1, 0, 0),
		ValidUntil: time.Now().AddDate(1, 0, 0),
		MaxUses:    10,
		UsedCount:  10, // already at limit
	}
	code := "FULL"
	uc := newUCWith(repo, newMockVoucherRepo(exhausted))

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 1, VoucherCode: &code})

	assert.Nil(t, sub)
	assert.EqualError(t, err, "voucher is expired or has reached its usage limit")
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestBuySubscription_VoucherProductMismatch(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{}, nil)

	otherProductID := uint(2)
	scoped := &entity.Voucher{
		Code:       "SWIM10",
		Type:       entity.DiscountPercent,
		Value:      10,
		ValidFrom:  time.Now().AddDate(-1, 0, 0),
		ValidUntil: time.Now().AddDate(1, 0, 0),
		MaxUses:    100,
		ProductID:  &otherProductID, // scoped to product 2, but user is buying product 1
	}
	code := "SWIM10"
	uc := newUCWith(repo, newMockVoucherRepo(scoped))

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 1, VoucherCode: &code})

	assert.Nil(t, sub)
	assert.EqualError(t, err, "voucher is not valid for this product")
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

// ---------------------------------------------------------------------------
// Race condition test
// ---------------------------------------------------------------------------

// TestBuySubscription_VoucherDepletedBeforeLock simulates the TOCTOU race:
// the outside IsValid check passes (one use left), but by the time the transaction
// acquires the row lock, another concurrent request has already consumed the last use.
// The in-transaction re-validation must catch this and return an error.
func TestBuySubscription_VoucherDepletedBeforeLock(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{}, nil)

	now := time.Now()
	validUntil := now.AddDate(1, 0, 0)

	// state seen by GetByCode (outside tx): 9 of 10 uses consumed — IsValid passes
	outsideView := &entity.Voucher{
		Code: "LAST1", Type: entity.DiscountPercent, Value: 10,
		ValidFrom: now, ValidUntil: validUntil, MaxUses: 10, UsedCount: 9,
	}
	// state seen by GetByCodeForUpdate (inside tx, after lock): concurrent request
	// consumed the last use between the outside check and this lock acquisition
	lockedView := &entity.Voucher{
		Code: "LAST1", Type: entity.DiscountPercent, Value: 10,
		ValidFrom: now, ValidUntil: validUntil, MaxUses: 10, UsedCount: 10,
	}

	voucherRepo := newMockVoucherRepo(outsideView)
	voucherRepo.lockedByCode = map[string]*entity.Voucher{"LAST1": lockedView}

	code := "LAST1"
	uc := newUCWith(repo, voucherRepo)

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 1, VoucherCode: &code})

	assert.Nil(t, sub)
	assert.EqualError(t, err, "voucher is expired or has reached its usage limit")
	// transaction rolled back: subscription was never created
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

// ---------------------------------------------------------------------------
// Transaction atomicity tests
// ---------------------------------------------------------------------------

// TestBuySubscription_VoucherSaveFailsInTx verifies that if the voucher update
// fails inside the transaction, Create is never called (the whole tx is aborted).
func TestBuySubscription_VoucherSaveFailsInTx(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{}, nil)

	now := time.Now()
	voucher := &entity.Voucher{
		Code: "SAVE10", Type: entity.DiscountPercent, Value: 10,
		ValidFrom: now, ValidUntil: now.AddDate(1, 0, 0), MaxUses: 100,
	}
	voucherRepo := newMockVoucherRepo(voucher)
	voucherRepo.saveErr = errors.New("db connection lost")

	code := "SAVE10"
	uc := newUCWith(repo, voucherRepo)

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 1, VoucherCode: &code})

	assert.Nil(t, sub)
	assert.EqualError(t, err, "db connection lost")
	// Create was never reached — in a real DB the partial voucher write would also roll back
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

// TestBuySubscription_CreateFailsInTx verifies that if subscription Create fails,
// the error is propagated. In a real DB the voucher UsedCount increment also rolls back.
func TestBuySubscription_CreateFailsInTx(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{}, nil)
	repo.On("Create", mock.AnythingOfType("*entity.Subscription")).Return(errors.New("unique constraint violated"))

	uc := newUC(repo)

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 1})

	assert.Nil(t, sub)
	assert.EqualError(t, err, "unique constraint violated")
	repo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Plan / product validation failure tests
// ---------------------------------------------------------------------------

func TestBuySubscription_PlanNotBelongToProduct(t *testing.T) {
	repo := new(mockSubRepo)
	repo.On("GetByUserID", "user-1").Return([]entity.Subscription{}, nil)

	// plan belongs to product 2 but request targets product 1
	wrongPlan := &entity.Plan{ID: 99, ProductID: 2, Name: "Monthly", DurationMonths: 1, Price: 9.99, TaxRate: 0.19}
	voucherRepo := newMockVoucherRepo()
	txr := &mockTransactor{subscriptionRepo: repo, voucherRepo: voucherRepo}
	uc := NewSubscriptionUsecase(repo, newMockProductRepo(defaultProduct()), newMockPlanRepo(wrongPlan), voucherRepo, txr)

	sub, err := uc.Buy(domainusecase.BuyRequest{UserID: "user-1", ProductID: 1, PlanID: 99})

	assert.Nil(t, sub)
	assert.EqualError(t, err, "plan does not belong to this product")
	repo.AssertNotCalled(t, "Create", mock.Anything)
}
