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

// Reset drops all tables, recreates them, and seeds fresh data.
// It accepts a raw (unmigrated) DB connection so AutoMigrate can run clean.
func Reset(rawDB *gorm.DB) {
	log.Println("reset: dropping tables...")
	rawDB.Exec("DROP TABLE IF EXISTS subscriptions, vouchers, plans, products CASCADE")
	log.Println("reset: tables dropped")

	rawDB.AutoMigrate(
		&entity.Product{},
		&entity.Plan{},
		&entity.Voucher{},
		&entity.Subscription{},
	)
	log.Println("reset: tables recreated")
	Seed(rawDB)
}

func seedProducts(db *gorm.DB) {
	var count int64
	db.Model(&entity.Product{}).Count(&count)
	if count > 0 {
		return
	}

	type courseSpec struct {
		name        string
		description string
		prices      [4]float64 // monthly, quarterly, semi-annual, annual
	}

	courses := []courseSpec{
		{"Yoga", "Mindful yoga sessions for all levels", [4]float64{9.99, 24.99, 44.99, 79.99}},
		{"Swimming", "Professional swimming coaching", [4]float64{14.99, 34.99, 59.99, 99.99}},
		{"CrossFit", "High-intensity CrossFit training", [4]float64{19.99, 49.99, 84.99, 149.99}},
		{"Cycling", "Indoor and outdoor cycling programs", [4]float64{12.99, 29.99, 54.99, 94.99}},
	}

	planDefs := [4]struct {
		name    string
		months  int
	}{
		{"Monthly", 1},
		{"Quarterly", 3},
		{"Semi-Annual", 6},
		{"Annual", 12},
	}

	for _, c := range courses {
		product := entity.Product{Name: c.name, Description: c.description}
		if err := db.Create(&product).Error; err != nil {
			log.Printf("seed: failed to insert product %s: %v", c.name, err)
			continue
		}

		for i, pd := range planDefs {
			plan := entity.Plan{
				ProductID:      product.ID,
				Name:           pd.name,
				DurationMonths: pd.months,
				Price:          c.prices[i],
				TaxRate:        0.19,
			}
			if err := db.Create(&plan).Error; err != nil {
				log.Printf("seed: failed to insert plan %s for %s: %v", pd.name, c.name, err)
			}
		}
	}

	log.Printf("seed: inserted %d products with 4 plans each", len(courses))
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
