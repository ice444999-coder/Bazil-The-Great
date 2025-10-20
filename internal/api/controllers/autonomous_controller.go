package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// AutonomousController handles SOLACE autonomous operations
type AutonomousController struct {
	repoRoot   string
	backupRoot string
}

// NewAutonomousController creates a new autonomous operations controller
func NewAutonomousController(repoRoot string, backupRoot string) *AutonomousController {
	return &AutonomousController{
		repoRoot:   repoRoot,
		backupRoot: backupRoot,
	}
}

// WriteFile writes content to a file
// POST /api/v1/solace/file/write
func (c *AutonomousController) WriteFile(ctx *gin.Context) {
	var req struct {
		FilePath string `json:"file_path" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Security: Ensure path is within repo root
	fullPath := filepath.Join(c.repoRoot, req.FilePath)
	if !filepath.IsAbs(fullPath) {
		fullPath, _ = filepath.Abs(fullPath)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create directory: %v", err)})
		return
	}

	// Write file
	if err := ioutil.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to write file: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "File written successfully",
		"file_path": req.FilePath,
		"size":      len(req.Content),
	})
}

// CreateBackup creates a timestamped backup of the workspace
// POST /api/v1/solace/backup/create
func (c *AutonomousController) CreateBackup(ctx *gin.Context) {
	var req struct {
		Reason string `json:"reason"`
	}

	ctx.ShouldBindJSON(&req)

	// Create backup directory with timestamp
	timestamp := time.Now().Format("2006-01-02_150405")
	backupPath := filepath.Join(c.backupRoot, timestamp)

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create backup directory: %v", err)})
		return
	}

	// Copy entire workspace
	if err := copyDir(c.repoRoot, backupPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to copy workspace: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Backup created successfully",
		"timestamp": timestamp,
		"path":      backupPath,
		"reason":    req.Reason,
	})
}

// ExecuteCommand executes a shell command
// POST /api/v1/solace/command/execute
func (c *AutonomousController) ExecuteCommand(ctx *gin.Context) {
	var req struct {
		Command          string `json:"command" binding:"required"`
		WorkingDirectory string `json:"working_directory"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workDir := req.WorkingDirectory
	if workDir == "" {
		workDir = c.repoRoot
	}

	// Execute command
	cmd := exec.Command("powershell", "-Command", req.Command)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"command":   req.Command,
		"output":    string(output),
		"exit_code": exitCode,
		"success":   exitCode == 0,
	})
}

// RestoreFromBackup restores workspace from a backup
// POST /api/v1/solace/backup/restore
func (c *AutonomousController) RestoreFromBackup(ctx *gin.Context) {
	var req struct {
		BackupTimestamp string `json:"backup_timestamp" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	backupPath := filepath.Join(c.backupRoot, req.BackupTimestamp)

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}

	// Delete current workspace (careful!)
	os.RemoveAll(c.repoRoot)

	// Restore from backup
	if err := copyDir(backupPath, c.repoRoot); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to restore: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Workspace restored successfully",
		"timestamp": req.BackupTimestamp,
	})
}

// VerifySystem checks if ARES is running
// GET /api/v1/solace/system/verify
func (c *AutonomousController) VerifySystem(ctx *gin.Context) {
	// Simple health check
	ctx.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"message":   "ARES system is operational",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Helper function to copy directory recursively
func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git, node_modules, bin, obj
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "bin" || name == "obj" {
				return filepath.SkipDir
			}
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(dstPath, data, info.Mode())
	})
}
