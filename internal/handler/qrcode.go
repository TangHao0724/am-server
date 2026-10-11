package handler

import (
	"io"
	"net/http"
	"strconv"

	"amserver/internal/repository"
	"amserver/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func QRCodeHadler(c *gin.Context, db *gorm.DB) {
	// 1. 取得 URL 參數。
	gameID64, err := strconv.ParseInt(c.Param("game_id"), 10, 32)
	if err != nil || gameID64 <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid game_id",
		})
		return
	}

	// stage_id 在此假設是 QR Code UUID。
	qrcodeUID := c.Param("qrcode_uid")
	if qrcodeUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing QR Code UUID",
		})
		return
	}

	// 2. 取得玩家座標。
	var req UserlocationRequest
	if c.Request.Body != nil {
		err := c.ShouldBindJSON(&req)

		if err != nil && err != io.EOF {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body",
			})
			return
		}
	}

	// 3. 身份驗證。
	// [非 Repository] JWT middleware 應負責驗證 token。
	// 這裡假設 middleware 將 userID 放進 Gin Context。
	rawUserID, exists := c.Get("userID")

	if !exists {
		c.Redirect(302, "https://service.auqamarine.date/welecome/"+strconv.FormatInt(gameID64, 10))
		return
	}

	userID, ok := rawUserID.(uint)
	if !ok || userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user identity",
		})
		return
	}

	// 4. [Repository] 建立 Repository 物件。
	stageRepo := repository.NewStageRepository(db)
	progressRepo := repository.NewGamePlayStagesRepository(db)

	// 5. [Service] 執行關卡業務邏輯。
	qrService := service.NewQRCodeService(
		stageRepo,
		progressRepo,
	)

	result, err := qrService.ScanStage(service.ScanStageInput{
		GameID:    int32(gameID64),
		UserID:    userID,
		QRCodeUID: qrcodeUID,
		Lon:       req.Lon,
		Lat:       req.Lat,
	})

	if err != nil {
		switch err {
		case service.ErrStageNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "stage not found",
			})
		case service.ErrWrongStage:
			c.JSON(http.StatusConflict, gin.H{
				"error": "stage is not the expected next stage",
			})
		case service.ErrNoStages:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "game has no stages",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to process stage",
			})
		}
		return
	}

	// 6. [非 傳Repository] 回 HTTP 結果。
	c.JSON(http.StatusOK, result)
}
