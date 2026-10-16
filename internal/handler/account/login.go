package account

import (
	"amserver/internal/repository"
	account_service "amserver/internal/service/account"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context, db *gorm.DB) {
	ctx := c.Request.Context()
	var req LoginUserRequest

	// 1. 驗證 Request Body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_request",
			"message": "請輸入有效的電子信箱與密碼",
		})
		return
	}

	// 2. 依 Email 查詢使用者
	userRepos := repository.NewUserRepos(db)

	existingUser, err := userRepos.GetByEmail(ctx, req.Email)
	if err != nil {
		// Email 不存在時，也回傳一般登入失敗，避免洩漏帳戶資訊
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"error":   "invalid_credentials",
				"message": "電子信箱或密碼錯誤",
			})
			return
		}

		// 其他錯誤，例如資料庫連線失敗
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "login_failed",
			"message": "登入服務發生錯誤，請稍後再試",
		})
		return
	}

	// 3. 確認帳戶已設定密碼
	if existingUser.Password_hash == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "invalid_credentials",
			"message": "電子信箱或密碼錯誤",
		})
		return
	}

	// 4. 驗證密碼
	if !account_service.CheckPasswordHash(
		req.Password,
		*existingUser.Password_hash,
	) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "invalid_credentials",
			"message": "電子信箱或密碼錯誤",
		})
		return
	}

	// 5. 產生 JWT
	token, err := account_service.GenerateJWT(
		existingUser.ID,
		existingUser.Name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "token_generation_failed",
			"message": "登入憑證產生失敗，請稍後再試",
		})
		return
	}

	// 6. 登入成功
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"token":  token,
	})
}
