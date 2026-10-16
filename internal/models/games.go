package models

import (
	"time"

	"gorm.io/datatypes"
)

type Games struct {
	ID           uint            `gorm:"column:id;primaryKey"`
	CreatorID    uint            `gorm:"column:creator_id"`
	GameCode     string          `gorm:"column:game_code"`
	StartStageID *uint           `gorm:"column:start_stage_id"`
	Name         *string         `gorm:"column:name"`
	AreaGeojson  *datatypes.JSON `gorm:"column:area_geojson"`
	Setting      *datatypes.JSON `gorm:"column:setting"`
	LimitSeconds *int64          `gorm:"column:limit_seconds"`
	IsPublic     *bool           `gorm:"column:is_public"`
	Description  *string         `gorm:"column:description"`
	PublishAt    *time.Time      `gorm:"column:publish_at"`
	CreatedAt    time.Time       `gorm:"column:created_at"`
}

func (Games) TableName() string {
	return "games"
}
