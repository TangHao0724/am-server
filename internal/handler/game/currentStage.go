package game

import (
	"amserver/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CurrentStage(c *gin.Context, db *gorm.DB) {
	// 獲取param gameID
	gameIDValue := c.Param("gameplayid")
	parsedGameID, parseErr := strconv.ParseUint(gameIDValue, 10, strconv.IntSize)
	if parseErr != nil || parsedGameID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_gameplay_id",
			"message": "遊戲 ID 格式錯誤",
		})
		return
	}

	gameplayID := uint(parsedGameID)

	// 獲取userID
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
	// 基本查詢
	ctx := c.Request.Context()
	gameRepos := repository.NewGameRepos(db)
	gamePlayRepos := repository.NewGamePlayRepos(db)
	stageRepos := repository.NewStagesRepos(db)
	locationRepos := repository.NewLocationRepos(db)

	// 1. 查詢遊戲
	gamePlayData, err := gamePlayRepos.GetByUserIDAndID(ctx, userID, gameplayID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "query_gamePlay_error",
			"message": "錯誤的 gamePlay 查詢",
		})
		return
	}
	gameData, err := gameRepos.GetByID(ctx, gamePlayData.GameID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "query_game_error",
			"message": "錯誤的 game 查詢",
		})
		return
	}
	// 獲取目前stage
	currentStage := gamePlayData.CurrentStageID
	// 空值判斷為自由探索
	if currentStage == nil {
		// 顯示關卡列表為額外API處哩，由setting開關
		c.JSON(http.StatusOK, gin.H{
			"status":             "freeSearch",
			"gameplay_id":        gamePlayData.ID,
			"game_id":            gameData.ID,
			"game_area_geojson":  gameData.AreaGeojson,
			"game_name":          gameData.Name,
			"game_limit_seconds": gameData.LimitSeconds,
			"game_description":   gameData.Description,
		})
		return
	}
	// 獲取stage資訊
	stageData, err := stageRepos.GetByIDAndGameID(ctx, *currentStage, gameData.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "query_stage_error",
			"message": "錯誤的 stage 查詢",
		})
		return
	}

	location, err := locationRepos.GetByID(
		ctx,
		stageData.LocationId,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "query_location_error",
			"message": "無法讀取關卡地點",
		})
		return
	}

	if location.Lon == nil || location.Lat == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "invalid_location",
			"message": "關卡地點沒有有效座標",
		})
		return
	}

	// 帶有關卡的探索，
	c.JSON(http.StatusOK, gin.H{
		"status":      "graph",
		"gameplay_id": gamePlayData.ID,
		"game_id":     gameData.ID,

		"game_area_geojson":  gameData.AreaGeojson,
		"game_name":          gameData.Name,
		"game_limit_seconds": gameData.LimitSeconds,
		"game_description":   gameData.Description,

		"stage_id":            stageData.ID,
		"stage_name":          stageData.Name,
		"stage_description":   stageData.Description,
		"stage_limit_seconds": stageData.LimitSeconds,

		"location": gin.H{
			"lon": location.Lon,
			"lat": location.Lat,
		},
	})
}
