package account

import (
	"amserver/internal/models"
	"amserver/internal/repository"
	account_service "amserver/internal/service/account"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateUserAnonymousRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

func AnonymousRigister(c *gin.Context, db *gorm.DB) {
	ctx := c.Request.Context()

	var req CreateUserAnonymousRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_request",
			"message": "請輸入有效的暱稱，長度不可超過 50 個字元",
		})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_request",
			"message": "請輸入暱稱",
		})
		return
	}

	userRepos := repository.NewUserRepos(db)
	user := &models.User{
		Name: name,
	}

	if err := userRepos.Create(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "create_user_failed",
			"message": "建立匿名帳戶失敗，請稍後再試",
		})
		return
	}

	token, err := account_service.GenerateJWT(user.ID, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "token_generation_failed",
			"message": "帳戶已建立，但登入憑證產生失敗，請稍後再試",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": true,
		"token":  token,
	})
}
