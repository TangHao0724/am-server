package models

import (
	"time"

	"gorm.io/datatypes"
)

type Stages struct {
	ID             uint            `gorm:"column:id;primaryKey"`
	GameID         uint            `gorm:"column:game_id"`
	SortID         int             `gorm:"column:sort_id"`
	LocationId     uint            `gorm:"column:location_id"`
	QrcodeUid      *string         `gorm:"column:qrcode_uid"`
	Name           *string         `gorm:"column:name"`
	Description    *string         `gorm:"column:description"`
	DetermineRange *float64        `gorm:"column:determine_range"`
	LimitSeconds   *int64          `gorm:"column:limit_seconds"`
	Quest          *datatypes.JSON `gorm:"column:quest"`
	IsEnd          bool            `gorm:"column:is_end"`
	CreatedAt      time.Time       `gorm:"column:created_at"`
}

func (Stages) TableName() string {
	return "stages"
}
