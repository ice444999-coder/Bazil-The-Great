/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"ares_api/pkg/llm"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LLMHealthController handles LLM health monitoring
type LLMHealthController struct {
	Client *llm.Client
}

// NewLLMHealthController creates a new LLM health controller
func NewLLMHealthController(client *llm.Client) *LLMHealthController {
	return &LLMHealthController{Client: client}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       string        `json:"status"`
	Healthy      bool          `json:"healthy"`
	Model        string        `json:"model"`
	LatencyMs    int64         `json:"latency_ms"`
	ErrorMessage string        `json:"error_message,omitempty"`
	CheckedAt    string        `json:"checked_at"`
}

// CheckHealth godoc
// @Summary Check LLM service health
// @Description Checks if DeepSeek-R1 14B is available and responsive
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health/llm [get]
func (ctrl *LLMHealthController) CheckHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Now()
	healthStatus, err := ctrl.Client.Health(ctx)
	latency := time.Since(startTime)

	response := HealthResponse{
		Model:     ctrl.Client.Model,
		LatencyMs: latency.Milliseconds(),
		CheckedAt: time.Now().Format(time.RFC3339),
	}

	if err != nil || !healthStatus.Healthy {
		response.Status = "unhealthy"
		response.Healthy = false
		if err != nil {
			response.ErrorMessage = err.Error()
		} else {
			response.ErrorMessage = healthStatus.ErrorMessage
		}
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response.Status = "healthy"
	response.Healthy = true
	c.JSON(http.StatusOK, response)
}
