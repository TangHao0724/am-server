package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"amserver/internal/models"
	"amserver/internal/repository"
)

var (
	ErrStageNotFound   = errors.New("stage not found")
	ErrWrongStage      = errors.New("stage is not the expected next stage")
	ErrNoStages        = errors.New("game has no stages")
	ErrProgressMissing = errors.New("game progress not found")
)

type QRCodeService struct {
	stageRepo    *repository.StageRepository
	progressRepo *repository.GamePlayStagesRepository
}

func NewQRCodeService(
	stageRepo *repository.StageRepository,
	progressRepo *repository.GamePlayStagesRepository,
) *QRCodeService {
	return &QRCodeService{
		stageRepo:    stageRepo,
		progressRepo: progressRepo,
	}
}

type ScanStageInput struct {
	GameID    int32
	UserID    uint
	QRCodeUID string
	Lon       float64
	Lat       float64
}

type ScanStageResult struct {
	Status      string `json:"status"`
	NextStageID *int32 `json:"next_stage_id"`
	IsEnd       bool   `json:"is_end"`
}

// ScanStage 處理掃描 QR Code 後的關卡驗證與進度更新。
func (s *QRCodeService) ScanStage(
	input ScanStageInput,
) (*ScanStageResult, error) {
	if input.GameID <= 0 || input.UserID <= 0 {
		return nil, ErrProgressMissing
	}

	// [Repository] 1. 從 QR Code UUID 取得關卡。
	stage, err := s.stageRepo.GetByQRCodeUID(
		uint(input.GameID),
		input.QRCodeUID,
	)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, ErrStageNotFound
		}
		return nil, err
	}

	// [Repository] 2. 取得該遊戲所有關卡。
	stages, err := s.stageRepo.ListByGameID(uint(input.GameID))
	if err != nil {
		return nil, err
	}
	if len(stages) == 0 {
		return nil, ErrNoStages
	}

	// 3. 找出目前關卡的下一關。
	var nextStageID *int32

	for i, item := range stages {
		if item.ID != stage.ID {
			continue
		}

		if i+1 < len(stages) {
			id := int32(stages[i+1].ID)
			nextStageID = &id
		}
		break
	}

	// [Repository] 4. 取得玩家目前進度。
	progress, err := s.progressRepo.GetByGameAndUser(
		input.GameID,
		input.UserID,
	)

	now := time.Now()

	if err != nil {
		if !repository.IsNotFound(err) {
			return nil, err
		}

		// 首次遊玩：必須從第一關開始。
		if stage.ID != stages[0].ID {
			fmt.Println("stage.ID != stages[0].ID 這裡判斷又問題")
			return nil, ErrWrongStage
		}

		// [Repository] 5. 建立初始進度。
		// StartAt 記錄首次進入此關的時間。
		// CheckedAt 記錄本次掃碼成功的時間。
		progress = &models.Game_play_stages_test{
			GameID:      input.GameID,
			UserID:      &input.UserID,
			NowStageID:  int32(stage.ID),
			NextStageID: nextStageID,
			Lon:         stringPtr(strconv.FormatFloat(input.Lon, 'f', -1, 64)),
			Lat:         stringPtr(strconv.FormatFloat(input.Lat, 'f', -1, 64)),
			StartAt:     &now,
			CheckedAt:   &now,
		}

		if err := s.progressRepo.Create(progress); err != nil {
			return nil, err
		}

		return &ScanStageResult{
			Status:      "Saved",
			NextStageID: nextStageID,
			IsEnd:       stage.IsEnd || nextStageID == nil,
		}, nil
	}

	// 6. 驗證玩家掃描的是預期關卡。
	//
	// 這裡假設 NowStageID 表示目前應完成的關卡。
	if progress.NextStageID == nil {
		// 沒有下一關，代表遊戲已經結束
		return nil, ErrWrongStage
	}

	if *progress.NextStageID != int32(stage.ID) {
		// 玩家掃描的不是預期的下一關
		fmt.Println("progress.NowStageID != int32(stage.ID) 這裡判斷又問題")
		return nil, ErrWrongStage
	}
	// [Repository] 7. 更新玩家進度。
	//
	// 這個版本將 NowStageID 保留為剛完成的關卡，
	// NextStageID 記錄下一關。
	//
	// 若要在下一次掃描時自動前進，必須另外定義
	// NowStageID 是「目前關卡」還是「下一個待完成關卡」。
	startAt := now
	if progress.StartAt != nil {
		startAt = *progress.StartAt
	}

	lon := strconv.FormatFloat(input.Lon, 'f', -1, 64)
	lat := strconv.FormatFloat(input.Lat, 'f', -1, 64)

	if err := s.progressRepo.UpdateProgress(
		progress.ID,
		int32(stage.ID),
		nextStageID,
		lon,
		lat,
		startAt,
		now,
	); err != nil {
		return nil, err
	}

	return &ScanStageResult{
		Status:      "Saved",
		NextStageID: nextStageID,
		IsEnd:       stage.IsEnd || nextStageID == nil,
	}, nil
}

func stringPtr(value string) *string {
	return &value
}
