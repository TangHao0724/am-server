package models

import (
	"time"

	"gorm.io/datatypes"
)

type Game_play_stage struct {
	ID              uint            `gorm:"column:id;primaryKey"`
	GamePlayID      uint            `gorm:"column:gameplay_id"`
	StageID         uint            `gorm:"column:stage_id"`
	Variables       *datatypes.JSON `gorm:"column:variables"`
	QuestAnswerData *datatypes.JSON `gorm:"column:quest_answer_data"`
	CheckedAt       *time.Time      `gorm:"column:checked_at"`
	StartAt         time.Time       `gorm:"column:start_at"`
}

func (Game_play_stage) TableName() string {
	return "game_play_stages"
}
