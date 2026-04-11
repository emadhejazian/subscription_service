package postgres

import (
	domainrepo "github.com/emadhejazian/subscription_service/internal/domain/repository"
	"gorm.io/gorm"
)

type transactor struct {
	db *gorm.DB
}

func NewTransactor(db *gorm.DB) *transactor {
	return &transactor{db: db}
}

// WithinTransaction runs fn inside a single GORM transaction.
// Both repo arguments passed to fn share the same *gorm.DB transaction,
// so any error returned by fn causes an automatic ROLLBACK; a nil return commits.
func (t *transactor) WithinTransaction(fn func(
	subscriptionRepo domainrepo.SubscriptionRepository,
	voucherRepo domainrepo.VoucherRepository,
) error) error {
	return t.db.Transaction(func(tx *gorm.DB) error {
		return fn(NewSubscriptionRepo(tx), NewVoucherRepo(tx))
	})
}
