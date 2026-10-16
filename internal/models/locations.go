package models

import (
	"time"
)

type Locations struct {
	ID          uint       `gorm:"column:id;primaryKey"`
	CreatorID   uint       `gorm:"column:creator_id"`
	Name        *string    `gorm:"column:name"`
	Description *string    `gorm:"column:description"`
	Lon         *float64   `gorm:"column:lon"`
	Lat         *float64   `gorm:"column:lat"`
	PublicAt    *time.Time `gorm:"column:public_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (Locations) TableName() string {
	return "locations"
}
