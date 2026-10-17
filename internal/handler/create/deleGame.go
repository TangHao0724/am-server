package create

import (
	"amserver/internal/models"
	"amserver/internal/repository"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeleteRequest struct {
	DeleteLocationIDs []uint `json:"delete_location_ids"`
}

func DeleteGameHandler(c *gin.Context, db *gorm.DB) {
	// gameID param
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
	// gameID jwt
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
	// body
	var req DeleteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_request",
			"message": "請輸入正確格式的請求",
		})
		return
	}

	ctx := c.Request.Context()
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		gameRepos := repository.NewGameRepos(tx)
		stageRepos := repository.NewStagesRepos(tx)
		locationsRepos := repository.NewLocationRepos(tx)
		nsRepos := repository.NewNextStageRepos(tx)

		// 1. 先驗證 Game 擁有權
		if _, err := gameRepos.GetByIDAndUserID(
			ctx, gameID, userID,
		); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newCreateGameError(
					http.StatusNotFound,
					"game_not_found",
					"遊戲不存在或沒有權限刪除",
					err,
				)
			}

			return newCreateGameError(
				http.StatusInternalServerError,
				"select_game_failed",
				"查詢遊戲失敗",
				err,
			)
		}

		// 2. 取得此 Game 的所有 Location ID
		gameLocationIDs, err := stageRepos.GetLocationIDsByGameID(
			ctx, gameID,
		)
		if err != nil {
			return newCreateGameError(
				http.StatusInternalServerError,
				"select_locations_failed",
				"查詢遊戲地點失敗",
				err,
			)
		}

		allowedLocations := make(map[uint]bool)
		for _, id := range gameLocationIDs {
			allowedLocations[id] = true
		}

		// 3. 先驗證所有要求刪除的 Location
		selectedLocations := make(map[uint]bool)

		for _, locationID := range req.DeleteLocationIDs {
			if !allowedLocations[locationID] {
				return newCreateGameError(
					http.StatusBadRequest,
					"invalid_location_id",
					"指定地點不屬於此遊戲",
					nil,
				)
			}

			if selectedLocations[locationID] {
				return newCreateGameError(
					http.StatusBadRequest,
					"duplicate_location_id",
					"地點 ID 重複",
					nil,
				)
			}
			selectedLocations[locationID] = true

			location, err := locationsRepos.GetByID(
				ctx, locationID,
			)
			if err != nil {
				return newCreateGameError(
					http.StatusBadRequest,
					"location_not_found",
					"指定地點不存在",
					err,
				)
			}

			if location.CreatorID != userID {
				return newCreateGameError(
					http.StatusForbidden,
					"location_forbidden",
					"沒有權限刪除指定地點",
					nil,
				)
			}
		}

		// 4. 解除 Game -> StartStage 的 FK
		if err := tx.WithContext(ctx).
			Model(&models.Games{}).
			Where("id = ?", gameID).
			Update("start_stage_id", nil).
			Error; err != nil {

			return newCreateGameError(
				http.StatusInternalServerError,
				"clear_start_stage_failed",
				"清除起始關卡失敗",
				err,
			)
		}

		// 5. 刪除 NextStages
		if err := nsRepos.DeleteByGameID(ctx, gameID); err != nil {
			return newCreateGameError(
				http.StatusInternalServerError,
				"remove_nextstage_failed",
				"移除關卡關聯失敗",
				err,
			)
		}

		// 此處仍須先處理其他引用 Stage 的資料表，
		// 例如玩家進度；依實際 FK 設定決定。

		// 6. 刪除 Stages
		if err := stageRepos.DeleteByGameID(ctx, gameID); err != nil {
			return newCreateGameError(
				http.StatusInternalServerError,
				"remove_stage_failed",
				"移除關卡失敗",
				err,
			)
		}

		// 7. 刪除 Game
		if err := gameRepos.DeleteByID(ctx, gameID); err != nil {
			return newCreateGameError(
				http.StatusInternalServerError,
				"remove_game_failed",
				"移除遊戲失敗",
				err,
			)
		}

		// 8. 刪除使用者指定的 Locations
		for locationID := range selectedLocations {

			count, err := stageRepos.CountByLocationID(
				ctx, locationID,
			)
			if err != nil {
				return newCreateGameError(
					http.StatusInternalServerError,
					"check_location_usage_failed",
					"檢查地點使用情況失敗",
					err,
				)
			}

			if count > 0 {
				return newCreateGameError(
					http.StatusConflict,
					"location_in_use",
					"指定地點仍被其他關卡使用",
					nil,
				)
			}

			if err := locationsRepos.DeleteByID(
				ctx, locationID,
			); err != nil {
				return newCreateGameError(
					http.StatusInternalServerError,
					"remove_location_failed",
					"移除地點失敗",
					err,
				)
			}
		}

		return nil
	})
	if err != nil {
		var apiErr *createGameError

		if errors.As(err, &apiErr) {
			if apiErr.Cause != nil {
				log.Printf(
					"NewGameHandler: code=%s cause=%v",
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
		log.Printf("NewGameHandler: unexpected error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "internal_error",
			"message": "伺服器發生錯誤，請稍後再試",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "遊戲移除成功",
		"game_id": gameID,
	})
}
