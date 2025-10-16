package handlers

import (
	"net/http"

	"ares_api/internal/models"
	"ares_api/internal/repositories"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	repo *repositories.AgentRepository
}

func NewAgentHandler(repo *repositories.AgentRepository) *AgentHandler {
	return &AgentHandler{repo: repo}
}

// GetAgents retrieves all registered agents
// GET /api/v1/agents
func (h *AgentHandler) GetAgents(c *gin.Context) {
	agents, err := h.repo.GetAllAgents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, agents)
}

// GetAgent retrieves a specific agent by name
// GET /api/v1/agents/:name
func (h *AgentHandler) GetAgent(c *gin.Context) {
	name := c.Param("name")
	agent, err := h.repo.GetAgentByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

// CreateTask creates a new task in the queue
// POST /api/v1/agents/tasks
func (h *AgentHandler) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default created_by to DAVID if not specified
	createdBy := "DAVID"
	if user, exists := c.Get("user"); exists {
		if userStr, ok := user.(string); ok {
			createdBy = userStr
		}
	}

	taskID, err := h.repo.CreateTask(&req, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Auto-assign to SOLACE if not specified
	if err := h.repo.AssignTask(taskID, "SOLACE"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Task created but assignment failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":     taskID,
		"assigned_to": "SOLACE",
		"message":     "Task created and assigned to SOLACE",
	})
}

// GetTask retrieves a specific task by ID
// GET /api/v1/agents/tasks/:id
func (h *AgentHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.repo.GetTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// GetPendingTasks retrieves pending tasks
// GET /api/v1/agents/tasks/pending
func (h *AgentHandler) GetPendingTasks(c *gin.Context) {
	limit := 10
	tasks, err := h.repo.GetPendingTasks(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// AssignTask assigns a task to a specific agent
// POST /api/v1/agents/tasks/:id/assign
func (h *AgentHandler) AssignTask(c *gin.Context) {
	taskID := c.Param("id")
	var req models.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.AssignTask(taskID, req.AgentName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task assigned successfully"})
}

// CompleteTask marks a task as completed
// POST /api/v1/agents/tasks/:id/complete
func (h *AgentHandler) CompleteTask(c *gin.Context) {
	taskID := c.Param("id")
	var result map[string]interface{}
	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CompleteTask(taskID, result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task completed successfully"})
}

// FailTask marks a task as failed
// POST /api/v1/agents/tasks/:id/fail
func (h *AgentHandler) FailTask(c *gin.Context) {
	taskID := c.Param("id")
	var req struct {
		ErrorLog string `json:"error_log"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.FailTask(taskID, req.ErrorLog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task marked as failed"})
}

// GetAgentPerformance retrieves performance history for an agent
// GET /api/v1/agents/:name/performance
func (h *AgentHandler) GetAgentPerformance(c *gin.Context) {
	agentName := c.Param("name")
	limit := 100

	history, err := h.repo.GetAgentPerformance(agentName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

// GetFiles retrieves all files from registry
// GET /api/v1/agents/files
func (h *AgentHandler) GetFiles(c *gin.Context) {
	files, err := h.repo.GetAllFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

// GetFileByPath retrieves a specific file
// GET /api/v1/agents/files/by-path
func (h *AgentHandler) GetFileByPath(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter required"})
		return
	}

	file, err := h.repo.GetFileByPath(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	c.JSON(http.StatusOK, file)
}

// GetRecentBuilds retrieves recent build history
// GET /api/v1/agents/builds
func (h *AgentHandler) GetRecentBuilds(c *gin.Context) {
	limit := 20
	builds, err := h.repo.GetRecentBuilds(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, builds)
}

// RegisterRoutes registers all agent-related routes
func (h *AgentHandler) RegisterRoutes(router *gin.RouterGroup) {
	agents := router.Group("/agents")
	{
		// Agent endpoints
		agents.GET("", h.GetAgents)
		agents.GET("/:name", h.GetAgent)
		agents.GET("/:name/performance", h.GetAgentPerformance)

		// Task endpoints
		agents.POST("/tasks", h.CreateTask)
		agents.GET("/tasks/pending", h.GetPendingTasks)
		agents.GET("/tasks/:id", h.GetTask)
		agents.POST("/tasks/:id/assign", h.AssignTask)
		agents.POST("/tasks/:id/complete", h.CompleteTask)
		agents.POST("/tasks/:id/fail", h.FailTask)

		// File registry endpoints
		agents.GET("/files", h.GetFiles)
		agents.GET("/files/by-path", h.GetFileByPath)

		// Build history endpoints
		agents.GET("/builds", h.GetRecentBuilds)
	}
}
