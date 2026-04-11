package entity

import "time"

type Product struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"not null"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	Plans       []Plan    `gorm:"foreignKey:ProductID"`
}
