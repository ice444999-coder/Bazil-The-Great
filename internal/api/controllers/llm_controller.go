package controllers

import (
	"ares_api/internal/common"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type LLMController struct{}

func NewLLMController() *LLMController {
	return &LLMController{}
}

// @Summary Get local LLM status
// @Description Check if Ollama is installed and list available models
// @Tags LLM
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /llm/status [get]
func (lc *LLMController) GetStatus(c *gin.Context) {
	status := map[string]interface{}{
		"ollama_installed": false,
		"models":           []string{},
		"message":          "",
	}

	// Check if Ollama is installed
	cmd := exec.Command("ollama", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		status["message"] = "Ollama not installed or not in PATH. Install from https://ollama.ai"
		common.JSON(c, http.StatusOK, status)
		return
	}

	status["ollama_installed"] = true

	// Parse model list
	lines := strings.Split(string(output), "\n")
	models := []string{}
	for i, line := range lines {
		if i == 0 || line == "" {
			continue // Skip header and empty lines
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			models = append(models, fields[0])
		}
	}

	status["models"] = models
	status["message"] = "Ollama is ready for local LLM inference"

	common.JSON(c, http.StatusOK, status)
}

// @Summary Test local LLM inference
// @Description Send a test message to Ollama (if available) or fallback to Claude API
// @Tags LLM
// @Accept json
// @Produce json
// @Param request body map[string]string true "Test request"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /llm/test [post]
func (lc *LLMController) TestInference(c *gin.Context) {
	var req struct {
		Message string `json:"message"`
		Model   string `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Model == "" {
		req.Model = "llama2"
	}

	// Try Ollama first
	cmd := exec.Command("ollama", "run", req.Model, req.Message)
	output, err := cmd.CombinedOutput()

	if err != nil {
		common.JSON(c, http.StatusOK, gin.H{
			"source":   "fallback",
			"message":  "Ollama not available. Would fallback to Claude API in production.",
			"response": "",
		})
		return
	}

	common.JSON(c, http.StatusOK, gin.H{
		"source":   "ollama",
		"model":    req.Model,
		"response": string(output),
	})
}
