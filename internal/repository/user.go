package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"amserver/internal/models"
)

type UserRepos struct {
	db *gorm.DB
}

func NewUserRepos(db *gorm.DB) *UserRepos {
	return &UserRepos{db: db}
}

func (r *UserRepos) GetByID(
	ctx context.Context,
	id uint,
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

func (r *UserRepos) GetByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	var user models.User

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).
		Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepos) Create(
	ctx context.Context,
	user *models.User,
) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepos) Update(
	ctx context.Context,
	user *models.User,
) error {
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(models.User{
			Name:          user.Name,
			Email:         user.Email,
			Password_hash: user.Password_hash,
		}).Error

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_email_key" {
			return ErrDuplicateEmail
		}
	}

	return err
}
