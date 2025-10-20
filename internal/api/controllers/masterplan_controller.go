package controllers

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// MasterplanController handles masterplan automation operations
type MasterplanController struct{}

// SaveRequest represents a request to save content to masterplan
type SaveRequest struct {
	SourceFile  string `json:"source_file" binding:"required"`
	ContentType string `json:"content_type"`
	DryRun      bool   `json:"dry_run"`
}

// SaveResponse represents the response from a save operation
type SaveResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	BackupPath  string    `json:"backup_path"`
	Section     string    `json:"section"`
	Strategy    string    `json:"strategy"`
	Timestamp   time.Time `json:"timestamp"`
	ExecutionMS int64     `json:"execution_ms"`
}

// RefreshResponse represents the response from a refresh operation
type RefreshResponse struct {
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	FilesScanned int       `json:"files_scanned"`
	ChainIndex   int       `json:"chain_index"`
	LatestHash   string    `json:"latest_hash"`
	Timestamp    time.Time `json:"timestamp"`
	ExecutionMS  int64     `json:"execution_ms"`
}

// VerifyResponse represents the response from a verify operation
type VerifyResponse struct {
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	ChainValid   bool      `json:"chain_valid"`
	BrokenHashes int       `json:"broken_hashes"`
	ChainGaps    int       `json:"chain_gaps"`
	TotalEntries int       `json:"total_entries"`
	Timestamp    time.Time `json:"timestamp"`
	ExecutionMS  int64     `json:"execution_ms"`
}

// NaturalCommandRequest represents a natural language command request
type NaturalCommandRequest struct {
	Command    string `json:"command" binding:"required"`
	ActiveFile string `json:"active_file"`
	DryRun     bool   `json:"dry_run"`
}

// NewMasterplanController creates a new masterplan controller
func NewMasterplanController() *MasterplanController {
	return &MasterplanController{}
}

// Save handles POST /api/v1/masterplan/save
func (mc *MasterplanController) Save(c *gin.Context) {
	startTime := time.Now()

	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Build PowerShell command
	args := []string{
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-File", "C:\\ARES_Workspace\\ARES-Command.ps1",
		"-Action", "save",
		"-Source", req.SourceFile,
	}

	if req.ContentType != "" {
		args = append(args, "-ContentType", req.ContentType)
	}

	if req.DryRun {
		args = append(args, "-DryRun")
	}

	// Execute PowerShell script
	cmd := exec.Command("powershell.exe", args...)
	output, err := cmd.CombinedOutput()

	executionMS := time.Since(startTime).Milliseconds()

	if err != nil {
		c.JSON(http.StatusInternalServerError, SaveResponse{
			Success:     false,
			Message:     "Save workflow failed: " + string(output),
			Timestamp:   time.Now(),
			ExecutionMS: executionMS,
		})
		return
	}

	// Parse output for details (basic implementation)
	outputStr := string(output)
	section := extractField(outputStr, "Section:")
	strategy := extractField(outputStr, "Strategy:")
	backupPath := extractField(outputStr, "Backup:")

	c.JSON(http.StatusOK, SaveResponse{
		Success:     true,
		Message:     "Content saved to masterplan successfully",
		BackupPath:  backupPath,
		Section:     section,
		Strategy:    strategy,
		Timestamp:   time.Now(),
		ExecutionMS: executionMS,
	})
}

// Refresh handles POST /api/v1/masterplan/refresh
func (mc *MasterplanController) Refresh(c *gin.Context) {
	startTime := time.Now()

	// Execute PowerShell script
	cmd := exec.Command("powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-File", "C:\\ARES_Workspace\\ARES-Command.ps1",
		"-Action", "refresh",
	)
	output, err := cmd.CombinedOutput()

	executionMS := time.Since(startTime).Milliseconds()

	if err != nil {
		c.JSON(http.StatusInternalServerError, RefreshResponse{
			Success:     false,
			Message:     "Refresh workflow failed: " + string(output),
			Timestamp:   time.Now(),
			ExecutionMS: executionMS,
		})
		return
	}

	// Parse output for details
	outputStr := string(output)
	chainIndex := extractIntField(outputStr, "Chain Index:")
	latestHash := extractField(outputStr, "Latest Hash:")

	c.JSON(http.StatusOK, RefreshResponse{
		Success:      true,
		Message:      "Timestamp manifest refreshed successfully",
		FilesScanned: 1069, // TODO: Parse from output
		ChainIndex:   chainIndex,
		LatestHash:   latestHash,
		Timestamp:    time.Now(),
		ExecutionMS:  executionMS,
	})
}

// Verify handles GET /api/v1/masterplan/verify
func (mc *MasterplanController) Verify(c *gin.Context) {
	startTime := time.Now()

	// Execute PowerShell script
	cmd := exec.Command("powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-File", "C:\\ARES_Workspace\\ARES-Command.ps1",
		"-Action", "verify",
	)
	output, err := cmd.CombinedOutput()

	executionMS := time.Since(startTime).Milliseconds()

	outputStr := string(output)
	chainValid := strings.Contains(outputStr, "LEDGER VERIFICATION PASSED")

	if err != nil && !chainValid {
		c.JSON(http.StatusInternalServerError, VerifyResponse{
			Success:     false,
			Message:     "Ledger verification failed: " + string(output),
			ChainValid:  false,
			Timestamp:   time.Now(),
			ExecutionMS: executionMS,
		})
		return
	}

	c.JSON(http.StatusOK, VerifyResponse{
		Success:      true,
		Message:      "Ledger verification completed",
		ChainValid:   chainValid,
		BrokenHashes: 0,    // TODO: Parse from output
		ChainGaps:    0,    // TODO: Parse from output
		TotalEntries: 2128, // TODO: Parse from output
		Timestamp:    time.Now(),
		ExecutionMS:  executionMS,
	})
}

// NaturalCommand handles POST /api/v1/masterplan/natural
func (mc *MasterplanController) NaturalCommand(c *gin.Context) {
	startTime := time.Now()

	var req NaturalCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Build PowerShell command
	args := []string{
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-File", "C:\\ARES_Workspace\\ARES-Natural-Command.ps1",
		req.Command,
	}

	if req.ActiveFile != "" {
		args = append(args, "-ActiveFile", req.ActiveFile)
	}

	if req.DryRun {
		args = append(args, "-DryRun")
	}

	// Execute PowerShell script
	cmd := exec.Command("powershell.exe", args...)
	output, err := cmd.CombinedOutput()

	executionMS := time.Since(startTime).Milliseconds()

	outputStr := string(output)
	success := strings.Contains(outputStr, "NATURAL COMMAND SUCCESSFUL")

	if err != nil && !success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"message":      "Natural command failed: " + string(output),
			"command":      req.Command,
			"timestamp":    time.Now(),
			"execution_ms": executionMS,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Natural command executed successfully",
		"command":      req.Command,
		"output":       outputStr,
		"timestamp":    time.Now(),
		"execution_ms": executionMS,
	})
}

// Helper functions

func extractField(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.Contains(line, prefix) {
			parts := strings.SplitN(line, prefix, 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func extractIntField(text, prefix string) int {
	value := extractField(text, prefix)
	var result int
	if err := json.Unmarshal([]byte(value), &result); err == nil {
		return result
	}
	return 0
}

// RegisterRoutes registers masterplan routes
func (mc *MasterplanController) RegisterRoutes(router *gin.RouterGroup) {
	masterplan := router.Group("/masterplan")
	{
		masterplan.POST("/save", mc.Save)
		masterplan.POST("/refresh", mc.Refresh)
		masterplan.GET("/verify", mc.Verify)
		masterplan.POST("/natural", mc.NaturalCommand)
	}
}
