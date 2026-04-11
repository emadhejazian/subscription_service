package database

import (
	"fmt"

	"github.com/emadhejazian/subscription_service/internal/config"
	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens the database connection and runs AutoMigrate.
func Connect(cfg config.Config) (*gorm.DB, error) {
	db, err := open(cfg)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

// ConnectRaw opens the database connection without running any migrations.
// Use this for operations that need to modify the schema before migrating.
func ConnectRaw(cfg config.Config) (*gorm.DB, error) {
	return open(cfg)
}

// open returns a raw GORM connection without running any migrations.
func open(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

// migrate runs AutoMigrate for all entities in dependency order.
func migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entity.Product{},
		&entity.Plan{},
		&entity.Voucher{},
		&entity.Subscription{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}
