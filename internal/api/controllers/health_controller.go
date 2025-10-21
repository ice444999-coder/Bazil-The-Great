/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"net/http"
	"time"

	"ares_api/internal/eventbus"
	"ares_api/internal/registry"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthController struct {
	db        *gorm.DB
	eventBus  *eventbus.EventBus
	startTime time.Time
}

func NewHealthController(db *gorm.DB, eb *eventbus.EventBus) *HealthController {
	return &HealthController{
		db:        db,
		eventBus:  eb,
		startTime: time.Now(),
	}
}

type DetailedHealthResponse struct {
	Service       string            `json:"service"`
	Version       string            `json:"version"`
	Status        string            `json:"status"`
	UptimeSeconds int64             `json:"uptime_seconds"`
	Dependencies  map[string]string `json:"dependencies"`
	LastCheck     string            `json:"last_check"`
}

// GetHealth godoc
// @Summary Quick health check
// @Description Returns basic health status of the service
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (hc *HealthController) GetHealth(c *gin.Context) {
	// Quick health check
	status := "healthy"

	// Check database connection
	sqlDB, err := hc.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"service": "ares-api",
		"status":  status,
	})
}

// GetDetailedHealth godoc
// @Summary Detailed health check
// @Description Returns comprehensive health status including uptime and dependencies
// @Tags Health
// @Produce json
// @Success 200 {object} DetailedHealthResponse
// @Router /health/detailed [get]
func (hc *HealthController) GetDetailedHealth(c *gin.Context) {
	dependencies := make(map[string]string)

	// Check database
	sqlDB, err := hc.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dependencies["database"] = "unhealthy"
	} else {
		dependencies["database"] = "healthy"
	}

	// Check EventBus (Phase 2)
	if hc.eventBus != nil {
		eventBusHealth := hc.eventBus.Health()
		if status, ok := eventBusHealth["status"].(string); ok {
			dependencies["event_bus"] = status + " (in-memory)"
		}
	} else {
		dependencies["event_bus"] = "not_configured"
	}

	// Check if Hedera is configured (optional)
	dependencies["hedera"] = "not_configured"

	// Determine overall status
	overallStatus := "healthy"
	if dependencies["database"] == "unhealthy" {
		overallStatus = "unhealthy"
	}

	response := DetailedHealthResponse{
		Service:       "ares-api",
		Version:       "1.0.0",
		Status:        overallStatus,
		UptimeSeconds: int64(time.Since(hc.startTime).Seconds()),
		Dependencies:  dependencies,
		LastCheck:     time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// GetServiceRegistry godoc
// @Summary Get all registered services
// @Description Returns list of all services in the service registry
// @Tags Health
// @Produce json
// @Success 200 {array} registry.ServiceInfo
// @Router /health/services [get]
func (hc *HealthController) GetServiceRegistry(c *gin.Context) {
	services, err := registry.GetAllServices(hc.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch service registry",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}

// GetHelloWorld godoc
// @Summary Health endpoint that returns a hello world message
// @Description Returns a hello world message
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health/hello-world [get]
func (hc *HealthController) GetHelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, world!",
	})
}

// GetNewHealthCheck godoc
// @Summary New health check endpoint
// @Description Returns a custom message
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health/new-check [get]
func (hc *HealthController) GetNewHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "New health check endpoint is working correctly!",
	})
}
