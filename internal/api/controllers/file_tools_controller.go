package controllers

import (
	"ares_api/pkg/llm"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// FileToolsController handles file access operations for AI
type FileToolsController struct {
	FileTools *llm.FileAccessTools
}

// NewFileToolsController creates a new file tools controller
func NewFileToolsController(fileTools *llm.FileAccessTools) *FileToolsController {
	return &FileToolsController{FileTools: fileTools}
}

// ReadFileRequest represents a file read request
type ReadFileRequest struct {
	Path     string `json:"path" binding:"required"`
	MaxLines int    `json:"max_lines,omitempty"`
}

// ListDirectoryRequest represents a directory listing request
type ListDirectoryRequest struct {
	Path string `json:"path" binding:"required"`
}

// SearchCodeRequest represents a code search request
type SearchCodeRequest struct {
	Pattern        string   `json:"pattern" binding:"required"`
	Directory      string   `json:"directory" binding:"required"`
	FileExtensions []string `json:"file_extensions,omitempty"`
	MaxResults     int      `json:"max_results,omitempty"`
}

// ReadFile godoc
// @Summary Read a file from workspace
// @Description Reads a file's content (with security restrictions)
// @Tags file-tools
// @Accept json
// @Produce json
// @Param request body ReadFileRequest true "File Read Request"
// @Success 200 {object} llm.ReadFileResult
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /file-tools/read [post]
func (ctrl *FileToolsController) ReadFile(c *gin.Context) {
	var req ReadFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := ctrl.FileTools.ReadFile(ctx, req.Path, req.MaxLines)
	if err != nil {
		if result != nil && result.Error == "Path not allowed - outside workspace" {
			c.JSON(http.StatusForbidden, result)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListDirectory godoc
// @Summary List directory contents
// @Description Lists files and folders in a directory
// @Tags file-tools
// @Accept json
// @Produce json
// @Param request body ListDirectoryRequest true "Directory Listing Request"
// @Success 200 {object} llm.ListDirectoryResult
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /file-tools/list [post]
func (ctrl *FileToolsController) ListDirectory(c *gin.Context) {
	var req ListDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := ctrl.FileTools.ListDirectory(ctx, req.Path)
	if err != nil {
		if result != nil && result.Error == "Path not allowed - outside workspace" {
			c.JSON(http.StatusForbidden, result)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SearchCode godoc
// @Summary Search for code patterns
// @Description Searches for a pattern in files within a directory
// @Tags file-tools
// @Accept json
// @Produce json
// @Param request body SearchCodeRequest true "Code Search Request"
// @Success 200 {array} llm.SearchCodeResult
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /file-tools/search [post]
func (ctrl *FileToolsController) SearchCode(c *gin.Context) {
	var req SearchCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default max results if not specified
	if req.MaxResults == 0 {
		req.MaxResults = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := ctrl.FileTools.SearchCode(ctx, req.Pattern, req.Directory, req.FileExtensions, req.MaxResults)
	if err != nil {
		if err.Error() == "path not allowed: "+req.Directory {
			c.JSON(http.StatusForbidden, gin.H{"error": "Path not allowed - outside workspace"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}
