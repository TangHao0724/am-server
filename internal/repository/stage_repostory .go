package repository

import (
	"amserver/internal/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type StageRepository struct {
	db *gorm.DB
}

func NewStageRepository(db *gorm.DB) *StageRepository {
	return &StageRepository{db: db}
}
func (r *StageRepository) GetByID(
	ctx context.Context,
	stageID uint,
	gameID uint,
) (*models.Stages_test, error) {
	var stage models.Stages_test

	err := r.db.WithContext(ctx).
		Where("id = ? AND game_id = ?", stageID, gameID).
		First(&stage).
		Error
	if err != nil {
		return nil, err
	}

	return &stage, nil
}
func (r *StageRepository) Create(
	ctx context.Context,
	stage *models.Stages_test,
) error {
	err := r.db.WithContext(ctx).Create(stage).Error
	if err != nil {
		return err
	}

	return nil
}
func (r *StageRepository) GetByQRCodeUID(
	gameID uint,
	qrcodeUID string,
) (*models.Stages_test, error) {
	var stage models.Stages_test

	err := r.db.
		Where("game_id = ? AND qrcode_uid = ?", gameID, qrcodeUID).
		First(&stage).Error

	if err != nil {
		return nil, err
	}

	return &stage, nil
}

// [Repository] 取得遊戲所有關卡，暫時依 ID 排序。
func (r *StageRepository) ListByGameID(
	gameID uint,
) ([]models.Stages_test, error) {
	var stages []models.Stages_test

	err := r.db.
		Where("game_id = ?", gameID).
		Order("id ASC").
		Find(&stages).Error

	if err != nil {
		return nil, err
	}

	return stages, nil
}
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
