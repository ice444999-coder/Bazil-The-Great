/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package handlers

import (
	"ares_api/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConfigHandler handles configuration endpoints
type ConfigHandler struct {
	db      *gorm.DB
	manager *config.Manager
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(db *gorm.DB, manager *config.Manager) *ConfigHandler {
	return &ConfigHandler{
		db:      db,
		manager: manager,
	}
}

// GetConfig retrieves a specific config value
// GET /api/v1/config/:service/:key
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	serviceName := c.Param("service")
	configKey := c.Param("key")

	var cfg config.ServiceConfig
	err := h.db.Where("service_name = ? AND config_key = ?", serviceName, configKey).First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cfg)
}

// GetAllConfigs retrieves all configs for a service
// GET /api/v1/config/:service
func (h *ConfigHandler) GetAllConfigs(c *gin.Context) {
	serviceName := c.Param("service")

	var configs []config.ServiceConfig
	err := h.db.Where("service_name = ?", serviceName).Order("config_key").Find(&configs).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"count":   len(configs),
		"configs": configs,
	})
}

// SetConfig creates or updates a config value
// PUT /api/v1/config/:service/:key
func (h *ConfigHandler) SetConfig(c *gin.Context) {
	serviceName := c.Param("service")
	configKey := c.Param("key")

	var req struct {
		Value       interface{} `json:"value" binding:"required"`
		Description string      `json:"description"`
		UpdatedBy   string      `json:"updated_by"`
		Reason      string      `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use manager to set config (with history tracking)
	if h.manager != nil && h.manager.GetServiceName() == serviceName {
		err := h.manager.Set(configKey, req.Value, req.UpdatedBy, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Config updated successfully",
		"service": serviceName,
		"key":     configKey,
	})
}

// GetConfigHistory retrieves change history for a config
// GET /api/v1/config/:service/:key/history
func (h *ConfigHandler) GetConfigHistory(c *gin.Context) {
	serviceName := c.Param("service")
	configKey := c.Param("key")
	limit := 50

	var history []config.ConfigHistory
	err := h.db.Where("service_name = ? AND config_key = ?", serviceName, configKey).
		Order("changed_at DESC").
		Limit(limit).
		Find(&history).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"key":     configKey,
		"history": history,
	})
}

// DeleteConfig deletes a config entry
// DELETE /api/v1/config/:service/:key
func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	serviceName := c.Param("service")
	configKey := c.Param("key")

	var cfg config.ServiceConfig
	err := h.db.Where("service_name = ? AND config_key = ?", serviceName, configKey).First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Delete(&cfg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config deleted successfully"})
}

// RegisterRoutes registers config routes
func (h *ConfigHandler) RegisterRoutes(r *gin.RouterGroup) {
	config := r.Group("/config")
	{
		config.GET("/:service", h.GetAllConfigs)
		config.GET("/:service/:key", h.GetConfig)
		config.PUT("/:service/:key", h.SetConfig)
		config.DELETE("/:service/:key", h.DeleteConfig)
		config.GET("/:service/:key/history", h.GetConfigHistory)
	}
}
