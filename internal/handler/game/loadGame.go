package game

import (
	"amserver/internal/repository"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type location struct {
	ID          uint       `json:"id"`
	CreatorID   uint       `json:"creator_id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Lon         *float64   `json:"lon"`
	Lat         *float64   `json:"lat"`
	PublicAt    *time.Time `json:"public_at"`
}

type stageResult struct {
	ID             uint            `json:"id"`
	GameID         uint            `json:"game_id"`
	SortID         int             `json:"sort_id"`
	Location       location        `json:"location"`
	QrcodeUid      *string         `json:"qrcode_uid"`
	Name           *string         `json:"name"`
	Description    *string         `json:"description"`
	DetermineRange *float64        `json:"determine_range"`
	LimitSeconds   *int64          `json:"limit_seconds"`
	Quest          *datatypes.JSON `json:"quest"`
	IsEnd          bool            `json:"is_end"`
}
type gameResult struct {
	ID           uint            `json:"id"`
	CreatorID    uint            `json:"creator_id"`
	GameCode     string          `json:"game_code"`
	Stages       []stageResult   `json:"stages"`
	StartStageID *uint           `json:"start_stage_id"`
	Name         *string         `json:"name"`
	AreaGeojson  *datatypes.JSON `json:"area_geojson"`
	Setting      *datatypes.JSON `json:"setting"`
	LimitSeconds *int64          `json:"limit_seconds"`
	IsPublic     *bool           `json:"is_public"`
	Description  *string         `json:"description"`
	PublishAt    *time.Time      `json:"publish_at"`
	CreatedAt    time.Time       `json:"created_at"`
}

// 從編輯器中載入編輯存檔
func LoadGameHandler(c *gin.Context, db *gorm.DB) {
	gameIDValue := c.Param("gameid")
	parsedGameID, parseErr := strconv.ParseUint(gameIDValue, 10, strconv.IntSize)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_game_id",
			"message": "遊戲 ID 格式錯誤",
		})
		return
	}
	gameID := uint(parsedGameID)

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "尚未登入或登入憑證失效",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "invalid_user_id",
			"message": "使用者身分驗證失敗",
		})
		return
	}

	ctx := c.Request.Context()
	var result gameResult
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		gameRepos := repository.NewGameRepos(tx)
		stageRepos := repository.NewStagesRepos(tx)
		locationsRepos := repository.NewLocationRepos(tx)

		// 1. game
		game, err := gameRepos.GetByIDAndUserID(ctx, gameID, userID)
		if err != nil {
			return newCreateGameError(
				http.StatusInternalServerError,
				"select_game_failed",
				"查詢遊戲失敗",
				err,
			)
		}

		// 2. stage
		stages, err := stageRepos.GetAllStagesByGameID(ctx, game.ID)
		if err != nil {
			return newCreateGameError(
				http.StatusInternalServerError,
				"select_stages_failed",
				"查詢關卡失敗",
				err,
			)
		}
		// 3. location
		stageResults := make([]stageResult, 0, len(stages))
		for _, i := range stages {
			locationData, err := locationsRepos.GetByStageID(ctx, i.ID)
			if err != nil {
				return newCreateGameError(
					http.StatusInternalServerError,
					"select_location_failed",
					"查詢地點失敗",
					err,
				)
			}

			location := location{
				ID:          locationData.ID,
				CreatorID:   locationData.CreatorID,
				Name:        locationData.Name,
				Description: locationData.Description,
				Lon:         locationData.Lon,
				Lat:         locationData.Lat,
				PublicAt:    locationData.PublicAt,
			}

			stageResult := stageResult{
				ID:             i.ID,
				GameID:         i.GameID,
				SortID:         i.SortID,
				Location:       location,
				QrcodeUid:      i.QrcodeUid,
				Name:           i.Name,
				Description:    i.Description,
				DetermineRange: i.DetermineRange,
				LimitSeconds:   i.LimitSeconds,
				Quest:          i.Quest,
				IsEnd:          i.IsEnd,
			}
			stageResults = append(stageResults, stageResult)

		}
		result = gameResult{
			ID:           game.ID,
			CreatorID:    game.CreatorID,
			GameCode:     game.GameCode,
			Stages:       stageResults,
			StartStageID: game.StartStageID,
			Name:         game.Name,
			AreaGeojson:  game.AreaGeojson,
			Setting:      game.Setting,
			LimitSeconds: game.LimitSeconds,
			IsPublic:     game.IsPublic,
			Description:  game.Description,
			PublishAt:    game.PublishAt,
			CreatedAt:    game.CreatedAt,
		}
		return nil
	})
	if err != nil {
		var apiErr *createGameError

		if errors.As(err, &apiErr) {
			if apiErr.Cause != nil {
				log.Printf(
					"LoadGameHandler: code=%s cause=%v",
					apiErr.Code,
					apiErr.Cause,
				)
			}

			c.JSON(apiErr.Status, gin.H{
				"status":  false,
				"error":   apiErr.Code,
				"message": apiErr.Message,
			})
			return
		}

		// 未預期的錯誤
		log.Printf("LoadGameHandler: unexpected error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "internal_error",
			"message": "伺服器發生錯誤，請稍後再試",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "遊戲讀取成功",
		"game":    result,
	})
}
