package routes

import (
	controllers "ares_api/internal/api/controllers"
	"ares_api/internal/middleware"
	"ares_api/internal/ollama"
	repositories "ares_api/internal/repositories"
	service "ares_api/internal/services"

	"fmt"
	"os"
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
	// MEMORY MODULE
	// --------------------------
	memoryRepo := repositories.NewMemoryRepository(db)
	memoryService := service.NewMemoryService(memoryRepo)
	memoryController := controllers.NewMemoryController(memoryService, ledgerService)

	// --------------------------
	// EMBEDDING SERVICE (Semantic Memory)
	// --------------------------
	embeddingService := service.NewEmbeddingService(memoryRepo)

	// --------------------------
	// CLAUDE AI MODULE (Now powered by Ollama + DeepSeek-R1)
	// --------------------------
	repoPath := os.Getenv("ARES_REPO_PATH")
	if repoPath == "" {
		repoPath = "C:/ARES_Workspace" // fallback - full workspace access
	}
	// Switch to Ollama-powered implementation (no API key required!)
	claudeService := service.NewClaudeServiceOllama(memoryRepo, embeddingService, repoPath)
	claudeController := controllers.NewClaudeController(claudeService, ledgerService)

	// --------------------------
	// SCANNER MODULE (File Fragment Recovery)
	// --------------------------
	scannerService := service.NewScannerService(memoryRepo)
	scannerController := controllers.NewScannerController(scannerService)

	// --------------------------
	// EDITOR MODULE (Monaco Editor File Operations)
	// --------------------------
	workspaceRoot := os.Getenv("ARES_WORKSPACE_ROOT")
	if workspaceRoot == "" {
		workspaceRoot = "C:/ARES_Workspace" // fallback
	}
	editorService := service.NewEditorService(workspaceRoot)
	editorController := controllers.NewEditorController(editorService)

	// --------------------------
	// LLM MODULE (Local LLM Integration)
	// --------------------------
	llmController := controllers.NewLLMController()

	// --------------------------
	// BACKUP MODULE (Database Export/Import)
	// --------------------------
	backupController := controllers.NewBackupController(db)

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
	//  BACKGROUND JOB TO PROCESS MEMORY EMBEDDINGS
	// --------------------------
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Process embeddings every 30 seconds
		defer ticker.Stop()

		for range ticker.C {
			processed, _ := embeddingService.ProcessEmbeddingQueue(50) // Process 50 at a time
			if processed > 0 {
				fmt.Printf("📊 Processed %d memory embeddings\n", processed)
			}
		}
	}()

	// --------------------------
	//  BACKGROUND JOB TO CONSOLIDATE OLD MEMORIES
	// --------------------------
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run daily
		defer ticker.Stop()

		for range ticker.C {
			consolidated, err := embeddingService.ConsolidateOldMemories(30, 0.85)
			if err != nil {
				fmt.Printf("⚠️ Memory consolidation error: %v\n", err)
			} else if consolidated > 0 {
				fmt.Printf("🗜️ Consolidated %d old memory groups\n", consolidated)
			}
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
		trades.GET("/performance", tradeController.GetPerformance)
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

	// --------------------------
	// Memory endpoints
	// --------------------------
	memory := api.Group("/memory")
	memory.Use(middleware.AuthMiddleware())
	{
		memory.POST("/learn", memoryController.Learn)
		memory.GET("/recall", memoryController.Recall)
		memory.POST("/import", memoryController.ImportConversation)
	}

	// --------------------------
	// Claude AI endpoints - Stateful consciousness with full context
	// --------------------------
	claude := api.Group("/claude")
	claude.Use(middleware.AuthMiddleware())
	{
		claude.POST("/chat", claudeController.Chat)
		claude.GET("/memory", claudeController.GetMemory)
		claude.POST("/file", claudeController.ReadFile)
		claude.GET("/repository", claudeController.GetRepositoryContext)
		// Semantic memory endpoints
		claude.POST("/semantic-search", claudeController.SemanticSearch)
		claude.POST("/process-embeddings", claudeController.ProcessEmbeddings)
	}

	// --------------------------
	// Scanner endpoints - File fragment recovery for Solace/ARES
	// --------------------------
	scanner := api.Group("/scanner")
	scanner.Use(middleware.AuthMiddleware())
	{
		scanner.POST("/scan", scannerController.ScanFiles)
		scanner.POST("/import", scannerController.ImportFragments)
		scanner.POST("/solace", scannerController.ImportSolaceData)
	}

	// --------------------------
	// Editor endpoints - Monaco Editor file operations
	// --------------------------
	editor := api.Group("/editor")
	editor.Use(middleware.AuthMiddleware())
	{
		editor.POST("/read", editorController.ReadFile)
		editor.POST("/save", editorController.SaveFile)
		editor.POST("/list", editorController.ListFiles)
		editor.POST("/create", editorController.CreateFile)
		editor.POST("/delete", editorController.DeleteFile)
		editor.POST("/rename", editorController.RenameFile)
	}

	// --------------------------
	// LLM endpoints - Local LLM integration (Ollama)
	// --------------------------
	llm := api.Group("/llm")
	llm.Use(middleware.AuthMiddleware())
	{
		llm.GET("/status", llmController.GetStatus)
		llm.POST("/test", llmController.TestInference)
	}

	// --------------------------
	// Backup endpoints - Database export/import
	// --------------------------
	backup := api.Group("/backup")
	backup.Use(middleware.AuthMiddleware())
	{
		backup.GET("/export", backupController.Export)
		backup.POST("/import", backupController.Import)
	}
}
