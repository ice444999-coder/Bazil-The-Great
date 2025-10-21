/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// FileOpsController handles file operations for SOLACE orchestration
type FileOpsController struct {
	repoRoot string // Root path of ARES repository
}

// NewFileOpsController creates a new file operations controller
func NewFileOpsController(repoRoot string) *FileOpsController {
	return &FileOpsController{
		repoRoot: repoRoot,
	}
}

// ReadFile reads a file from the repository
// POST /api/v1/solace/file/read
func (c *FileOpsController) ReadFile(ctx *gin.Context) {
	var req struct {
		FilePath string `json:"file_path" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Security: Ensure path is within repo root
	fullPath := filepath.Join(c.repoRoot, req.FilePath)
	if !strings.HasPrefix(fullPath, c.repoRoot) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Path outside repository"})
		return
	}

	// Read file
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Failed to read file: %v", err)})
		return
	}

	// Calculate hash
	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	// Get file info
	fileInfo, _ := os.Stat(fullPath)

	ctx.JSON(http.StatusOK, gin.H{
		"file_path":     req.FilePath,
		"content":       string(content),
		"content_hash":  hashStr,
		"size_bytes":    fileInfo.Size(),
		"last_modified": fileInfo.ModTime(),
	})
}

// ListDirectory lists files in a directory
// POST /api/v1/solace/file/list
func (c *FileOpsController) ListDirectory(ctx *gin.Context) {
	var req struct {
		DirPath string `json:"dir_path" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Security: Ensure path is within repo root
	fullPath := filepath.Join(c.repoRoot, req.DirPath)
	if !strings.HasPrefix(fullPath, c.repoRoot) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Path outside repository"})
		return
	}

	// Read directory
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Failed to read directory: %v", err)})
		return
	}

	// Build file list
	var fileList []map[string]interface{}
	for _, file := range files {
		fileList = append(fileList, map[string]interface{}{
			"name":          file.Name(),
			"is_dir":        file.IsDir(),
			"size_bytes":    file.Size(),
			"last_modified": file.ModTime(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"dir_path": req.DirPath,
		"files":    fileList,
		"count":    len(fileList),
	})
}
