package routes

import (
	"ares_api/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up the improvement routes
func SetupRoutes(router *gin.Engine) {
	improvementGroup := router.Group("/api/v1/improvements")
	{
		improvementGroup.GET("/queue", controllers.ListImprovements)
		improvementGroup.POST("/:id/approve", controllers.ApproveImprovement)
		improvementGroup.POST("/:id/reject", controllers.RejectImprovement)
		improvementGroup.POST("/execute-all", controllers.ExecuteAllImprovements)
	}
}
