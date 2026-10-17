package repository

import (
	"amserver/internal/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type GameRepos struct {
	db *gorm.DB
}

func NewGameRepos(db *gorm.DB) *GameRepos {
	return &GameRepos{db: db}
}

// CRUD

func (r *GameRepos) GetByID(
	ctx context.Context,
	id uint,
) (*models.Games, error) {
	var game models.Games
	err := r.db.WithContext(ctx).
		First(&game, id).
		Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *GameRepos) GetByIDAndUserID(
	ctx context.Context,
	id uint,
	userID uint,
) (*models.Games, error) {
	var game models.Games
	err := r.db.WithContext(ctx).
		Where("creator_id = ?", userID).
		First(&game, id).
		Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *GameRepos) GetByGameCode(
	ctx context.Context,
	gameCode string,
) (*models.Games, error) {
	var game models.Games
	err := r.db.WithContext(ctx).
		Where("game_code = ?", gameCode).
		First(&game).
		Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *GameRepos) Create(
	ctx context.Context,
	game *models.Games,
) error {
	err := r.db.WithContext(ctx).Create(game).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" &&
			pgErr.ConstraintName == "games_game_code_key" {
			return ErrDuplicateGameCode
		}
	}

	return err
}

func (r *GameRepos) DeleteByID(
	ctx context.Context,
	id uint,
) error {
	result := r.db.WithContext(ctx).
		Delete(&models.Games{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrGameNotFound
	}

	return nil
}

func (r *GameRepos) Update(
	ctx context.Context,
	game *models.Games,
) error {
	result := r.db.WithContext(ctx).
		Model(&models.Games{}).
		Where("id = ?", game.ID).
		Updates(map[string]interface{}{
			"game_code":      game.GameCode,
			"start_stage_id": game.StartStageID,
			"name":           game.Name,
			"description":    game.Description,
			"area_geojson":   game.AreaGeojson,
			"setting":        game.Setting,
			"limit_seconds":  game.LimitSeconds,
			"is_public":      game.IsPublic,
			"publish_at":     game.PublishAt,
		})
	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "games_game_code_key" {
			return ErrDuplicateGameCode
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrGameNotFound
	}

	return nil
}
