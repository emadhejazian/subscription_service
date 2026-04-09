package entity

import "time"

type Product struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	Name           string    `gorm:"not null"`
	Description    string    `gorm:"type:text"`
	DurationMonths int       `gorm:"not null"`
	Price          float64   `gorm:"not null"`
	TaxRate        float64   `gorm:"not null;default:0"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

func (p *Product) NetPrice() float64 {
	return p.Price * (1 + p.TaxRate)
}
