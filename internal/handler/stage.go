package handler

import (
	"amserver/internal/repository"
	"amserver/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

type UserlocationRequest struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

func GetStageData(c *gin.Context, db *gorm.DB) {
	stageid := c.Param("stage_id")
	gameid := c.Param("game_id")
	stageID, err := strconv.ParseUint(stageid, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid stage_id",
		})
		return
	}
	gameID, err := strconv.ParseUint(gameid, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid game_id",
		})
		return
	}
	var req UserlocationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	stageRepos := repository.NewStageRepository(db)
	stage, err := stageRepos.GetByID(c, uint(stageID), uint(gameID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	geoData := service.CalculateLocation(service.Userlocation{Lat: req.Lat, Lon: req.Lon}, *stage.Lat, *stage.Lon)
	c.JSON(http.StatusOK, gin.H{
		"name":      *stage.Name,
		"desc":      "一些對地點的描述",
		"img":       "S3圖片的連結，放置地圖的位置連結",
		"distance":  geoData.DistanceMeters,
		"direction": geoData.Direction,
		"degree":    geoData.BearingDegrees,
	})

}
func GetStageQRCode(c *gin.Context, db *gorm.DB) {
	stageID, err := strconv.ParseUint(
		c.Param("stage_id"), 10, 32,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid stage_id",
		})
		return
	}

	gameID, err := strconv.ParseUint(
		c.Param("game_id"), 10, 32,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid game_id",
		})
		return
	}

	stageRepos := repository.NewStageRepository(db)

	stage, err := stageRepos.GetByID(
		c.Request.Context(),
		uint(stageID),
		uint(gameID),
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "stage not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve stage",
		})
		return
	}

	if stage.QrcodeUid == nil || *stage.QrcodeUid == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "stage QR code UID is missing",
		})
		return
	}

	png, err := qrcode.Encode(
		"http://54.54.68.45:8080/qrcode/"+strconv.FormatUint(gameID, 10)+"/"+*stage.QrcodeUid,
		qrcode.Medium,
		256,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate QR code",
		})
		return
	}

	c.Data(
		http.StatusOK,
		"image/png",
		png,
	)
}
