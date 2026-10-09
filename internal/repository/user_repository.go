package repository

import (
	"context"

	"gorm.io/gorm"

	"amserver/internal/model"
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
) (*model.User, error) {
	var user model.User

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
	user *model.User,
) error {
	return r.db.WithContext(ctx).Create(user).Error
}
