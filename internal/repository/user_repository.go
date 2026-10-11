package repository

import (
	"context"

	"gorm.io/gorm"

	"amserver/internal/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id int64,
) (*models.User, error) {
	var user models.User

	err := r.db.WithContext(ctx).
		First(&user, id).
		Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Create(
	ctx context.Context,
	user *models.User,
) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return err
	}

	return nil
}
