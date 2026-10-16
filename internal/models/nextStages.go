package models

type NextStages struct {
	ID            uint  `gorm:"column:id;primaryKey"`
	GameID        uint  `gorm:"column:game_id"`
	NowStageID    uint  `gorm:"column:now_stage_id"`
	OptionStageId *uint `gorm:"column:option_stage_id"`
}

func (NextStages) TableName() string {
	return "next_stages"
}
