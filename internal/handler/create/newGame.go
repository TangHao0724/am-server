package create

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"amserver/internal/models"
	"amserver/internal/repository"
	"amserver/internal/service"
	"amserver/internal/service/questjson"
)

type CreateGameRequest struct {
	Name         string                `json:"name" binding:"required,min=1,max=50"`
	GameCode     string                `json:"game_code"`
	AreaGeojson  *datatypes.JSON       `json:"area_geojson"`
	Setting      *datatypes.JSON       `json:"setting"`
	LimitSeconds *int64                `json:"limit_seconds" binding:"omitempty,gte=0"`
	IsPublic     *bool                 `json:"is_public"`
	Description  *string               `json:"description"`
	PublishAt    *time.Time            `json:"publish_at"`
	Stages       *[]CreateStageRequest `json:"stages"`
}

type CreateStageRequest struct {
	Name           *string                `json:"name"`
	SortID         int                    `json:"sort_id"`
	IsStart        bool                   `json:"is_start"`
	Description    *string                `json:"description"`
	DetermineRange *float64               `json:"determine_range" binding:"omitempty,gte=0"`
	LimitSeconds   *int64                 `json:"limit_seconds" binding:"omitempty,gte=0"`
	Quest          *datatypes.JSON        `json:"quest"`
	IsEnd          bool                   `json:"is_end"`
	Locate         *CreateLocationRequest `json:"locate"`
}

type CreateLocationRequest struct {
	ID          *uint      `json:"id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Lon         *float64   `json:"lon" binding:"omitempty,gte=-180,lte=180"`
	Lat         *float64   `json:"lat" binding:"omitempty,gte=-90,lte=90"`
	PublicAt    *time.Time `json:"public_at"`
}

func NewGameHandler(c *gin.Context, db *gorm.DB) {
	// 1. 從 JWT middleware 取得 userID
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

	// 2. 解析 Request
	var req CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_request",
			"message": "請輸入正確格式的請求",
		})
		return
	}

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

		if locate.ID != nil {
			// 使用既有地點時，不允許同時傳入新地點資料
			if *locate.ID == 0 ||
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

		// 目前 AllOptionsStageID 使用 []int，
		// 寫入 uint 欄位前至少要確定 ID 不是負數。
		// for _, targetID := range parsedQuests[index].AllOptionsStageID {
		// 	if targetID < 0 {
		// 		c.JSON(http.StatusBadRequest, gin.H{
		// 			"status":  false,
		// 			"error":   "invalid_option_stage_id",
		// 			"message": "關聯關卡 ID 不可為負數",
		// 			"index":   index,
		// 		})
		// 		return
		// 	}
		// }
	}

	// 本版本採用 IsStart 決定起始關卡，因此必須剛好有一個。
	if startCount > 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_start_stage",
			"message": "只能設定一個起始關卡",
		})
		return
	}

	// 5. 開始 Transaction
	var createdGameID uint
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		gameRepos := repository.NewGameRepos(tx)
		stageRepos := repository.NewStagesRepos(tx)
		locationsRepos := repository.NewLocationRepos(tx)
		nsRepos := repository.NewNextStageRepos(tx)

		// 1. 建立 Game
		baseGame := &models.Games{
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

		if err := gameRepos.Create(ctx, baseGame); err != nil {
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
				"create_game_failed",
				"建立遊戲失敗",
				err,
			)
		}

		createdGameID = baseGame.ID

		// sort_id -> 實際資料庫 stage_id
		createdStages := make(map[int]uint)

		var startStageID *uint

		// 2. 建立全部 Location 與 Stage
		for _, stageReq := range *req.Stages {
			locateReq := stageReq.Locate

			uid, err := service.GenerateRandomString(32)
			if err != nil {
				return newCreateGameError(
					http.StatusInternalServerError,
					"generate_uid_failed",
					"產生 QR Code 識別碼失敗",
					err,
				)
			}

			var location *models.Locations

			if locateReq.ID != nil {
				source, err := locationsRepos.GetByID(ctx, *locateReq.ID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return newCreateGameError(
							http.StatusBadRequest,
							"location_not_found",
							"指定的地點不存在",
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

				// 複製既有地點
				location = &models.Locations{
					CreatorID:   userID,
					Name:        source.Name,
					Description: source.Description,
					Lon:         source.Lon,
					Lat:         source.Lat,
					PublicAt:    source.PublicAt,
				}
			} else {
				location = &models.Locations{
					CreatorID:   userID,
					Name:        locateReq.Name,
					Description: locateReq.Description,
					Lon:         locateReq.Lon,
					Lat:         locateReq.Lat,
					PublicAt:    locateReq.PublicAt,
				}
			}

			if err := locationsRepos.Create(ctx, location); err != nil {
				return newCreateGameError(
					http.StatusInternalServerError,
					"create_location_failed",
					"建立地點失敗",
					err,
				)
			}

			// 建立 Stage
			stage := &models.Stages{
				GameID:         baseGame.ID,
				LocationId:     location.ID,
				SortID:         stageReq.SortID,
				QrcodeUid:      &uid,
				Name:           stageReq.Name,
				Description:    stageReq.Description,
				DetermineRange: stageReq.DetermineRange,
				LimitSeconds:   stageReq.LimitSeconds,
				Quest:          stageReq.Quest,
				IsEnd:          stageReq.IsEnd,
			}

			if err := stageRepos.Create(ctx, stage); err != nil {
				return newCreateGameError(
					http.StatusInternalServerError,
					"create_stage_failed",
					"建立關卡失敗",
					err,
				)
			}

			// 檢查 sort_id 是否重複
			if _, exists := createdStages[stageReq.SortID]; exists {
				return newCreateGameError(
					http.StatusBadRequest,
					"duplicate_sort_id",
					"同一遊戲的關卡 sort_id 不可重複",
					nil,
				)
			}

			// 記錄 sort_id 對應的資料庫 ID
			createdStages[stageReq.SortID] = stage.ID

			if stageReq.IsStart {
				id := stage.ID
				startStageID = &id
			}
		}

		// 3. 所有 Stage 建立完成後，才建立 NextStages
		for index, stageReq := range *req.Stages {
			nowStageID := createdStages[stageReq.SortID]

			for _, targetSortIDUint := range parsedQuests[index].AllOptionsStageSortID {
				targetSortID := int(targetSortIDUint)

				// 從 map 取得實際資料庫 ID
				targetStageID, exists := createdStages[targetSortID]

				if !exists {
					return newCreateGameError(
						http.StatusBadRequest,
						"wrong_stage_nextstage",
						"指定的 sort_id 找不到對應關卡",
						nil,
					)
				}

				ns := &models.NextStages{
					GameID:        baseGame.ID,
					NowStageID:    nowStageID,
					OptionStageId: &targetStageID,
				}

				if err := nsRepos.Create(ctx, ns); err != nil {
					if errors.Is(err, repository.ErrDuplicateNextStageSort) {
						return newCreateGameError(
							http.StatusConflict,
							"duplicate_next_stage",
							"關卡關聯重複",
							err,
						)
					}

					return newCreateGameError(
						http.StatusInternalServerError,
						"create_next_stage_failed",
						"建立關卡關聯失敗",
						err,
					)
				}
			}
		}

		// 4. 更新起始關卡

		if startStageID == nil {
			return newCreateGameError(
				http.StatusBadRequest,
				"missing_start_stage",
				"未能取得起始關卡",
				nil,
			)
		}

		baseGame.StartStageID = startStageID

		if err := gameRepos.Update(ctx, baseGame); err != nil {
			return newCreateGameError(
				http.StatusInternalServerError,
				"update_game_failed",
				"更新遊戲起始關卡失敗",
				err,
			)
		}

		return nil
	})

	// 6. Transaction 外統一回覆錯誤
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

	// 7. 全部操作成功，才回覆成功
	c.JSON(http.StatusCreated, gin.H{
		"status":  true,
		"message": "遊戲建立成功",
		"game_id": createdGameID,
	})
}
