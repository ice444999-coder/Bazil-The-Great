package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"ares_api/config"
	"ares_api/internal/api/controllers"
	"ares_api/internal/api/handlers"
	"ares_api/internal/api/routes"
	"ares_api/internal/common"
	appConfig "ares_api/internal/config"
	"ares_api/internal/database"
	"ares_api/internal/eventbus"
	"ares_api/internal/grpo"
	"ares_api/internal/logger"
	"ares_api/internal/middleware"
	"ares_api/internal/observability"
	"ares_api/internal/registry"
	"ares_api/internal/services"
	"ares_api/internal/subscribers"

	"ares_api/internal/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// generateSwagger runs `swag init` to generate docs
func generateSwagger() {
	log.Println("Generating Swagger docs...")
	cmd := exec.Command("swag", "init",
		"--dir", "./cmd,./internal/api/controllers,./internal/api/dto",
		"--output", "./internal/docs",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to generate Swagger docs: %v", err)
	}
	log.Println("Swagger docs generated.")
}

func main() {
	// 🔧 Load environment variables first!
	// Try multiple paths in case executable is run from different locations
	envPaths := []string{".env", "../.env", "../../.env", "c:\\ARES_Workspace\\ARES_API\\.env"}
	loaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("✅ .env file loaded successfully from: %s", path)
			loaded = true
			break
		}
	}
	if !loaded {
		log.Println("⚠️ No .env file found in any expected location, using system environment variables")
	}

	// Load server port from env, fallback to 8080
	port := config.GetEnv("SERVER_PORT", "8080")

	// Setup logger
	common.SetupLogger()

	// Generate Swagger docs (skip if swag not installed)
	// generateSwagger()

	// Initialize PostgreSQL DB
	db := database.InitDB()

	// � Initialize Config Manager (Section 5 - Hot-reload config)
	configManager := appConfig.NewManager(db, "ares-api")
	defer configManager.Close()
	log.Println("✅ Config manager initialized with hot-reload support")

	// Get EventBus type from config (defaults to in-memory)
	eventbusType := configManager.GetString("eventbus.type", "memory")
	redisURL := ""
	if eventbusType == "redis" {
		redisURL = configManager.GetString("eventbus.redis_url", "localhost:6379")
	}

	// 🚀 Initialize EventBus (Section 4 - Enhanced with Redis support)
	ebInterface := eventbus.NewEventBusWithRedis(redisURL)
	log.Printf("✅ EventBus initialized (type: %s)", eventbusType)

	// Type assert to *EventBus for legacy code compatibility
	var eb *eventbus.EventBus
	if eventbusType == "memory" || redisURL == "" {
		eb = ebInterface.(*eventbus.EventBus)
	} else {
		// For Redis, we need to adapt - for now fallback to in-memory if needed
		if ebConcrete, ok := ebInterface.(*eventbus.EventBus); ok {
			eb = ebConcrete
		} else {
			log.Println("⚠️  Using in-memory EventBus for legacy compatibility")
			eb = eventbus.NewEventBus()
		}
	}

	// 🔍 Initialize Enhanced Observability (Section 6 - Distributed tracing + metrics)
	obsLogger := observability.NewLogger(db, "ares-api")
	metricsCollector := observability.NewMetricsCollector(db, "ares-api")
	log.Println("✅ Enhanced observability initialized (tracing + metrics)")

	// � Initialize Centralized Logger (Option B - Consolidate & Clean Up)
	centralLogger := logger.NewLogger("ARES_API", db)
	logger.SetGlobalLogger(centralLogger)
	logger.Info("Centralized logging system initialized")

	// 🔍 Initialize Audit Logger (Option A - Complete Integration)
	auditLogger := logger.NewAuditLogger(db, eb)
	auditLogger.Start()

	// � Initialize Event Subscribers (Phase 2 Integration)
	auditSubscriber := subscribers.NewTradeAuditSubscriber(db)
	auditSubscriber.Subscribe(eb)

	analyticsSubscriber := subscribers.NewAnalyticsSubscriber()
	analyticsSubscriber.Subscribe(eb)

	log.Println("✅ Event subscribers initialized (audit + analytics)")

	// 💾 Initialize Database Write Queue (Phase 3 - Graceful Degradation)
	writeQueue := database.NewWriteQueue(db, 1000)
	log.Println("✅ Write queue initialized (max: 1000, retry: 5s)")

	// Make write queue available globally for critical operations
	// Note: In production, inject this into services that need it
	_ = writeQueue // Prevent unused variable error for now

	// 🧠 Initialize GRPO Learning System (SOLACE reward-based evolution)
	grpoUpdater := grpo.NewUpdater(db, 0.01, 10) // 0.01 learning rate, 10-min intervals
	if err := grpoUpdater.Start(); err != nil {
		log.Printf("⚠️ GRPO updater failed to start: %v", err)
	} else {
		log.Println("✅ GRPO learning system initialized (lr=0.01, interval=10min)")
	}
	grpoAgent := grpoUpdater.GetAgent()

	// 🧠 Initialize SOLACE Δ3-2 Consciousness Substrate
	// DISABLED: Master Memory System deployed manually via migrations/001_master_memory_system.sql
	// if err := database.InitializeConsciousnessSubstrate(db); err != nil {
	// 	log.Printf("⚠️ Consciousness substrate initialization failed: %v", err)
	// 	log.Println("   Continuing with existing schema...")
	// }

	// 🔍 SOLACE Orchestration Services
	orchestrationService := services.NewOrchestrationService(db)
	repoInspectionService := services.NewRepoInspectionService(db, "C:\\ARES_Workspace\\ARES_API")

	// Run initial repository scan asynchronously to avoid blocking startup
	// TEMPORARILY DISABLED: causing server crash
	// go inspectRepoAsync(repoInspectionService)

	// Prevent unused variable warnings
	_ = orchestrationService
	_ = repoInspectionService

	// 📋 Register this service in the service registry (Phase 1 - Modular Architecture)
	if err := registry.RegisterService(db, "ares-api", "1.0.0", 8080, "http://localhost:8080/health"); err != nil {
		log.Printf("⚠️ Service registration failed: %v", err)
	}

	// 💓 Start heartbeat to keep service status updated
	go registry.ServiceHeartbeat(db, "ares-api", 30*time.Second)

	// Setup Gin
	gin.SetMode("debug") // or "release"
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// Add cache-busting middleware for HTML files
	r.Use(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 5 && c.Request.URL.Path[len(c.Request.URL.Path)-5:] == ".html" {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	// Swagger metadata
	docs.SwaggerInfo.Title = "ARES Platform API"
	docs.SwaggerInfo.Description = "API documentation for the ARES Platform service."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Add security definition for Bearer Auth
	// This tells Swagger that endpoints use JWT Bearer token in the header
	// @securityDefinitions.apikey BearerAuth
	// @in header
	// @name Authorization
	// @description Type "Bearer" followed by a space and your JWT token (e.g., "Bearer <token>")

	// NOTE: Health endpoint moved to routes.RegisterRoutes() for standardization (Phase 1)

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Static files for Monaco Editor
	r.Static("/static", "./static")

	// Serve HTML pages - with proper MIME types and cache control
	r.GET("/web/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		fullPath := "./web" + filepath

		// Set proper cache headers for HTML
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.File(fullPath)
	})

	// Legacy routes for backwards compatibility
	r.StaticFile("/login.html", "./web/login.html")       // CRITICAL: Explicit route for login
	r.StaticFile("/register.html", "./web/register.html") // CRITICAL: Explicit route for register
	r.StaticFile("/trading.html", "./web/trading.html")
	r.StaticFile("/dashboard.html", "./web/dashboard.html")
	r.StaticFile("/editor.html", "./static/editor.html")
	r.StaticFile("/code-ide.html", "./web/code-ide.html") // SOLACE Code IDE

	// Additional UI pages
	r.StaticFile("/chat.html", "./web/chat.html")
	r.StaticFile("/solace-control.html", "./web/solace-control.html")
	r.StaticFile("/solace-trading.html", "./web/solace-trading.html")
	r.StaticFile("/forge-dashboard.html", "./web/forge-dashboard.html")
	r.StaticFile("/memory.html", "./web/memory.html")
	r.StaticFile("/vision.html", "./web/vision.html")
	r.StaticFile("/health.html", "./web/health.html")

	// Serve React SPA from frontend/dist (for future use)
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/vite.svg", "./frontend/dist/vite.svg")

	// SPA catch-all route - serve trading by default
	r.NoRoute(func(c *gin.Context) {
		// Don't intercept API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}
		// Serve trading page as default
		c.File("./web/trading.html")
	})

	// Register API routes with DB dependency, EventBus, and GRPO Agent (Phase 2 + GRPO)
	routes.RegisterRoutes(r, db, eb, grpoAgent)

	// 🛡️ Approval Controller (Grok Protocol Safety Gates)
	approvalController := controllers.NewApprovalController(db)
	approvalGroup := r.Group("/api/approve")
	{
		approvalGroup.POST("/request", approvalController.RequestApproval)
		approvalGroup.POST("/:subtask_id", approvalController.ApproveSubtask)
		approvalGroup.POST("/:subtask_id/reject", approvalController.RejectSubtask)
		approvalGroup.GET("/:subtask_id", approvalController.GetApprovalStatus)
		approvalGroup.GET("/pending", approvalController.ListPendingApprovals)
		approvalGroup.POST("/all", approvalController.ApproveAll)
	}
	log.Println("✅ Approval controller registered for Grok protocol safety gates")

	// 🚀 Mission Progress Controller (Subtask 2: Mission Progress Bar)
	missionController := controllers.NewMissionController(db)
	missionGroup := r.Group("/api/mission")
	{
		missionGroup.GET("/progress", missionController.GetProgress)
		missionGroup.POST("/progress", missionController.UpdateProgress)
		missionGroup.POST("/progress/increment", missionController.IncrementProgress)
	}
	log.Println("✅ Mission controller registered for Phase 1 progress tracking")

	// �📊 Add analytics endpoint (Phase 2 Integration)
	r.GET("/api/v1/analytics/trading", func(c *gin.Context) {
		stats := analyticsSubscriber.GetStats()
		c.JSON(200, gin.H{
			"status":    "success",
			"analytics": stats,
		})
	})

	// 🔄 Initialize WebSocket Hub for real-time dashboard updates (Task #9)
	wsHub := handlers.NewWebSocketHub(eb)
	go wsHub.Run()
	log.Println("✅ WebSocket hub started for real-time updates")

	// WebSocket endpoint for dashboard
	r.GET("/ws", func(c *gin.Context) {
		wsHub.HandleWebSocket(c.Writer, c.Request)
	})

	// Serve real-time dashboard
	r.StaticFile("/dashboard_realtime.html", "./web/dashboard_realtime.html")

	// 🎛️ Register Modular Architecture endpoints (Sections 3-6)
	configHandler := handlers.NewConfigHandler(db, configManager)
	configHandler.RegisterRoutes(r.Group("/api/v1"))

	observabilityHandler := handlers.NewObservabilityHandler(db, obsLogger)
	observabilityHandler.RegisterRoutes(r.Group("/api/v1"))

	log.Println("✅ Modular architecture endpoints registered (config + observability)")

	// Record startup metric
	metricsCollector.RecordCounter("ares.startup", 1, map[string]string{
		"version": "1.0.0",
		"mode":    gin.Mode(),
	})

	// Start server
	addr := ":" + port
	log.Printf("🚀 Server running at http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// inspectRepoAsync performs repository inspection asynchronously
func inspectRepoAsync(service *services.RepoInspectionService) {
	log.Println("🔍 Starting asynchronous repository inspection...")
	if err := service.ScanRepository(); err != nil {
		log.Printf("⚠️ Repository inspection failed: %v", err)
	} else {
		log.Println("✅ Repository inspection completed")
	}
}
