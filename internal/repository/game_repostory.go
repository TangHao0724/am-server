package repository

import (
	"amserver/internal/models"
	"context"

	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}
func (r *GameRepository) GetByID(
	ctx context.Context,
	id int,
) (*models.Games_test, error) {
	var game models.Games_test

	err := r.db.WithContext(ctx).
		First(&game, id).
		Error
	if err != nil {
		return nil, err
	}

	return &game, nil
}
func (r *GameRepository) Create(
	ctx context.Context,
	game *models.Games_test,
) error {
	err := r.db.WithContext(ctx).Create(game).Error
	if err != nil {
		return err
	}

	return nil
}
