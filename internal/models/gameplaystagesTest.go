package models

import "time"

type Game_play_stages_test struct {
	ID          uint       `gorm:"column:id;primaryKey"`
	GameID      int32      `gorm:"column:game_id"`
	UserID      *uint      `gorm:"column:user_id"`
	NowStageID  int32      `gorm:"column:now_stage_id"`
	NextStageID *int32     `gorm:"column:next_stage_id"`
	Lon         *string    `gorm:"column:lon"`
	Lat         *string    `gorm:"column:lat"`
	StartAt     *time.Time `gorm:"column:start_at"`
	CheckedAt   *time.Time `gorm:"column:checked_at"`
}

func (Game_play_stages_test) TableName() string {
	return "game_play_stages_test"
}
