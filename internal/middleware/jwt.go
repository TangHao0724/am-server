package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	account_service "amserver/internal/service/account"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		fmt.Println("Authorization ：", authHeader)
		fmt.Println("Authorization 是否存在：", authHeader != "")
		fmt.Println("Authorization 長度：", len(authHeader))
		// 檢查header
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		// 檢查header中的字串是否bearer開頭
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		// 解析並對必
		claims, err := account_service.ParseJWT(tokenString)
		if err != nil {
			// 正式環境建議記錄錯誤類型，不要記錄完整 token 或 Secret。
			log.Printf("JWT validation failed: %v", err)

			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"error":   "invalid_token",
				"message": "登入憑證無效或已過期，請重新登入",
			})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()

	}
}

func OptionalJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 1. 沒有 Authorization，視為匿名玩家。
		if authHeader == "" {
			c.Next()
			return
		}

		// 2. 解析 Bearer Token。
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			c.Next()
			return
		}

		tokenString := strings.TrimSpace(parts[1])

		// 3. 驗證 JWT。
		token, err := jwt.ParseWithClaims(
			tokenString,
			&account_service.Claims{},
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}

				return []byte("aquamarina"), nil
			},
		)

		// 4. JWT 無效時，視為匿名玩家。
		if err != nil || token == nil || !token.Valid {
			c.Next()
			return
		}

		// 5. JWT 有效，才將 userID 存入 Context。
		claims, ok := token.Claims.(*account_service.Claims)
		if !ok || !token.Valid {
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
