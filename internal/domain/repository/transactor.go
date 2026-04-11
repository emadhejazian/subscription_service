package repository

// Transactor opens a database transaction and passes transactional repository
// instances to fn. If fn returns an error the transaction is rolled back;
// otherwise it is committed.
//
// The interface lives in the domain layer so the usecase can depend on it
// without importing any database driver.
type Transactor interface {
	WithinTransaction(fn func(
		subscriptionRepo SubscriptionRepository,
		voucherRepo VoucherRepository,
	) error) error
}
