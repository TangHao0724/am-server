package main

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"amserver/internal/database"
	"amserver/internal/handler"
	"amserver/internal/handler/account"
	"amserver/internal/handler/create"
	"amserver/internal/handler/game"
	"amserver/internal/middleware"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		panic(fmt.Sprintf("failed to connect db server: %v", err))
	}
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
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
	authorized.Use(middleware.JWTMiddleware())
	{
		// 單關資料，用在進入遊戲或者下一關後，呼叫
		authorized.GET("/stages/:game_id/:stage_id", func(ctx *gin.Context) {
			handler.GetStageData(ctx, db)
		})

		// 做完遊戲時，呼叫會產生JSON (STAGe 資料與二進制檔案)，展示用
		authorized.GET("/qrcode/:game_id/:stage_id", func(ctx *gin.Context) {
			handler.GetStageQRCode(ctx, db)
		})

		// 建立遊戲測試用
		authorized.POST("/create", func(ctx *gin.Context) {
			handler.CreateGameHandler(ctx, db)
		})
		// 遊戲編輯相關
		// 建立遊戲
		authorized.POST("/create/new", func(ctx *gin.Context) {
			create.NewGameHandler(ctx, db)
		})
		// 查詢遊戲，用作進入遊戲設置、編輯器
		authorized.GET("/create/:gameid", func(ctx *gin.Context) {
			create.LoadGameHandler(ctx, db)
		})
		// 編輯器內更新，用作進入遊戲設置、編輯器
		authorized.PUT("/create/:gameid", func(ctx *gin.Context) {
			create.UpdateGameHandler(ctx, db)
		})
		// 移除關卡
		authorized.DELETE("/create/:gameid", func(ctx *gin.Context) {
			create.DeleteGameHandler(ctx, db)
		})

		// 遊玩類型
		// 建立新遊戲
		authorized.POST("/game/join/:gamecode", func(ctx *gin.Context) {
			game.JoinGame(ctx, db)
		})
		authorized.GET("gameplays/:gameplayid/stages/current", func(ctx *gin.Context) {
			game.CurrentStage(ctx, db)
		})

		// 帳號相關
		// 將測試帳號升級
		authorized.POST("/user/updateUser", func(ctx *gin.Context) {
			account.UpdateUser(ctx, db)
		})

	}
	// 掃碼驗證
	router.GET(
		"/qrcode/:game_id/:qrcode_uid",
		func(c *gin.Context) {
			handler.QRCodeLandingHandler(c)
		},
	)

	// 網頁內掃描器：提交 GPS 與 JWT
	router.POST(
		"/qrcode/:game_id/:qrcode_uid",
		middleware.OptionalJWTMiddleware(),
		func(c *gin.Context) {
			handler.QRCodeHadler(c, db)
		},
	)
	// 註冊。登入相關
	//簡易註冊
	router.POST("/user/anonymous", func(c *gin.Context) {
		account.AnonymousRigister(c, db)
	})
	//正式註冊
	router.POST("/user/register", func(c *gin.Context) {
		account.Register(c, db)
	})
	// 登入
	router.POST("/user/login", func(c *gin.Context) {
		account.Login(c, db)
	})

	if err := router.Run(":8080"); err != nil {
		panic(fmt.Sprintf("failed to start server: %v", err))
	}

}
