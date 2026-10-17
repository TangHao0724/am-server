package game

import (
	"amserver/internal/models"
	"amserver/internal/repository"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func JoinGame(c *gin.Context, db *gorm.DB) {
	gameCode := c.Param("gamecode")
	if gameCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_game_code",
			"message": "缺少遊戲代碼",
		})
		return
	}

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
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "invalid_user_id",
			"message": "使用者身分驗證失敗",
		})
		return
	}

	ctx := c.Request.Context()
	var response gin.H

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		gameRepos := repository.NewGameRepos(tx)
		gamePlayRepos := repository.NewGamePlayRepos(tx)
		stageRepos := repository.NewStagesRepos(tx)
		locationRepos := repository.NewLocationRepos(tx)
		visitRepos := repository.NewGamePlayStageRepos(tx)

		// 1. 查詢遊戲
		gameData, err := gameRepos.GetByGameCode(ctx, gameCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newGameError(
					http.StatusNotFound,
					"game_not_found",
					"指定的遊戲不存在",
					err,
				)
			}
			return err
		}

		// 2. 有起始關卡時，先確認 Stage 與 Location
		var firstStage *models.Stages
		var location *models.Locations

		if gameData.StartStageID != nil {
			firstStage, err = stageRepos.GetByIDAndGameID(
				ctx,
				gameData.ID,
				*gameData.StartStageID,
			)
			if err != nil {
				return newGameError(
					http.StatusInternalServerError,
					"stage_lookup_failed",
					"無法讀取遊戲起始關卡",
					err,
				)
			}

			location, err = locationRepos.GetByID(
				ctx,
				firstStage.LocationId,
			)
			if err != nil {
				return newGameError(
					http.StatusInternalServerError,
					"location_lookup_failed",
					"無法讀取關卡地點",
					err,
				)
			}

			if location.Lon == nil || location.Lat == nil {
				return newGameError(
					http.StatusInternalServerError,
					"invalid_location",
					"關卡地點沒有有效座標",
					nil,
				)
			}
		}

		// 3. 建立 GamePlay
		playing := "playing"
		gamePlay := models.Game_play{
			UserID:         userID,
			GameID:         gameData.ID,
			CurrentStageID: gameData.StartStageID,
			FinishStatus:   &playing,
		}

		if err := gamePlayRepos.Create(ctx, &gamePlay); err != nil {
			return newGameError(
				http.StatusInternalServerError,
				"create_gameplay_failed",
				"建立遊玩紀錄失敗",
				err,
			)
		}

		// 4. 沒有起始關卡：回傳自由探索資訊
		if firstStage == nil {
			response = gin.H{
				"status":             "freeSearch",
				"gameplay_id":        gamePlay.ID,
				"game_id":            gameData.ID,
				"game_area_geojson":  gameData.AreaGeojson,
				"game_name":          gameData.Name,
				"game_limit_seconds": gameData.LimitSeconds,
				"game_description":   gameData.Description,
			}
			return nil
		}

		// 5. 建立首次 Stage Visit
		visit := models.Game_play_stage{
			GamePlayID: gamePlay.ID,
			StageID:    firstStage.ID,
			StartAt:    time.Now(),
		}

		if err := visitRepos.Create(ctx, &visit); err != nil {
			return newGameError(
				http.StatusInternalServerError,
				"create_stage_visit_failed",
				"建立關卡紀錄失敗",
				err,
			)
		}

		// 6. 設定目前 Stage Visit
		if err := tx.WithContext(ctx).
			Model(&gamePlay).
			Update(
				"current_stage_status_id",
				visit.ID,
			).Error; err != nil {
			return newGameError(
				http.StatusInternalServerError,
				"update_gameplay_failed",
				"更新遊戲進度失敗",
				err,
			)
		}

		// 7. 準備第一關資訊
		response = gin.H{
			"status":      "graph",
			"gameplay_id": gamePlay.ID,
			"game_id":     gameData.ID,

			"game_area_geojson":  gameData.AreaGeojson,
			"game_name":          gameData.Name,
			"game_limit_seconds": gameData.LimitSeconds,
			"game_description":   gameData.Description,

			"stage_id":            firstStage.ID,
			"stage_name":          firstStage.Name,
			"stage_description":   firstStage.Description,
			"stage_limit_seconds": firstStage.LimitSeconds,

			"location": gin.H{
				"lon": *location.Lon,
				"lat": *location.Lat,
			},
		}

		return nil
	})

	// 8. Transaction 結束後統一處理錯誤
	if err != nil {
		var ge *gameError
		if errors.As(err, &ge) {
			c.JSON(ge.Status, gin.H{
				"status":  false,
				"error":   ge.Code,
				"message": ge.Message,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "join_game_failed",
			"message": "加入遊戲失敗",
		})
		return
	}

	// 9. Commit 成功才回傳
	c.JSON(http.StatusOK, response)
}
