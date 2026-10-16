package repository

import (
	"amserver/internal/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type NextStageRepos struct {
	db *gorm.DB
}

func NewNextStageRepos(db *gorm.DB) *NextStageRepos {
	return &NextStageRepos{db: db}
}

func (r *NextStageRepos) GetByID(
	ctx context.Context,
	id uint,
) (*models.NextStages, error) {
	var ns models.NextStages

	err := r.db.WithContext(ctx).
		First(&ns, id).
		Error
	if err != nil {
		return nil, err
	}

	return &ns, nil
}

func (r *NextStageRepos) GetByGameID(
	ctx context.Context,
	gameID uint,
) ([]models.NextStages, error) {
	var nss []models.NextStages

	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Find(&nss).
		Error
	if err != nil {
		return nil, err
	}

	return nss, nil
}

func (r *NextStageRepos) Create(
	ctx context.Context,
	nextStage *models.NextStages,
) error {
	err := r.db.WithContext(ctx).Create(nextStage).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "nextStages_unique_0" {
		return ErrDuplicateNextStageSort
	}

	return err
}

func (r *NextStageRepos) DeleteByID(
	ctx context.Context,
	id uint,
) error {
	result := r.db.WithContext(ctx).
		Delete(&models.NextStages{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNextStageNotFound
	}

	return nil
}

func (r *NextStageRepos) DeleteByGameID(
	ctx context.Context,
	gameID uint,
) error {
	return r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Delete(&models.NextStages{}).
		Error
}

func (r *NextStageRepos) ReplaceByNowStageID(
	ctx context.Context,
	nowStageID uint,
	gameID uint,
	nextStages []models.NextStages,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1. 刪除原本的 NextStages
		if err := tx.
			Where(
				"now_stage_id = ? AND game_id = ?",
				nowStageID,
				gameID,
			).
			Delete(&models.NextStages{}).Error; err != nil {
			return err
		}

		// 2. 如果沒有新關聯，刪除後直接完成
		if len(nextStages) == 0 {
			return nil
		}

		// 3. 批次建立新的 NextStages
		if err := tx.Create(&nextStages).Error; err != nil {
			return err
		}

		return nil
	})
}
