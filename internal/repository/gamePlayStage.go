package repository

import (
	"amserver/internal/models"
	"context"

	"gorm.io/gorm"
)

type GamePlayStageRepos struct {
	db *gorm.DB
}

func NewGamePlayStageRepos(db *gorm.DB) *GamePlayStageRepos {
	return &GamePlayStageRepos{db: db}
}

func (r *GamePlayStageRepos) GetByGamePlayIDAndCurrentID(
	ctx context.Context,
	currentStageID uint,
	gamePlayID uint,
) (*models.Game_play_stage, error) {
	var gamePlayS models.Game_play_stage
	err := r.db.WithContext(ctx).
		Where("gameplay_id = ?", gamePlayID).
		First(&gamePlayS, currentStageID).
		Error

	if err != nil {
		return nil, err
	}

	return &gamePlayS, nil
}

func (r *GamePlayStageRepos) Create(
	ctx context.Context,
	gamePlay *models.Game_play_stage,
) error {
	return r.db.WithContext(ctx).Create(gamePlay).Error
}
