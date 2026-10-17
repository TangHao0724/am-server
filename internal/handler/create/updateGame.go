package create

import (
	"amserver/internal/models"
	"amserver/internal/repository"
	"amserver/internal/service/questjson"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type UpdateGameRequest struct {
	GameID       uint                  `json:"game_id"`
	Name         string                `json:"name" binding:"required,min=1,max=50"`
	GameCode     string                `json:"game_code"`
	AreaGeojson  *datatypes.JSON       `json:"area_geojson"`
	Setting      *datatypes.JSON       `json:"setting"`
	LimitSeconds *int64                `json:"limit_seconds" binding:"omitempty,gte=0"`
	IsPublic     *bool                 `json:"is_public"`
	Description  *string               `json:"description"`
	PublishAt    *time.Time            `json:"publish_at"`
	Stages       *[]UpdateStageRequest `json:"stages"`
}

type UpdateStageRequest struct {
	StageID        uint                   `json:"stage_id"`
	Name           *string                `json:"name"`
	SortID         int                    `json:"sort_id"`
	IsStart        bool                   `json:"is_start"`
	Description    *string                `json:"description"`
	DetermineRange *float64               `json:"determine_range" binding:"omitempty,gte=0"`
	LimitSeconds   *int64                 `json:"limit_seconds" binding:"omitempty,gte=0"`
	Quest          *datatypes.JSON        `json:"quest"`
	IsEnd          bool                   `json:"is_end"`
	Locate         *UpdateLocationRequest `json:"locate"`
}

type UpdateLocationRequest struct {
	LocationID  *uint      `json:"location_id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Lon         *float64   `json:"lon" binding:"omitempty,gte=-180,lte=180"`
	Lat         *float64   `json:"lat" binding:"omitempty,gte=-90,lte=90"`
	PublicAt    *time.Time `json:"public_at"`
}

func UpdateGameHandler(c *gin.Context, db *gorm.DB) {
	// 獲取事先資料
	// param gameid
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

	// jwt userID
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
	// request ID
	var req UpdateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_request",
			"message": "請輸入正確格式的請求",
		})
		return
	}
	// 檢查輸入資訊
	// 3. 驗證遊戲名稱與關卡數量
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_game_name",
			"message": "遊戲名稱不可為空白",
		})
		return
	}

	if req.Stages == nil || len(*req.Stages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "missing_stages",
			"message": "至少需要建立一個關卡",
		})
		return
	}

	// 4. 在寫入資料庫前驗證 Stage 與地點資料
	startCount := 0

	parsedQuests := make([]questjson.Quests, len(*req.Stages))

	for index, stage := range *req.Stages {
		if stage.IsStart {
			startCount++
		}

		if stage.Locate == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"error":   "missing_location",
				"message": "關卡缺少地點資料",
				"index":   index,
			})
			return
		}

		locate := stage.Locate

		if locate.LocationID != nil {
			// 使用既有地點時，不允許同時傳入新地點資料
			if *locate.LocationID == 0 ||
				locate.Name != nil ||
				locate.Description != nil ||
				locate.Lon != nil ||
				locate.Lat != nil ||
				locate.PublicAt != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  false,
					"error":   "invalid_existing_location",
					"message": "使用既有地點時，只能提供地點 ID",
					"index":   index,
				})
				return
			}
		} else {
			// 新地點需要名稱、經度與緯度
			if locate.Name == nil ||
				strings.TrimSpace(*locate.Name) == "" ||
				locate.Lon == nil ||
				locate.Lat == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  false,
					"error":   "invalid_new_location",
					"message": "新地點需要名稱、經度與緯度",
					"index":   index,
				})
				return
			}

			if *locate.Lon < -180 || *locate.Lon > 180 ||
				*locate.Lat < -90 || *locate.Lat > 90 {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  false,
					"error":   "invalid_coordinates",
					"message": "經緯度超出允許範圍",
					"index":   index,
				})
				return
			}
		}

		// 先解析 Quest JSON，避免資料寫入後才發現格式錯誤
		if stage.Quest != nil {

			// 1. 解析 JSON
			if err := json.Unmarshal(
				[]byte(*stage.Quest),
				&parsedQuests[index],
			); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  false,
					"error":   "invalid_quest_json",
					"message": "關卡題目 JSON 格式錯誤",
					"index":   index,
				})
				return
			}

			// 2. 驗證 Quest 結構與規則
			if err := questjson.ValidateQuests(parsedQuests[index]); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  false,
					"error":   "invalid_quest_config",
					"message": err.Error(),
					"index":   index,
				})
				return
			}
		}
	}

	if startCount > 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_start_stage_count",
			"message": "只能設定一個起始關卡",
		})
		return
	}
	ctx := c.Request.Context()

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		gameRepos := repository.NewGameRepos(tx)
		stageRepos := repository.NewStagesRepos(tx)
		locationsRepos := repository.NewLocationRepos(tx)
		nsRepos := repository.NewNextStageRepos(tx)

		// 1. 驗證遊戲擁有權
		_, err := gameRepos.GetByIDAndUserID(ctx, gameID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newCreateGameError(
					http.StatusNotFound,
					"game_not_found",
					"遊戲不存在或沒有權限修改",
					err,
				)
			}

			return newCreateGameError(
				http.StatusInternalServerError,
				"get_game_failed",
				"查詢遊戲失敗",
				err,
			)
		}

		// 2. 更新 Game 基本資料
		updateGame := &models.Games{
			ID:           gameID,
			Name:         &req.Name,
			CreatorID:    userID,
			GameCode:     req.GameCode,
			AreaGeojson:  req.AreaGeojson,
			Setting:      req.Setting,
			LimitSeconds: req.LimitSeconds,
			IsPublic:     req.IsPublic,
			Description:  req.Description,
			PublishAt:    req.PublishAt,
		}

		if err := gameRepos.Update(ctx, updateGame); err != nil {
			if errors.Is(err, repository.ErrDuplicateGameCode) {
				return newCreateGameError(
					http.StatusConflict,
					"duplicate_game_code",
					"遊戲代碼已被使用",
					err,
				)
			}

			return newCreateGameError(
				http.StatusInternalServerError,
				"update_game_failed",
				"更新遊戲失敗",
				err,
			)
		}

		// sort_id -> stage_id
		updatedStages := make(map[int]uint)

		// 避免同一個 stage_id 被重複提交
		usedStageIDs := make(map[uint]bool)

		var startStageID *uint

		// 3. 第一階段：更新全部 Stage
		for _, stageReq := range *req.Stages {

			// 3-1. 驗證 Stage ID
			if stageReq.StageID == 0 {
				return newCreateGameError(
					http.StatusBadRequest,
					"invalid_stage_id",
					"更新關卡必須提供 stage_id",
					nil,
				)
			}

			if usedStageIDs[stageReq.StageID] {
				return newCreateGameError(
					http.StatusBadRequest,
					"duplicate_stage_id",
					"同一個關卡不可重複提交",
					nil,
				)
			}
			usedStageIDs[stageReq.StageID] = true

			// 3-2. 驗證 sort_id
			if stageReq.SortID <= 0 {
				return newCreateGameError(
					http.StatusBadRequest,
					"invalid_sort_id",
					"sort_id 必須大於 0",
					nil,
				)
			}

			if _, exists := updatedStages[stageReq.SortID]; exists {
				return newCreateGameError(
					http.StatusBadRequest,
					"duplicate_sort_id",
					"同一遊戲的 sort_id 不可重複",
					nil,
				)
			}

			// 3-3. 查詢既有 Stage，確認它屬於這個遊戲

			existingStage, err := stageRepos.GetByIDAndGameID(ctx, stageReq.StageID, gameID)

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return newCreateGameError(
						http.StatusNotFound,
						"stage_not_found",
						"指定關卡不存在或不屬於此遊戲",
						err,
					)
				}

				return newCreateGameError(
					http.StatusInternalServerError,
					"get_stage_failed",
					"查詢關卡失敗",
					err,
				)
			}

			// 3-4. 處理 Location
			locateReq := stageReq.Locate
			var locationID uint

			if locateReq.LocationID == nil {
				// 建立新地點
				location := &models.Locations{
					CreatorID:   userID,
					Name:        locateReq.Name,
					Description: locateReq.Description,
					Lon:         locateReq.Lon,
					Lat:         locateReq.Lat,
					PublicAt:    locateReq.PublicAt,
				}

				if err := locationsRepos.Create(ctx, location); err != nil {
					return newCreateGameError(
						http.StatusInternalServerError,
						"create_location_failed",
						"建立地點失敗",
						err,
					)
				}

				locationID = location.ID

			} else {
				// 使用既有地點
				location, err := locationsRepos.GetByID(
					ctx,
					*locateReq.LocationID,
				)

				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return newCreateGameError(
							http.StatusNotFound,
							"location_not_found",
							"指定地點不存在",
							err,
						)
					}

					return newCreateGameError(
						http.StatusInternalServerError,
						"get_location_failed",
						"查詢地點失敗",
						err,
					)
				}

				// 檢查是否為自己的地點或已公開地點
				if location.CreatorID != userID &&
					(location.PublicAt == nil ||
						location.PublicAt.After(time.Now())) {

					return newCreateGameError(
						http.StatusForbidden,
						"location_forbidden",
						"沒有權限使用這個地點",
						nil,
					)
				}

				locationID = location.ID
			}

			// 3-5. 更新 Stage
			// 不重新產生 QR Code UID
			updateStage := &models.Stages{
				ID:             existingStage.ID,
				GameID:         gameID,
				LocationId:     locationID,
				SortID:         stageReq.SortID,
				QrcodeUid:      existingStage.QrcodeUid,
				Name:           stageReq.Name,
				Description:    stageReq.Description,
				DetermineRange: stageReq.DetermineRange,
				LimitSeconds:   stageReq.LimitSeconds,
				Quest:          stageReq.Quest,
				IsEnd:          stageReq.IsEnd,
			}

			if err := stageRepos.Update(ctx, updateStage); err != nil {
				return newCreateGameError(
					http.StatusInternalServerError,
					"update_stage_failed",
					"更新關卡失敗",
					err,
				)
			}

			// 記錄 sort_id 與 stage_id 的對應
			updatedStages[stageReq.SortID] = existingStage.ID

			if stageReq.IsStart {
				id := existingStage.ID
				startStageID = &id
			}
		}

		// 4. 第二階段：更新 NextStages
		// 此時全部關卡都已更新完成
		for index, stageReq := range *req.Stages {
			nowStageID := updatedStages[stageReq.SortID]

			var updateNextStages []models.NextStages

			for _, targetSortIDUint := range parsedQuests[index].AllOptionsStageSortID {

				targetSortID := int(targetSortIDUint)

				// 由 sort_id 取得資料庫 stage_id
				targetStageID, exists := updatedStages[targetSortID]

				if !exists {
					return newCreateGameError(
						http.StatusBadRequest,
						"wrong_stage_nextstage",
						"指定的 sort_id 找不到對應關卡",
						nil,
					)
				}

				ns := models.NextStages{
					GameID:        gameID,
					NowStageID:    nowStageID,
					OptionStageId: &targetStageID,
				}

				updateNextStages = append(updateNextStages, ns)
			}

			// 清除這個 Stage 的舊關聯並建立新關聯
			if err := nsRepos.ReplaceByNowStageID(
				ctx,
				nowStageID,
				gameID,
				updateNextStages,
			); err != nil {
				return newCreateGameError(
					http.StatusInternalServerError,
					"update_next_stage_failed",
					"更新關卡關聯失敗",
					err,
				)
			}
		}

		// 5. 更新 Game 起始關卡
		if startStageID == nil {
			return newCreateGameError(
				http.StatusBadRequest,
				"missing_start_stage",
				"必須指定一個起始關卡",
				nil,
			)
		}

		updateGame.StartStageID = startStageID

		if err := gameRepos.Update(ctx, updateGame); err != nil {
			return newCreateGameError(
				http.StatusInternalServerError,
				"update_game_failed",
				"更新遊戲起始關卡失敗",
				err,
			)
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
	c.JSON(http.StatusCreated, gin.H{
		"status":  true,
		"message": "遊戲更新成功",
		"game_id": gameID,
	})
}
