package model

import (
	"time"
)

type User struct {
	ID             uint       `gorm:"column:id;primaryKey"`
	Name           string     `gorm:"column:name"`
	Password_hash  *string    `gorm:"column:password_hash"`
	Email          *string    `gorm:"column:email"`
	AuthProvider   *string    `gorm:"column:auth_provider"`
	AuthProviderID *string    `gorm:"column:auth_provider_id"`
	CreatedAt      *time.Time `gorm:"column:created_at"`
}

func (User) TableName() string {
	return "users"
}
