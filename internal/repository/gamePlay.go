package repository

import (
	"amserver/internal/models"
	"context"

	"gorm.io/gorm"
)

type GamePlayRepos struct {
	db *gorm.DB
}

func NewGamePlayRepos(db *gorm.DB) *GamePlayRepos {
	return &GamePlayRepos{db: db}
}

func (r *GamePlayRepos) GetByUserIDAndGameID(
	ctx context.Context,
	userid uint,
	gameid int,
) (*models.Game_play, error) {
	var gamePlay models.Game_play
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND game_id = ?", userid, gameid).
		First(&gamePlay).Error

	if err != nil {
		return nil, err
	}
	return &gamePlay, nil
}
func (r *GamePlayRepos) GetByUserIDAndID(
	ctx context.Context,
	userid uint,
	id uint,
) (*models.Game_play, error) {
	var gamePlay models.Game_play
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userid).
		First(&gamePlay, id).Error

	if err != nil {
		return nil, err
	}
	return &gamePlay, nil
}

func (r *GamePlayRepos) Create(
	ctx context.Context,
	gamePlay *models.Game_play,
) error {
	return r.db.WithContext(ctx).Create(gamePlay).Error
}
