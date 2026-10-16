package repository

import (
	"amserver/internal/models"
	"context"

	"gorm.io/gorm"
)

type LocationRepos struct {
	db *gorm.DB
}

func NewLocationRepos(db *gorm.DB) *LocationRepos {
	return &LocationRepos{db: db}
}

// CRUD

func (r *LocationRepos) GetByID(
	ctx context.Context,
	id uint,
) (*models.Locations, error) {
	var locate models.Locations
	err := r.db.WithContext(ctx).
		First(&locate, id).
		Error
	if err != nil {
		return nil, err
	}
	return &locate, nil
}
func (r *LocationRepos) GetByStageID(
	ctx context.Context,
	stageID uint,
) (*models.Locations, error) {
	var locate models.Locations

	err := r.db.WithContext(ctx).
		Table("locations").
		Joins("JOIN stages ON stages.location_id = locations.id").
		Where("stages.id = ?", stageID).
		First(&locate).
		Error

	if err != nil {
		return nil, err
	}

	return &locate, nil
}

func (r *LocationRepos) Create(
	ctx context.Context,
	locate *models.Locations,
) error {
	return r.db.WithContext(ctx).Create(locate).Error

}

func (r *LocationRepos) DeleteByID(
	ctx context.Context,
	id uint,
) error {
	result := r.db.WithContext(ctx).
		Delete(&models.Locations{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrLocationNotFound
	}

	return nil
}

func (r *LocationRepos) Update(
	ctx context.Context,
	locate *models.Locations,
) error {
	result := r.db.WithContext(ctx).
		Model(&models.Locations{}).
		Where("id = ?", locate.ID).
		Updates(map[string]interface{}{
			"name":        locate.Name,
			"description": locate.Description,
			"lon":         locate.Lon,
			"lat":         locate.Lat,
			"public_at":   locate.PublicAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrLocationNotFound
	}

	return nil
}
func (r *StageRepos) GetLocationIDsByGameID(
	ctx context.Context,
	gameID uint,
) ([]uint, error) {
	var ids []uint

	err := r.db.WithContext(ctx).
		Model(&models.Stages{}).
		Where("game_id = ?", gameID).
		Distinct("location_id").
		Pluck("location_id", &ids).
		Error

	return ids, err
}
func (r *StageRepos) CountByLocationID(
	ctx context.Context,
	locationID uint,
) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Stages{}).
		Where("location_id = ?", locationID).
		Count(&count).
		Error

	return count, err
}
