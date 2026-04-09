package database

import (
	"log"
	"time"

	"github.com/emadhejazian/subscription_service/internal/domain/entity"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	seedProducts(db)
	seedVouchers(db)
}

func seedProducts(db *gorm.DB) {
	var count int64
	db.Model(&entity.Product{}).Count(&count)
	if count > 0 {
		return
	}

	products := []entity.Product{
		{Name: "Monthly Plan", Description: "1-month subscription", DurationMonths: 1, Price: 9.99, TaxRate: 0.19},
		{Name: "Quarterly Plan", Description: "3-month subscription", DurationMonths: 3, Price: 24.99, TaxRate: 0.19},
		{Name: "Semi-Annual Plan", Description: "6-month subscription", DurationMonths: 6, Price: 44.99, TaxRate: 0.19},
		{Name: "Annual Plan", Description: "12-month subscription", DurationMonths: 12, Price: 79.99, TaxRate: 0.19},
	}

	if err := db.Create(&products).Error; err != nil {
		log.Printf("seed: failed to insert products: %v", err)
		return
	}
	log.Printf("seed: inserted %d products", len(products))
}

func seedVouchers(db *gorm.DB) {
	var count int64
	db.Model(&entity.Voucher{}).Count(&count)
	if count > 0 {
		return
	}

	now := time.Now()
	validUntil := now.AddDate(1, 0, 0)

	vouchers := []entity.Voucher{
		{
			Code:       "SAVE10",
			Type:       entity.DiscountPercent,
			Value:      10,
			ValidFrom:  now,
			ValidUntil: validUntil,
			MaxUses:    100,
		},
		{
			Code:       "FLAT5",
			Type:       entity.DiscountFixed,
			Value:      5,
			ValidFrom:  now,
			ValidUntil: validUntil,
			MaxUses:    50,
		},
	}

	if err := db.Create(&vouchers).Error; err != nil {
		log.Printf("seed: failed to insert vouchers: %v", err)
		return
	}
	log.Printf("seed: inserted %d vouchers", len(vouchers))
}
