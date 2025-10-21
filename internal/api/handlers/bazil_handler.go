/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package handlers

import (
	"ares_api/internal/agent"
	"ares_api/internal/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BazilHandler handles Bazil sniffer API requests
type BazilHandler struct {
	bazil *agent.BazilSniffer
}

// NewBazilHandler creates a new Bazil handler
func NewBazilHandler(db *gorm.DB) *BazilHandler {
	return &BazilHandler{
		bazil: agent.NewBazilSniffer(db),
	}
}

// GetBazilRewards returns all Bazil reward points
func GetBazilRewards(c *gin.Context) {
	// Get DB from context (assumes it's set by middleware)
	db, exists := c.Get("db")
	if !exists {
		// Fallback: Try to get from request context or use global
		log.Println("⚠️ DB not found in gin context for Bazil rewards")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
		return
	}

	database, ok := db.(*gorm.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection"})
		return
	}

	var rewards []models.BazilReward
	if err := database.Find(&rewards).Error; err != nil {
		log.Printf("❌ Failed to fetch Bazil rewards: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rewards"})
		return
	}

	// Convert to map format
	rewardMap := make(map[string]int)
	for _, r := range rewards {
		rewardMap[r.FaultType] = r.Points
	}

	c.JSON(http.StatusOK, rewardMap)
}
