package models

import (
	"time"

	"gorm.io/datatypes"
)

type Game_play struct {
	ID                   uint            `gorm:"column:id;primaryKey"`
	UserID               uint            `gorm:"column:user_id"`
	GameID               uint            `gorm:"column:game_id"`
	CurrentStageStatusID *uint           `gorm:"column:current_stage_status_id"`
	CurrentQuestID       *uint           `gorm:"column:current_quest_id"`
	CurrentStageID       *uint           `gorm:"column:current_stage_id"`
	Variable             *datatypes.JSON `gorm:"column:variable"`
	Message              *string         `gorm:"column:message"`
	Rate                 *int            `gorm:"column:rate"`
	FinishStatus         *string         `gorm:"column:finish_status"`
	FinishAt             *time.Time      `gorm:"column:finish_at"`
	CreatedAt            time.Time       `gorm:"column:created_at"`
}

func (Game_play) TableName() string {
	return "game_plays"
}
