package models

import "time"

type Games_test struct {
	ID           uint       `gorm:"column:id;primaryKey"`
	CreatorID    uint       `gorm:"column:creator_id"`
	GameCode     string     `gorm:"column:game_code"`
	StartStageID *uint      `gorm:"column:start_stage_id"`
	Name         *string    `gorm:"column:name"`
	Description  *string    `gorm:"column:description"`
	CreatedAt    *time.Time `gorm:"column:created_at"`
}

func (Games_test) TableName() string {
	return "games_test"
}
