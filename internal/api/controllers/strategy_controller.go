/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"net/http"

	"ares_api/internal/services"

	"github.com/gin-gonic/gin"
)

type StrategyController struct {
	service *services.StrategyService
}

func NewStrategyController(service *services.StrategyService) *StrategyController {
	return &StrategyController{service: service}
}

func (c *StrategyController) GetStrategies(ctx *gin.Context) {
	strategies, err := c.service.ListStrategies()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, strategies)
}

func (c *StrategyController) ToggleStrategy(ctx *gin.Context) {
	name := ctx.Param("name")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := c.service.ToggleStrategy(name, req.Enabled)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func (c *StrategyController) GetStrategyMetrics(ctx *gin.Context) {
	name := ctx.Param("name")
	metrics, err := c.service.GetStrategyMetrics(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, metrics)
}

func (c *StrategyController) RunBacktest(ctx *gin.Context) {
	name := ctx.Param("name")
	var req struct {
		Data []byte `json:"data,omitempty"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := c.service.RunBacktest(name, req.Data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *StrategyController) CreateVersion(ctx *gin.Context) {
	name := ctx.Param("name")
	err := c.service.CreateStrategyVersion(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func (c *StrategyController) RollbackStrategy(ctx *gin.Context) {
	name := ctx.Param("name")
	var req struct {
		Version string `json:"version"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := c.service.RollbackStrategy(name, req.Version)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}
