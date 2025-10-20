package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ares_api/internal/agent"
	"ares_api/internal/api/routes"
	"ares_api/internal/config"
	"ares_api/internal/eventbus"
	Repositories "ares_api/internal/interfaces/repository"
	"ares_api/internal/middleware"
	"ares_api/internal/observability"
	"ares_api/internal/services"
	"ares_api/internal/trading"
	"ares_api/pkg/llm"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load config from .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config load failed: ", err)
	}

	// Connect to PostgreSQL with advanced options (pooling, timeouts)
	dsn := cfg.DBDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatal("DB connection failed: ", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Run migrations
	// Temporarily disabled due to constraint issues
	// if err := database.Migrate(db); err != nil {
	//	log.Fatal("Migration failed: ", err)
	// }

	// Initialize observability (OpenTelemetry)
	otelShutdown, err := observability.SetupOTelSDK(context.Background())
	if err != nil {
		log.Fatal("OTel setup failed: ", err)
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	// Initialize EventBus (in-memory with Redis fallback)
	eb := eventbus.NewEventBus() // Empty for in-memory

	// Initialize services (inject dependencies)
	histMgr := services.NewHistoricalDataManager(cfg)
	backtester := services.NewBacktester()
	versionMgr := services.NewStrategyVersionManager(db)
	autoGrad := services.NewAutoGraduateMonitor(db, eb)
	stratSvc := services.NewStrategyService(db, trading.NewMultiStrategyOrchestrator(db, eb, histMgr), backtester, versionMgr, autoGrad, eb, histMgr, nil)

	// Initialize SOLACE agent for Master Control Room
	var memoryRepo Repositories.MemoryRepository                // TODO: Initialize proper memory repo
	llmClient := llm.NewClient()                                // Initialize LLM client
	contextMgr := llm.NewContextManager(100000, 1*time.Hour)    // 100k token context, 1 hour window
	tradingEngine := trading.NewSandboxTrader(10000.0, nil, db) // $10k initial balance, nil repo for now
	fileTools := llm.NewFileAccessTools("/workspace")           // File operations
	workspaceRoot := "C:\\ARES_Workspace\\ARES_API"

	solaceAgent := agent.NewSOLACE(
		1, // User ID
		memoryRepo,
		llmClient,
		contextMgr,
		tradingEngine,
		fileTools,
		workspaceRoot,
		db,
	)

	// Setup Gin router with production middleware
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())
	r.Use(middleware.AuthMiddleware())
	r.Use(middleware.RateLimiter(100, time.Minute)) // Global rate limit

	// Register routes (including strategies)
	routes.RegisterV1Routes(r, stratSvc, eb, db, solaceAgent) // Wire all endpoints

	// Setup WebSocket hub for real-time
	go services.RunWebSocketHub()

	// Background schedulers/goroutines
	ctx, cancel := context.WithCancel(context.Background())
	go services.RunMemoryConsolidation(ctx, db, 24*time.Hour)
	go services.RunOpenOrdersProcessor(ctx, db, 10*time.Second)
	go services.RunEmbeddingsQueue(ctx, db, 30*time.Second)
	go services.RunStrategyAutoPromotion(ctx, db, eb, 5*time.Minute)

	// Graceful shutdown server
	srv := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown context
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
	cancel() // Stop schedulers
	log.Println("Server exiting")
}
