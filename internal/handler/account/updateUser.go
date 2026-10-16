package account

import (
	"amserver/internal/repository"
	account_service "amserver/internal/service/account"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateUserRequest struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=50"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8"`
}

// 在持有JWT的情況下，將匿名登入升級成正式帳戶
func UpdateUser(c *gin.Context, db *gorm.DB) {

	fmt.Println("進入ｈａｎｄｅｌｒ")
	// 從 JWT 獲取 userID
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

	// 取得 Request Body
	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_request",
			"message": "帳戶資料格式不正確",
		})
		return
	}
	ctx := c.Request.Context()

	// 檢查使用者是否存在
	userRepos := repository.NewUserRepos(db)

	existingUser, err := userRepos.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "get_user_failed",
			"message": "讀取使用者資料失敗",
		})
		return
	}

	// 確認使用者目前仍為匿名帳戶
	if existingUser.Password_hash != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  false,
			"error":   "account_already_upgraded",
			"message": "此帳戶已升級為正式帳戶",
		})
		return
	}

	// 加密密碼
	pw_hash, err := account_service.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "pw_pattern_error",
			"message": "密碼格式錯誤",
		})
		return
	}
	// 更新進資料庫
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if len(name) == 0 || len([]rune(name)) > 50 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"error":   "name_length_error",
				"message": "暱稱長度必須為 1 到 50 個字元",
			})
			return
		}

		existingUser.Name = name
	}
	existingUser.Email = &req.Email
	existingUser.Password_hash = &pw_hash

	if err := userRepos.Update(ctx, existingUser); err != nil {
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
			"error":   "update_user_failed",
			"message": "更新帳戶失敗",
		})
		return
	}
	// 產生token
	token, err := account_service.GenerateJWT(userID, existingUser.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"error":   "generate_tokn_failed",
			"message": "產生憑證失敗",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"token":  token,
	})
}
