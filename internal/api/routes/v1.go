package routes

import (
	controllers "ares_api/internal/api/controllers"
	"ares_api/internal/middleware"
	"ares_api/internal/ollama"
	repositories "ares_api/internal/repositories"
	service "ares_api/internal/services"

	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes sets up all API routes with their dependencies
func RegisterRoutes(r *gin.Engine, db *gorm.DB) {

	// --------------------------
	// LEDGER MODULE
	// --------------------------
	ledgerRepo := repositories.NewLedgerRepository(db)
	ledgerService := service.NewLedgerService(ledgerRepo)

	// --------------------------
	// USER MODULE
	// --------------------------
	userRepo := repositories.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controllers.NewUserController(userService, ledgerService)

	// --------------------------
	// BALANCE MODULE
	// --------------------------
	balanceRepo := repositories.NewBalanceRepository(db)
	balanceService := service.NewBalanceService(balanceRepo)
	balanceController := controllers.NewBalanceController(balanceService , ledgerService)

	// --------------------------
	// CHAT MODULE
	// --------------------------
	ollamaService := ollama.NewClientFromEnv()
	chatRepo := repositories.NewChatRepository(db)
	chatService := service.NewChatService(chatRepo, ollamaService)
	chatController := controllers.NewChatController(chatService, ledgerService)

	// --------------------------
	//  ASSETS MODULE
	// --------------------------
	assetRepo := repositories.NewAssetRepository()
	assetService := service.NewAssetService(assetRepo)
	assetContoller := controllers.NewAssetController(assetService , ledgerService)
	// --------------------------
	// TRADE MODULE
	// --------------------------
	tradeRepo := repositories.NewTradeRepository(db)
	tradeService := service.NewTradeService(tradeRepo, balanceRepo, assetRepo)
	tradeController := controllers.NewTradeController(tradeService, ledgerService)

	// --------------------------
	// SETTINGS MODULE
	// --------------------------
	settingsRepo := repositories.NewSettingsRepository(db)
	settingsService := service.NewSettingsService(settingsRepo)
	settingsController := controllers.NewSettingsController(settingsService, ledgerService)

	// --------------------------
	//  BACKGROUND JOB TO PROCESS OPEN LIMIT ORDERS
	// --------------------------
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			tradeService.ProcessOpenLimitOrders()
		}
	}()

	// --------------------------
	//  API V1 GROUP
	// --------------------------
	api := r.Group("/api/v1")
	api.Use(middleware.CORSMiddleware())

	// --------------------------
	//  User endpoints
	// --------------------------
	users := api.Group("/users")
	{
		users.POST("/signup", userController.Signup)
		users.POST("/login", userController.Login)
		users.POST("/refresh", userController.RefreshToken)
	}

	// --------------------------
	// Chat endpoints
	// --------------------------
	chats := api.Group("/chat")
	chats.Use(middleware.AuthMiddleware())
	{
		chats.POST("/send", chatController.SendMessage)
		chats.GET("/history", chatController.GetHistory)
	}

	// --------------------------
	// Trade endpoints
	// --------------------------
	trades := api.Group("/trades")
	trades.Use(middleware.AuthMiddleware())
	{
		trades.POST("/market", tradeController.MarketOrder)
		trades.POST("/limit", tradeController.LimitOrder)
		trades.GET("/history", tradeController.GetHistory)
		trades.GET("/pending", tradeController.GetPendingLimitOrders)
	}

	// --------------------------
	// Settings endpoints
	// --------------------------
	settings := api.Group("/settings")
	settings.Use(middleware.AuthMiddleware())
	{

		settings.POST("/apikey", settingsController.SaveAPIKey)

	}

	// --------------------------
	// Balance endpoints
	balances := api.Group("/balances")
	balances.Use(middleware.AuthMiddleware())
	{
		balances.GET("/", balanceController.GetUSDBalance)
		balances.POST("/init", balanceController.InitializeBalance)
		balances.POST("/reset", balanceController.ResetBalance)
		balances.POST("/update", balanceController.UpdateBalance)
	}
	// --------------------------
	// Asset endpoints
	assets := api.Group("/assets")
	{

		assets.GET("/coins", assetContoller.GetAllCoins)
		assets.GET("/coins/:id/market", assetContoller.GetCoinMarket)
		assets.GET("/coins/top-movers", assetContoller.GetTopMovers)
		assets.GET("/vs_currencies", assetContoller.GetSupportedVSCurrencies)
	}
}
