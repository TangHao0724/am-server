package repository

import (
	"time"

	"amserver/internal/models"

	"gorm.io/gorm"
)

type GamePlayStagesRepository struct {
	db *gorm.DB
}

func NewGamePlayStagesRepository(
	db *gorm.DB,
) *GamePlayStagesRepository {
	return &GamePlayStagesRepository{db: db}
}

// [Repository] 查詢玩家在指定遊戲中的最新進度。
// 此實作假設每位玩家在每個遊戲中只有一筆進度紀錄。
func (r *GamePlayStagesRepository) GetByGameAndUser(
	gameID int32,
	userID uint,
) (*models.Game_play_stages_test, error) {
	var progress models.Game_play_stages_test

	err := r.db.
		Where("game_id = ? AND user_id = ?", gameID, userID).
		Order("id DESC").
		First(&progress).Error

	if err != nil {
		return nil, err
	}

	return &progress, nil
}

// [Repository] 建立初始進度。
func (r *GamePlayStagesRepository) Create(
	progress *models.Game_play_stages_test,
) error {
	return r.db.Create(progress).Error
}

// [Repository] 更新玩家進度。
func (r *GamePlayStagesRepository) UpdateProgress(
	id uint,
	nowStageID int32,
	nextStageID *int32,
	lon string,
	lat string,
	startAt time.Time,
	checkedAt time.Time,
) error {
	return r.db.
		Model(&models.Game_play_stages_test{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"now_stage_id":  nowStageID,
			"next_stage_id": nextStageID,
			"lon":           lon,
			"lat":           lat,
			"start_at":      startAt,
			"checked_at":    checkedAt,
		}).Error
}
