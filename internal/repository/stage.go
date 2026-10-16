package repository

import (
	"amserver/internal/models"
	"context"

	"gorm.io/gorm"
)

type StageRepos struct {
	db *gorm.DB
}

func NewStagesRepos(db *gorm.DB) *StageRepos {
	return &StageRepos{db: db}
}

// CRUD

func (r *StageRepos) GetByIDAndGameID(
	ctx context.Context,
	id uint,
	gameID uint,
) (*models.Stages, error) {
	var stage models.Stages
	err := r.db.WithContext(ctx).
		Where(
			"id = ? AND game_id = ?",
			id,
			gameID,
		).
		First(&stage).Error

	if err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *StageRepos) GetByGameID(
	ctx context.Context,
	gameID uint,
) ([]models.Stages, error) {
	var stages []models.Stages
	err := r.db.WithContext(ctx).
		Where(
			" game_id = ?",
			gameID,
		).
		First(stages).Error

	if err != nil {
		return nil, err
	}
	return stages, nil
}

func (r *StageRepos) GetBySortID(
	ctx context.Context,
	sortID int,
	gameID uint,
) (*models.Stages, error) {
	var stage models.Stages

	err := r.db.WithContext(ctx).
		Where("game_id = ? AND sort_id = ?", gameID, sortID).
		First(&stage).
		Error

	if err != nil {
		return nil, err
	}

	return &stage, nil
}

func (r *StageRepos) GetAllStagesByGameID(
	ctx context.Context,
	gameID uint,
) ([]models.Stages, error) {
	var stages []models.Stages

	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Find(&stages).
		Error

	if err != nil {
		return nil, err
	}

	return stages, nil
}

func (r *StageRepos) Create(
	ctx context.Context,
	stage *models.Stages,
) error {
	return r.db.WithContext(ctx).Create(stage).Error
}

func (r *StageRepos) DeleteByID(
	ctx context.Context,
	id uint,
) error {
	result := r.db.WithContext(ctx).
		Delete(&models.Stages{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrStageNotFound
	}

	return nil
}
func (r *StageRepos) DeleteByGameID(
	ctx context.Context,
	gameID uint,
) error {
	return r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Delete(&models.Stages{}).
		Error
}

func (r *StageRepos) Update(
	ctx context.Context,
	stage *models.Stages,
) error {
	result := r.db.WithContext(ctx).
		Model(&models.Stages{}).
		Where("id = ? and game_id = ?", stage.ID, stage.GameID).
		Updates(map[string]interface{}{
			"sort_id":         stage.SortID,
			"name":            stage.Name,
			"description":     stage.Description,
			"location_id":     stage.LocationId,
			"determine_range": stage.DetermineRange,
			"limit_seconds":   stage.LimitSeconds,
			"quest":           stage.Quest,
			"is_end":          stage.IsEnd,
		})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrStageNotFound
	}

	return nil
}
