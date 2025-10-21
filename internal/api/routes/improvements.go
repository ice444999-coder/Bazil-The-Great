/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
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
