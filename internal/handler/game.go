package handler

import (
	"amserver/internal/models"
	"math/rand/v2"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreateStageRequest struct {
	Name            string         `json:"name"`
	Lon             float64        `json:"lon"`
	Lat             float64        `json:"lat"`
	QrcodeUid       string         `json:"qrcode_uid"`
	DetermineRrange *string        `json:"determine_range"`
	LimitSeconds    *int32         `json:"limit_seconds"`
	Quest           datatypes.JSON `json:"quest"`
	IsEnd           bool           `json:"is_end"`
}

type CreateGameRequest struct {
	StartStageIndex int                  `json:"start_stage_index"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Stages          []CreateStageRequest `json:"stages"`
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString 產生指定長度的隨機字串
func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

// 建立遊戲
func CreateGameHandler(c *gin.Context, db *gorm.DB) {
	userID, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	id := userID.(uint)

	ctx := c.Request.Context()

	var req CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 驗證起始關卡索引
	if len(req.Stages) == 0 ||
		req.StartStageIndex < 0 ||
		req.StartStageIndex >= len(req.Stages) {
		c.JSON(400, gin.H{
			"error": "invalid start_stage_index",
		})
		return
	}

	var game models.Games_test

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		game = models.Games_test{
			CreatorID:    id,
			GameCode:     generateRandomString(8),
			Name:         &req.Name,
			Description:  &req.Description,
			StartStageID: nil,
		}

		// 1. 建立 Game，取得 game.ID
		if err := tx.Create(&game).Error; err != nil {
			return err
		}

		// 2. 建立所有 Stage，並記錄其資料庫 ID
		stages := make([]models.Stages_test, len(req.Stages))

		for i, s := range req.Stages {
			stage := models.Stages_test{
				GameID:          game.ID,
				Name:            &s.Name,
				Lon:             &s.Lon,
				Lat:             &s.Lat,
				QrcodeUid:       &s.QrcodeUid,
				DetermineRrange: s.DetermineRrange,
				LimitSeconds:    s.LimitSeconds,
				Quest:           &s.Quest,
				IsEnd:           s.IsEnd,
			}

			if err := tx.Create(&stage).Error; err != nil {
				return err
			}

			stages[i] = stage
		}

		// 3. 取得起始 Stage 的資料庫 ID
		startStageID := stages[req.StartStageIndex].ID
		game.StartStageID = &startStageID

		// 4. 更新 Game 的 start_stage_id
		if err := tx.Model(&game).
			Update("start_stage_id", startStageID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"game_id":        game.ID,
		"start_stage_id": game.StartStageID,
	})
}
