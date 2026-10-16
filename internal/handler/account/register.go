package account

import (
	"errors"
	"net/http"

	"amserver/internal/models"
	"amserver/internal/repository"
	account_service "amserver/internal/service/account"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func Register(c *gin.Context, db *gorm.DB) {
	ctx := c.Request.Context()

	// 1. 解析並驗證 Request Body
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_request",
			"message": "帳戶資料格式不正確，請檢查暱稱、電子信箱與密碼",
		})
		return
	}

	// 2. 加密密碼
	pwHash, err := account_service.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "password_hash_failed",
			"message": "密碼處理失敗，請稍後再試",
		})
		return
	}

	// 3. 建立使用者
	user := &models.User{
		Name:          req.Name,
		Email:         &req.Email,
		Password_hash: &pwHash,
	}

	userRepos := repository.NewUserRepos(db)

	if err := userRepos.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, gin.H{
				"status":  false,
				"error":   "email_already_exists",
				"message": "此電子信箱已被其他帳戶使用",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "create_user_failed",
			"message": "建立帳戶失敗，請稍後再試",
		})
		return
	}

	// 4. 產生 JWT
	token, err := account_service.GenerateJWT(user.ID, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "token_generation_failed",
			"message": "帳戶已建立，但登入憑證產生失敗，請稍後嘗試登入",
		})
		return
	}

	// 5. 回傳成功結果
	c.JSON(http.StatusCreated, gin.H{
		"status": true,
		"token":  token,
	})
}
