package models

import (
	"time"

	"gorm.io/datatypes"
)

type Stages_test struct {
	ID              uint            `gorm:"column:id;primaryKey"`
	GameID          uint            `gorm:"column:game_id"`
	Name            *string         `gorm:"column:name"`
	Lon             *float64        `gorm:"column:lon"`
	Lat             *float64        `gorm:"column:lat"`
	QrcodeUid       *string         `gorm:"column:qrcode_uid"`
	DetermineRrange *string         `gorm:"column:determine_range"`
	LimitSeconds    *int32          `gorm:"column:limit_seconds"`
	Quest           *datatypes.JSON `gorm:"column:quest"`
	IsEnd           bool            `gorm:"column:is_end"`
	CreatedAt       *time.Time      `gorm:"column:created_at"`
}

func (Stages_test) TableName() string {
	return "stages_test"
}
