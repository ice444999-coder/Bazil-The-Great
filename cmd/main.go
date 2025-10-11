package main

import (
	"log"
	"os"
	"os/exec"

	"ares_api/config"
	"ares_api/internal/api/routes"
	"ares_api/internal/common"
	"ares_api/internal/database"
	"ares_api/internal/middleware"

	"ares_api/internal/docs"
	"github.com/gin-gonic/gin"
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
	// Load server port from env, fallback to 8080
	port := config.GetEnv("SERVER_PORT", "8080")

	// Setup logger
	common.SetupLogger()

	// Generate Swagger docs
	generateSwagger()

	// Initialize PostgreSQL DB
	db := database.InitDB()

	// Setup Gin
	gin.SetMode("debug") // or "release"
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

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

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Static files for Monaco Editor
	r.Static("/static", "./static")

	// Register API routes with DB dependency
	routes.RegisterRoutes(r, db)

	// Start server
	addr := ":" + port
	log.Printf("🚀 Server running at http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
