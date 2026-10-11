package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"amserver/internal/database"
	"amserver/internal/repository"

	"amserver/internal/handler"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		panic(fmt.Sprintf("failed to connect db server: %v", err))
	}
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://service.auqamarine.date",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
	}))
	authorized := router.Group("/api")
	authorized.Use(JWTMiddleware())
	{
		// 單關資料，用在進入遊戲或者下一關後，呼叫
		authorized.GET("/stages/:game_id/:stage_id", func(ctx *gin.Context) {
			handler.GetStageData(ctx, db)
		})

		// 做完遊戲時，呼叫會產生JSON (STAGe 資料與二進制檔案)，展示用
		authorized.GET("/qrcode/:game_id/:stage_id", func(ctx *gin.Context) {
			handler.GetStageQRCode(ctx, db)
		})

		// 建立遊戲
		authorized.POST("/create", func(ctx *gin.Context) {
			handler.CreateGameHandler(ctx, db)
		})

	}
	// 掃碼驗證
	router.GET(
		"/qrcode/:game_id/:qrcode_uid",
		OptionalJWTMiddleware(),
		func(c *gin.Context) {
			handler.QRCodeHadler(c, db)
		},
	)
	router.POST("/user", func(c *gin.Context) {
		handler.Register(c, db)
	})
	router.GET("/user/:id", func(c *gin.Context) {
		ctx := c.Request.Context()
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if db, err := database.Connect(); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		} else {
			userRepos := repository.NewUserRepository(db)
			userRepos.GetByID(ctx, id)
		}
	})

	if err := router.Run(":8080"); err != nil {
		panic(fmt.Sprintf("failed to start server: %v", err))
	}
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

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
		token, err := jwt.ParseWithClaims(
			tokenString,
			&handler.Claims{},
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, fmt.Errorf("unexpected signing method")
				}

				return []byte("aquamarina"), nil
			},
		)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*handler.Claims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
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
			&handler.Claims{},
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
		claims, ok := token.Claims.(*handler.Claims)
		if !ok || !token.Valid {
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
