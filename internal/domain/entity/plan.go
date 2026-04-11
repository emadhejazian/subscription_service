package entity

import "time"

type Plan struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	ProductID      uint      `gorm:"not null;index"`
	Name           string    `gorm:"not null"`
	DurationMonths int       `gorm:"not null"`
	Price          float64   `gorm:"not null"`
	TaxRate        float64   `gorm:"not null;default:0"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

func (p *Plan) NetPrice() float64 {
	return p.Price * (1 + p.TaxRate)
}
