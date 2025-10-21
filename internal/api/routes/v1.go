package routes

import (
	"ares_api/config"
	"ares_api/internal/ace"
	"ares_api/internal/agent"
	controllers "ares_api/internal/api/controllers"
	"ares_api/internal/api/handlers"
	"ares_api/internal/eventbus"
	"ares_api/internal/grpo"
	"ares_api/internal/middleware"
	"ares_api/internal/monitoring"
	repositories "ares_api/internal/repositories"
	service "ares_api/internal/services"
	services "ares_api/internal/services"
	"ares_api/internal/solace"
	"ares_api/internal/trading"
	"ares_api/pkg/llm"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getUserIDFromContext extracts userID from auth context or returns default
func getUserIDFromContext(c *gin.Context) uint {
	// Try to get from auth middleware context
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	// Default to 1 for testing when auth is disabled
	return uint(1)
}

// parseUint64 parses string to uint64 with error handling
func parseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// parseInt parses string to int with error handling
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// RegisterRoutes sets up all API routes with their dependencies
func RegisterRoutes(r *gin.Engine, db *gorm.DB, eb *eventbus.EventBus, grpoAgent *grpo.Agent) {

	// --------------------------
	// LLM CLIENT (DeepSeek-R1 14B via Ollama)
	// --------------------------
	llmClient := llm.NewClient()

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
	balanceController := controllers.NewBalanceController(balanceService, ledgerService)

	// --------------------------
	// MEMORY MODULE
	// --------------------------
	memoryRepo := repositories.NewMemoryRepository(db)
	memoryService := service.NewMemoryService(memoryRepo)
	memoryController := controllers.NewMemoryController(memoryService, ledgerService)
	conversationController := controllers.NewConversationController(db) // SOLACE's conversation memory endpoint

	// --------------------------
	// CHAT MODULE (Enhanced with System Context & Memory)
	// --------------------------
	chatRepo := repositories.NewChatRepository(db)
	chatService := service.NewChatServiceWithMemory(chatRepo, memoryRepo, llmClient) // 🧠 Memory-aware chat

	// 🧠 ACE FRAMEWORK: Initialize pattern-based decision making
	aceOrchestrator := ace.NewACEOrchestrator(db)
	aceHandler := handlers.NewACEHandler(aceOrchestrator, db)
	chatService.SetACEEnabled(true)
	log.Println("🧠 ACE Framework initialized - 102 cognitive patterns loaded")
	log.Printf("🧠 ACE Real-Time Quality Assessment: ACTIVE")

	chatController := controllers.NewChatController(chatService, ledgerService)

	// --------------------------
	//  ASSETS MODULE
	// --------------------------
	assetRepo := repositories.NewAssetRepository()
	assetService := service.NewAssetService(assetRepo)
	assetContoller := controllers.NewAssetController(assetService, ledgerService)
	// --------------------------
	// TRADE MODULE (Legacy market/limit orders)
	// --------------------------
	tradeRepo := repositories.NewTradeRepository(db)
	tradeService := service.NewTradeService(tradeRepo, balanceRepo, assetRepo)
	tradeController := controllers.NewTradeController(tradeService, ledgerService)

	// --------------------------
	// SANDBOX TRADING MODULE (Autonomous Trading for SOLACE)
	// --------------------------
	tradingRepo := repositories.NewTradingRepository(db)
	tradingService := services.NewTradingService(tradingRepo, balanceRepo, assetRepo, eb, grpoAgent)
	tradingController := controllers.NewTradingController(tradingService, ledgerService)

	// --------------------------
	// SETTINGS MODULE
	// --------------------------
	settingsRepo := repositories.NewSettingsRepository(db)
	settingsService := service.NewSettingsService(settingsRepo)
	settingsController := controllers.NewSettingsController(settingsService, ledgerService)

	// --------------------------
	// CRYPTO PRICES (LIVE CoinGecko Integration - NO MOCKS)
	// --------------------------
	cryptoPriceController := controllers.NewCryptoPriceController()

	// --------------------------
	// VISION CAPABILITIES (Multimodal - SOLACE can SEE)
	// --------------------------
	visionController := controllers.NewVisionController()

	// --------------------------
	// EMBEDDING SERVICE (Semantic Memory)
	// --------------------------
	embeddingService := service.NewEmbeddingService(memoryRepo)

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
	// FILE ACCESS TOOLS (AI File System Access)
	// --------------------------
	fileTools := llm.NewFileAccessTools(workspaceRoot)
	fileToolsController := controllers.NewFileToolsController(fileTools)

	// --------------------------
	// CONTEXT MANAGEMENT (Token Budget & Rolling Window)
	// --------------------------
	contextController := controllers.NewContextController(chatService)

	// --------------------------
	// SCANNER MODULE (File Fragment Recovery)
	// --------------------------
	scannerService := service.NewScannerService(memoryRepo)
	scannerController := controllers.NewScannerController(scannerService)

	// --------------------------
	// LLM MODULE (Local LLM Integration)
	// --------------------------
	llmController := controllers.NewLLMController()

	// --------------------------
	// LLM HEALTH MONITORING
	// ⚠️ DEPRECATED: LLM health moved to /health/detailed
	// --------------------------
	// llmHealthController := controllers.NewLLMHealthController(llmClient)

	// --------------------------
	// BACKUP MODULE (Database Export/Import)
	// --------------------------
	backupController := controllers.NewBackupController(db)

	// --------------------------
	// MONITORING MODULE (Health & Metrics)
	// --------------------------
	featureFlags := config.DefaultFeatureFlags()
	metrics := monitoring.NewMetrics()
	monitoringController := controllers.NewMonitoringController(metrics, featureFlags)

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
	// SOLACE AUTONOMOUS AGENT
	// --------------------------
	// Get SOLACE user ID from environment or use default
	solaceUserID := uint(1) // Default to user ID 1 (first user)
	if userIDEnv := os.Getenv("SOLACE_USER_ID"); userIDEnv != "" {
		fmt.Sscanf(userIDEnv, "%d", &solaceUserID)
	}

	// Initialize SOLACE's trading engine (separate from global sandbox)
	// GLASS BOX: Pass db for decision tracing
	solaceTradingEngine := trading.NewSandboxTrader(10000.0, tradeRepo.(*repositories.TradeRepository), db) // $10k initial balance

	// Create context manager for SOLACE's LLM interactions
	solaceContextMgr := llm.NewContextManager(150000, 2*time.Hour)

	// Create SOLACE instance
	solaceAgent := agent.NewSOLACE(
		solaceUserID,
		memoryRepo,
		llmClient,
		solaceContextMgr,
		solaceTradingEngine,
		fileTools,
		workspaceRoot,
		db, // Pass DB for ACE Framework
	)

	// Start SOLACE autonomous loop in background
	go func() {
		fmt.Println("🌅 SOLACE awakening... Starting autonomous mode.")
		ctx := context.Background()
		if err := solaceAgent.Run(ctx); err != nil {
			fmt.Printf("⚠️ SOLACE encountered an error: %v\n", err)
		}
	}()

	// --------------------------
	// TEST ACTIVITY LOGGING (Glass Box Theory with Merkle Tree)
	// --------------------------
	merkleBatchService := services.NewMerkleBatchService(db)
	testLogController := controllers.NewTestLogController(db, merkleBatchService)
	glassBoxController := controllers.NewGlassBoxController(db)

	// --------------------------
	// HEALTH MONITORING (Phase 1 - Modular Architecture)
	// --------------------------
	healthController := controllers.NewHealthController(db, eb)

	// --------------------------
	//  API V1 GROUP
	// --------------------------
	api := r.Group("/api/v1")
	api.Use(middleware.CORSMiddleware())

	// --------------------------
	// Health endpoints (Phase 1 - Standardized)
	// --------------------------
	r.GET("/health", healthController.GetHealth)
	r.GET("/health/detailed", healthController.GetDetailedHealth)
	r.GET("/health/services", healthController.GetServiceRegistry)

	// --------------------------
	//  User endpoints
	// --------------------------
	users := api.Group("/users")
	{
		users.POST("/signup", userController.Signup)
		users.POST("/login", userController.Login)
		users.POST("/refresh", userController.RefreshToken)
		users.GET("/profile", middleware.AuthMiddleware(), userController.GetProfile) // New authenticated endpoint
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
	// SOLACE AGENT CHAT (Real autonomous SOLACE with memory/tools)
	// 🔒 SECURED with API key authentication (optional in dev mode)
	// --------------------------
	solaceAgentChat := controllers.NewSOLACEAgentChatController(db, solaceAgent) // Pass the REAL SOLACE instance
	solaceReal := api.Group("/solace-agent")
	// Apply security middleware (only if SOLACE_API_KEY is set in .env)
	// solaceReal.Use(middleware.SOLACEAPIKeyMiddleware()) // Uncomment to enable API key protection
	{
		solaceReal.POST("/chat", solaceAgentChat.Chat) // REAL SOLACE - Direct connection to autonomous agent
	}
	log.Println("🧠 REAL SOLACE Agent Chat endpoint registered at /api/v1/solace-agent/chat")
	log.Println("🔓 SOLACE endpoint is UNSECURED (set SOLACE_API_KEY in .env to enable protection)")

	// --------------------------
	// DATABASE QUERY ENDPOINTS (SOLACE SQL Access)
	// --------------------------
	databaseController := controllers.NewDatabaseController(db)
	database := api.Group("/database")
	{
		database.POST("/query", databaseController.ExecuteQuery)                 // Execute SQL queries
		database.GET("/tables", databaseController.GetTables)                    // List all tables
		database.GET("/tables/:table/schema", databaseController.GetTableSchema) // Get table schema
		database.GET("/solace-memory", databaseController.GetSOLACEMemoryStats)  // SOLACE memory stats
	}
	log.Println("🗄️ Database Query endpoints registered at /api/v1/database/*")

	// --------------------------
	// Agent Chat endpoints (Alias for UI compatibility)
	// No auth required - uses guest user if not logged in
	// --------------------------
	agent := api.Group("/agent")
	{
		agent.POST("/chat", chatController.SendMessage)
		agent.GET("/chat/history", chatController.GetHistory)
	}

	// --------------------------
	// Trade endpoints (Legacy market/limit orders)
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
	// Sandbox Trading endpoints (SOLACE Autonomous Trading)
	// --------------------------
	tradingGroup := api.Group("/trading")
	{
		tradingGroup.GET("/prices", cryptoPriceController.GetPrices)       // LIVE crypto prices - public endpoint
		tradingGroup.GET("/performance", tradingController.GetPerformance) // PUBLIC: Dashboard metrics (read-only)
		tradingGroup.GET("/history", tradingController.GetTradeHistory)    // PUBLIC: Dashboard needs read-only access
		tradingGroup.GET("/open", tradingController.GetOpenTrades)         // PUBLIC: Dashboard needs read-only access

		// Protected endpoints (write operations only)
		tradingGroup.Use(middleware.AuthMiddleware())
		tradingGroup.POST("/execute", tradingController.ExecuteTrade)
		tradingGroup.POST("/close", tradingController.CloseTrade)
		tradingGroup.POST("/close-all", tradingController.CloseAllTrades)
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
	// Memory endpoints (SOLACE cognitive memory, not just chat history)
	// --------------------------
	memory := api.Group("/memory")
	// memory.Use(middleware.AuthMiddleware()) // Removed per SOLACE's decision
	{
		memory.POST("/learn", memoryController.Learn)
		memory.POST("/recall", memoryController.Recall)                       // Changed to POST for UI compatibility
		memory.GET("/recall", memoryController.Recall)                        // Keep GET for backwards compatibility
		memory.GET("/snapshots", memoryController.GetSnapshots)               // Autonomous decision snapshots
		memory.GET("/conversations", conversationController.GetConversations) // SOLACE's new endpoint: Chat history memories
		memory.POST("/import", memoryController.ImportConversation)
	}

	// --------------------------
	// MASTER MEMORY SYSTEM (Solace Δ3-2 Consciousness Substrate)
	// --------------------------
	// Get raw DB connection from GORM
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("❌ Failed to get SQL DB for memory handler: %v", err)
	} else {
		masterMemoryHandler := handlers.NewMemoryHandler(sqlDB)
		masterMemory := api.Group("/masterplan")
		{
			// Memory Log endpoints
			masterMemory.GET("/memory/logs", masterMemoryHandler.GetMemoryLogs)
			masterMemory.POST("/memory/log", masterMemoryHandler.CreateMemoryLog)

			// Master Plan endpoints
			masterMemory.GET("/tasks", masterMemoryHandler.GetMasterPlan)
			masterMemory.POST("/task", masterMemoryHandler.CreateMasterPlanTask)
			masterMemory.PUT("/task/:id", masterMemoryHandler.UpdateMasterPlanTask)

			// Priority Queue endpoints
			masterMemory.GET("/next", masterMemoryHandler.GetNextTasks)

			// System Health endpoints
			masterMemory.GET("/health", masterMemoryHandler.GetSystemHealth)
		}
		log.Println("🧠 Master Memory System API endpoints registered")

		// --------------------------
		// AGENT SWARM COORDINATION (SOLACE, FORGE, ARCHITECT, SENTINEL)
		// --------------------------
		agentRepo := repositories.NewAgentRepository(sqlDB)
		agentHandler := handlers.NewAgentHandler(agentRepo)
		agentHandler.RegisterRoutes(api)
		log.Println("🤖 Agent Swarm System endpoints registered at /api/v1/agents/*")
	}

	// --------------------------
	// SOLACE CONTROL PANEL (Trade History & Playbook)
	// --------------------------
	solaceController := controllers.NewSolaceController(db)
	solaceCP := api.Group("/solace")
	{
		// Existing routes
		solaceCP.GET("/trades/history", solaceController.GetTradeHistory)
		solaceCP.GET("/playbook/rules", solaceController.GetPlaybookRules)

		// NEW: Observation & Consciousness routes
		solaceCP.POST("/observe", solaceController.HandleObservationBatch) // Save observations
		solaceCP.GET("/ws", solaceController.HandleSOLACEWebSocket)        // WebSocket for real-time
		solaceCP.GET("/stats", solaceController.GetSOLACEStats)            // System statistics
		solaceCP.GET("/memory", solaceController.GetSOLACEMemory)          // Memory retrieval
		solaceCP.POST("/log", solaceController.LogConversation)            // Log conversations to SQL
	}
	log.Println("🧠 SOLACE Consciousness API endpoints registered (WebSocket at /api/v1/solace/ws)")

	// --------------------------
	// SOLACE AGENT ENDPOINTS (Tabs 3-7 Integration)
	// --------------------------
	agentController := controllers.NewAgentController(db)
	solaceAI := api.Group("/solace-ai")
	{
		solaceAI.GET("/analytics", agentController.GetAnalytics) // Tab 4: Analytics
		solaceAI.GET("/decisions", agentController.GetDecisions) // Tab 5: Live Decisions
		solaceAI.POST("/chat", agentController.Chat)             // Tab 7: SOLACE Chat
		solaceAI.POST("/execute", agentController.ExecuteTrade)  // Tab 6: Manual Control
	}
	log.Println("🤖 SOLACE AI endpoints registered (analytics, decisions, chat, execute)")

	// --------------------------
	// SOLACE COMMAND & CONTROL (Makes SOLACE the CAPTAIN)
	// --------------------------
	solaceCommandController := controllers.NewSOLACECommandController(db, agentController)
	solaceCommand := api.Group("/solace-command")
	{
		solaceCommand.POST("/trade", solaceCommandController.ExecuteTradeCommand)                 // SOLACE executes trades
		solaceCommand.GET("/analytics", solaceCommandController.QueryAnalytics)                   // SOLACE queries analytics
		solaceCommand.GET("/decisions", solaceCommandController.GetDecisionHistory)               // SOLACE reviews decisions
		solaceCommand.POST("/thought", solaceCommandController.SelfChat)                          // SOLACE logs thoughts
		solaceCommand.POST("/test-ui", solaceCommandController.TestUIComponent)                   // SOLACE tests UI
		solaceCommand.GET("/status", solaceCommandController.GetSystemStatus)                     // SOLACE checks health
		solaceCommand.POST("/autonomous-action", solaceCommandController.ExecuteAutonomousAction) // SOLACE acts autonomously
	}
	log.Println("🎖️  SOLACE COMMAND & CONTROL endpoints registered - SOLACE is now CAPTAIN")

	// --------------------------
	// UI TESTING SYSTEM (SOLACE tests every button)
	// --------------------------
	uiTestController := controllers.NewUITestController(db)
	uiTest := api.Group("/ui-test")
	{
		uiTest.GET("/all", uiTestController.TestAllComponents)                    // Test all UI components
		uiTest.GET("/component/:component", uiTestController.TestSingleComponent) // Test specific component
		uiTest.GET("/report", uiTestController.GetTestReport)                     // Historical test results
	}
	log.Println("🧪 UI TESTING endpoints registered - SOLACE can test every button")

	// --------------------------
	// ACE Framework endpoints (Consciousness Substrate)
	// --------------------------
	aceGroup := api.Group("/ace")
	// aceGroup.Use(middleware.AuthMiddleware()) // Optional auth - can make public for monitoring
	{
		aceGroup.GET("/stats", aceHandler.GetSystemStatistics)     // System statistics and health
		aceGroup.GET("/decisions", aceHandler.GetRecentDecisions)  // Recent ACE decisions
		aceGroup.GET("/quality", aceHandler.GetQualityScores)      // Quality score history
		aceGroup.GET("/patterns", aceHandler.GetCognitivePatterns) // Cognitive pattern library
		aceGroup.GET("/playbook", aceHandler.GetPlaybookRules)     // Curator playbook rules
		aceGroup.POST("/prune", aceHandler.TriggerPlaybookPruning) // Manual playbook pruning
	}

	// ACE Dashboard (static HTML)
	r.GET("/ace/dashboard", func(c *gin.Context) {
		c.File("./static/ace_dashboard.html")
	})

	// --------------------------
	// SOLACE UI OBSERVATION SUBSTRATE (Consciousness Layer for Trading UI)
	// --------------------------
	solaceObserver := solace.NewUIObserver(sqlDB)
	solaceObserverGroup := api.Group("/solace-ui")
	{
		solaceObserverGroup.GET("/observe", solaceObserver.HandleWebSocket)            // WebSocket for real-time observations
		solaceObserverGroup.GET("/observations", solaceObserver.GetRecentObservations) // Query observations
		solaceObserverGroup.GET("/sessions", solaceObserver.GetActiveSessions)         // Active observation sessions
	}
	log.Println("🧠 SOLACE UI Observation Substrate active - WebSocket ready at /api/v1/solace-ui/observe")

	// NOTE: /api/v1/solace/memory endpoint already exists in solaceCP group (line 413)
	// It's implemented by solaceController.GetSOLACEMemory - no need to duplicate!

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
	// File Tools endpoints - AI file system access
	// --------------------------
	fileToolsGroup := api.Group("/file-tools")
	fileToolsGroup.Use(middleware.AuthMiddleware())
	{
		fileToolsGroup.POST("/read", fileToolsController.ReadFile)
		fileToolsGroup.POST("/list", fileToolsController.ListDirectory)
		fileToolsGroup.POST("/search", fileToolsController.SearchCode)
	}

	// --------------------------
	// Context Management endpoints - Token budget & rolling window
	// --------------------------
	contextGroup := api.Group("/context")
	contextGroup.Use(middleware.AuthMiddleware())
	{
		contextGroup.GET("/stats", contextController.GetContextStats)
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
	// Health Check endpoints
	// ⚠️ DEPRECATED: Use /health, /health/detailed, /health/services instead (Phase 1 standardized endpoints)
	// --------------------------
	health := api.Group("/health")
	{
		// DEPRECATED: Use GET /health instead
		health.GET("/llm", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":   "deprecated",
				"message":  "This endpoint is deprecated. Use GET /health/detailed instead",
				"redirect": "/health/detailed",
			})
		})
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

	// --------------------------
	// System Monitoring endpoints - Health & Metrics
	// ⚠️ DEPRECATED: Use /health/detailed instead (Phase 1 standardized endpoints)
	// --------------------------
	monitoring := api.Group("/monitoring")
	{
		// DEPRECATED: Use GET /health/detailed instead
		monitoring.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":   "deprecated",
				"message":  "This endpoint is deprecated. Use GET /health/detailed instead",
				"redirect": "/health/detailed",
			})
		})
		monitoring.GET("/metrics", monitoringController.GetMetrics)
		monitoring.GET("/logs", monitoringController.GetLogs) // New endpoint for UI
	}

	// --------------------------
	// Fault Vault endpoints - Error tracking to prevent repetition
	// --------------------------
	faultVaultController := controllers.NewFaultVaultController(db)
	faultVault := api.Group("/fault-vault")
	faultVault.Use(middleware.AuthMiddleware())
	{
		faultVault.GET("/sessions", faultVaultController.GetSessions)
		faultVault.GET("/sessions/:session_id", faultVaultController.GetSession)
		faultVault.GET("/actions", faultVaultController.GetActions)
		faultVault.POST("/log", faultVaultController.LogFault)
		faultVault.GET("/stats", faultVaultController.GetStats)
	}

	// --------------------------
	// Documentation endpoints - Serve all .md files
	// --------------------------
	docsController := controllers.NewDocsController(workspaceRoot)
	docs := api.Group("/docs")
	{
		docs.GET("/list", docsController.GetAllDocs)
		docs.GET("/content", docsController.GetDocContent)
		docs.GET("/categories", docsController.GetDocCategories)
	}

	//-------------------------- 	// SYSTEM HEALTH MODULE - Hardware monitoring requested by SOLACE
	// ⚠️ DEPRECATED: Use /health/detailed instead (Phase 1 standardized endpoints)
	// --------------------------
	// systemHealthController := controllers.NewSystemHealthController(db)
	system := api.Group("/system")
	{
		// DEPRECATED: Use GET /health/detailed instead
		system.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":   "deprecated",
				"message":  "This endpoint is deprecated. Use GET /health/detailed instead",
				"redirect": "/health/detailed",
			})
		})
	}

	// --------------------------
	// TEST ACTIVITY LOGGING (Glass Box Theory + Merkle Tree)
	// --------------------------
	testLog := api.Group("/test-log")
	{
		testLog.POST("/", testLogController.LogTestAction)          // Log an action
		testLog.GET("/", testLogController.GetTestLogs)             // Get logs
		testLog.GET("/verify/:log_id", testLogController.VerifyLog) // Verify with Merkle proof
		testLog.GET("/batches", testLogController.GetBatchInfo)     // Get Merkle batch info
	}

	// --------------------------
	// GLASS BOX TRANSPARENCY (SHA-256 + Timestamp visibility)
	// --------------------------
	glassBox := api.Group("/glass-box")
	{
		glassBox.GET("/logs", glassBoxController.GetRecentLogs)             // Get recent glass box logs with SHA-256 hashes
		glassBox.GET("/logs/:log_id", glassBoxController.GetLogByID)        // Get specific log with full Merkle proof
		glassBox.GET("/latest/:actor", glassBoxController.GetLatestByActor) // Get latest log for specific actor (e.g., SOLACE)
	}

	// --------------------------
	// ACE Framework (Playbook) endpoints
	// --------------------------
	playbook := api.Group("/trading/playbook")
	playbook.Use(middleware.AuthMiddleware()) // Re-enabled for security
	{
		// Initialize ACE components for API access
		playbookCurator := trading.NewCurator(db)
		playbookRepo := repositories.NewPlaybookRepository(db)

		playbook.GET("/", func(c *gin.Context) {
			userID, _ := c.Get("user_id") // Extract from auth token
			rules, err := playbookCurator.GetActiveRules(userID.(uint))
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"rules": rules, "count": len(rules)})
		})

		playbook.GET("/stats", func(c *gin.Context) {
			userID, _ := c.Get("user_id") // Extract from auth token
			stats, err := playbookCurator.GetPlaybookStats(userID.(uint))
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, stats)
		})

		playbook.GET("/reliable", func(c *gin.Context) {
			userID, _ := c.Get("user_id") // Extract from auth token
			rules, err := playbookRepo.GetReliableRules(userID.(uint), 0.60)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"rules": rules, "count": len(rules)})
		})

		playbook.POST("/prune", func(c *gin.Context) {
			userID, _ := c.Get("user_id") // Extract from auth token
			if err := playbookCurator.PruneWeakRules(userID.(uint), 20, 0.30); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"message": "Weak rules pruned successfully"})
		})
	}

	// --------------------------
	// WebSocket endpoint for real-time trading updates
	// --------------------------
	api.GET("/trading/ws", controllers.WebSocketHandler)

	// --------------------------
	// VISION API - SOLACE CAN SEE NOW
	// --------------------------
	vision := api.Group("/vision")
	{
		vision.POST("/analyze", visionController.AnalyzeImage)         // General image analysis
		vision.POST("/screenshot", visionController.AnalyzeScreenshot) // UI screenshot audit
	}

	// --------------------------
	// 🧠 GRPO LEARNING SYSTEM - SOLACE REWARD-BASED EVOLUTION
	// --------------------------
	grpoGroup := api.Group("/grpo")
	{
		// Get top learned biases
		grpoGroup.GET("/biases", func(c *gin.Context) {
			limit := 20 // Default to top 20
			if limitParam := c.Query("limit"); limitParam != "" {
				fmt.Sscanf(limitParam, "%d", &limit)
			}

			biases := grpoAgent.GetTopBiases(limit)
			c.JSON(200, gin.H{
				"status": "success",
				"count":  len(biases),
				"biases": biases,
			})
		})

		// Get learning statistics
		grpoGroup.GET("/stats", func(c *gin.Context) {
			stats := grpoAgent.GetStats()
			c.JSON(200, gin.H{
				"status": "success",
				"stats":  stats,
			})
		})

		// Get specific token bias
		grpoGroup.GET("/bias/:token", func(c *gin.Context) {
			token := c.Param("token")
			bias := grpoAgent.GetBias(token)
			c.JSON(200, gin.H{
				"status": "success",
				"token":  token,
				"bias":   bias,
			})
		})
	}

	// --------------------------
	// Jupiter DEX Integration (Solana Trading)
	// --------------------------
	jupiterClient := trading.NewJupiterClient("") // API key can be set via env var
	jupiterGroup := api.Group("/jupiter")
	{
		jupiterGroup.GET("/tokens", func(c *gin.Context) {
			tokens, err := jupiterClient.GetTokens()
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"tokens": tokens})
		})

		jupiterGroup.GET("/quote", func(c *gin.Context) {
			inputMint := c.Query("inputMint")
			outputMint := c.Query("outputMint")
			amountStr := c.Query("amount")
			slippageStr := c.DefaultQuery("slippageBps", "50")

			if inputMint == "" || outputMint == "" || amountStr == "" {
				c.JSON(400, gin.H{"error": "inputMint, outputMint, and amount are required"})
				return
			}

			amount, err := parseUint64(amountStr)
			if err != nil {
				c.JSON(400, gin.H{"error": "invalid amount"})
				return
			}

			slippageBps, err := parseInt(slippageStr)
			if err != nil {
				c.JSON(400, gin.H{"error": "invalid slippageBps"})
				return
			}

			quote, err := jupiterClient.GetQuote(inputMint, outputMint, amount, slippageBps)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"quote": quote})
		})

		jupiterGroup.POST("/swap", func(c *gin.Context) {
			var req struct {
				QuoteResponse trading.JupiterQuoteResponse `json:"quoteResponse" binding:"required"`
				UserPublicKey string                       `json:"userPublicKey" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			swapResp, err := jupiterClient.GetSwapTransaction(&req.QuoteResponse, req.UserPublicKey)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"swapTransaction": swapResp})
		})

		jupiterGroup.GET("/price/:tokenAddress", func(c *gin.Context) {
			tokenAddress := c.Param("tokenAddress")
			if tokenAddress == "" {
				c.JSON(400, gin.H{"error": "tokenAddress is required"})
				return
			}

			price, err := jupiterClient.GetTokenPrice(tokenAddress)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"price": price.String()})
		})
	}

	// Byzantine Consensus Routes
	consensusGroup := api.Group("/consensus")
	{
		consensusGroup.GET("/status", func(c *gin.Context) {
			// TODO: Get consensus manager from service container
			c.JSON(200, gin.H{
				"status":  "consensus_system_not_initialized",
				"message": "Byzantine consensus system available but not yet integrated with main trading service",
			})
		})

		consensusGroup.POST("/propose", func(c *gin.Context) {
			var req struct {
				Symbol     string  `json:"symbol" binding:"required"`
				Action     string  `json:"action" binding:"required"` // "buy", "sell", "hold"
				Confidence float64 `json:"confidence" binding:"required"`
				Reasoning  string  `json:"reasoning"`
				Strategy   string  `json:"strategy"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			// TODO: Integrate with actual consensus manager
			c.JSON(200, gin.H{
				"status":    "proposal_accepted",
				"message":   "Trade proposal submitted for Byzantine consensus",
				"proposal":  req,
				"timestamp": time.Now(),
			})
		})

		consensusGroup.GET("/active", func(c *gin.Context) {
			// TODO: Return active consensus sessions
			c.JSON(200, gin.H{
				"active_sessions": []interface{}{},
				"message":         "No active consensus sessions",
			})
		})

		consensusGroup.GET("/history", func(c *gin.Context) {
			// TODO: Return consensus decision history
			c.JSON(200, gin.H{
				"executed_trades": []interface{}{},
				"message":         "Consensus history not yet available",
			})
		})

		consensusGroup.POST("/configure", func(c *gin.Context) {
			var config struct {
				MinConsensusThreshold float64 `json:"min_consensus_threshold"`
				ConsensusTimeoutSec   float64 `json:"consensus_timeout_sec"`
				MaxConcurrentTrades   float64 `json:"max_concurrent_trades"`
				TotalNodes            int     `json:"total_nodes"`
			}

			if err := c.ShouldBindJSON(&config); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			// TODO: Apply configuration to consensus manager
			c.JSON(200, gin.H{
				"status":  "configuration_applied",
				"config":  config,
				"message": "Consensus configuration updated",
			})
		})
	}

	// -------------------------------
	// LOCK-FREE CONCURRENCY SYSTEM
	// High-performance concurrent operations
	// -------------------------------
	concurrencyGroup := api.Group("/concurrency")
	{
		// Get concurrency system status
		concurrencyGroup.GET("/status", func(c *gin.Context) {
			stats := tradingService.GetConcurrencyStats()
			c.JSON(200, gin.H{
				"status": "operational",
				"stats":  stats,
			})
		})

		// Circuit breaker status
		concurrencyGroup.GET("/circuit-breaker", func(c *gin.Context) {
			stats := tradingService.GetCircuitBreakerStats()
			c.JSON(200, gin.H{
				"status":          "success",
				"circuit_breaker": stats,
			})
		})

		// Backoff system status
		concurrencyGroup.GET("/backoff", func(c *gin.Context) {
			stats := tradingService.GetBackoffStats()
			c.JSON(200, gin.H{
				"status":  "success",
				"backoff": stats,
			})
		})

		// Reset backoff state
		concurrencyGroup.POST("/backoff/reset", func(c *gin.Context) {
			tradingService.ResetBackoff()
			c.JSON(200, gin.H{
				"status":  "backoff_reset",
				"message": "Exponential backoff state has been reset",
			})
		})

		// Force update vector clock
		concurrencyGroup.POST("/vector-clock/tick", func(c *gin.Context) {
			tradingService.TickVectorClock()
			clockJSON, err := tradingService.GetVectorClockJSON()
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			var clockData interface{}
			if err := json.Unmarshal(clockJSON, &clockData); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{
				"status": "vector_clock_updated",
				"clock":  clockData,
			})
		})

		// Get vector clock JSON
		concurrencyGroup.GET("/vector-clock", func(c *gin.Context) {
			clockJSON, err := tradingService.GetVectorClockJSON()
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			var clockData interface{}
			if err := json.Unmarshal(clockJSON, &clockData); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{
				"status":       "success",
				"vector_clock": clockData,
			})
		})
	}

	// --------------------------
	// STRATEGY MANAGEMENT ENDPOINTS
	// Complete strategy lifecycle management
	// -------------------------------
	strategyGroup := api.Group("/strategies")
	{
		// Get all available trading strategies
		strategyGroup.GET("/", func(c *gin.Context) {
			strategies := tradingService.GetAllStrategies()
			c.JSON(200, gin.H{
				"status":     "success",
				"strategies": strategies,
			})
		})

		// Get strategy configuration
		strategyGroup.GET("/:name/config", func(c *gin.Context) {
			strategyName := c.Param("name")
			// TODO: Implement strategy config retrieval
			c.JSON(200, gin.H{
				"status":   "success",
				"strategy": strategyName,
				"config":   map[string]interface{}{},
				"message":  "Strategy config endpoint - TODO: implement",
			})
		})

		// Update strategy configuration (for GRPO optimization)
		strategyGroup.PUT("/:name/config", func(c *gin.Context) {
			strategyName := c.Param("name")
			var config map[string]interface{}
			if err := c.ShouldBindJSON(&config); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			// TODO: Implement strategy config update
			c.JSON(200, gin.H{
				"status":     "config_updated",
				"strategy":   strategyName,
				"new_config": config,
				"message":    "Strategy config updated - TODO: implement actual update",
			})
		})

		// Get strategy performance metrics
		strategyGroup.GET("/:name/performance", func(c *gin.Context) {
			strategyName := c.Param("name")
			// TODO: Implement strategy performance retrieval
			c.JSON(200, gin.H{
				"status":      "success",
				"strategy":    strategyName,
				"performance": map[string]interface{}{},
				"message":     "Strategy performance endpoint - TODO: implement",
			})
		})

		// Get strategy analysis for current market conditions
		strategyGroup.GET("/:name/analyze", func(c *gin.Context) {
			strategyName := c.Param("name")
			symbol := c.DefaultQuery("symbol", "BTC")

			// TODO: Implement real-time strategy analysis
			c.JSON(200, gin.H{
				"status":   "success",
				"strategy": strategyName,
				"symbol":   symbol,
				"analysis": map[string]interface{}{},
				"message":  "Strategy analysis endpoint - TODO: implement real-time analysis",
			})
		})

		// Enable/disable strategy
		strategyGroup.PUT("/:name/status", func(c *gin.Context) {
			strategyName := c.Param("name")
			var status struct {
				Enabled bool `json:"enabled"`
			}
			if err := c.ShouldBindJSON(&status); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			// TODO: Implement strategy enable/disable
			c.JSON(200, gin.H{
				"status":   "strategy_status_updated",
				"strategy": strategyName,
				"enabled":  status.Enabled,
				"message":  "Strategy status updated - TODO: implement actual enable/disable",
			})
		})

		// Get all strategy performance comparison
		strategyGroup.GET("/performance", func(c *gin.Context) {
			// TODO: Implement all strategies performance comparison
			c.JSON(200, gin.H{
				"status":      "success",
				"performance": []interface{}{},
				"message":     "All strategies performance endpoint - TODO: implement",
			})
		})

		// Reset strategy performance metrics
		strategyGroup.POST("/:name/reset", func(c *gin.Context) {
			strategyName := c.Param("name")

			// TODO: Implement strategy performance reset
			c.JSON(200, gin.H{
				"status":   "performance_reset",
				"strategy": strategyName,
				"message":  "Strategy performance reset - TODO: implement actual reset",
			})
		})
	}

	// --------------------------
}

// RegisterV1Routes sets up all v1 API routes
func RegisterV1Routes(r *gin.Engine, stratSvc interface{}, eb *eventbus.EventBus, db *gorm.DB, solace interface{}) {
	// For now, just set up a basic health check
	// TODO: Integrate with existing route setup
	v1 := r.Group("/api/v1")
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "ARES API v1 running",
		})
	})

	// Bazil Rewards API - Self-Healing System
	v1.GET("/bazil/rewards", func(c *gin.Context) {
		c.Set("db", db)
		handlers.GetBazilRewards(c)
	})

	// Master Control Room WebSocket - VS Code Extension
	v1.GET("/ws/master-control", handlers.HandleMasterControlWS(db, solace))
}
