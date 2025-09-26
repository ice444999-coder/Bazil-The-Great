package routes

import (
	controllers "ares_api/internal/api/controllers"
	"ares_api/internal/middleware"
	"ares_api/internal/ollama"
	repositories "ares_api/internal/repositories"
	service "ares_api/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes sets up all API routes with their dependencies
func RegisterRoutes(r *gin.Engine, db *gorm.DB) {

	// --------------------------
	// User Module
	// --------------------------
	userRepo := repositories.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	// --------------------------
	// Chat Module
	// --------------------------
	ollamaService := ollama.NewClientFromEnv()
	chatRepo := repositories.NewChatRepository(db)
	chatService := service.NewChatService(chatRepo , ollamaService) 
	chatController := controllers.NewChatController(chatService)

	// --------------------------
	// API versioning
	// --------------------------
	api := r.Group("/api/v1")
	api.Use(middleware.CORSMiddleware())

	// --------------------------
	// User endpoints
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
}
