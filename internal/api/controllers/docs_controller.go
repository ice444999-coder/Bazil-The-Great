package controllers

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// DocsController handles documentation and markdown file serving
type DocsController struct {
	workspaceRoot string
}

// NewDocsController creates a new docs controller
func NewDocsController(workspaceRoot string) *DocsController {
	return &DocsController{
		workspaceRoot: workspaceRoot,
	}
}

// GetAllDocs returns all markdown documents in the workspace
// @Summary Get all documentation files
// @Description Returns list of all .md files with metadata
// @Tags Documentation
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /docs/list [get]
func (dc *DocsController) GetAllDocs(c *gin.Context) {
	docs := []DocFile{}
	
	// Scan workspace root
	err := filepath.Walk(dc.workspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		
		// Only include .md files
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			relativePath, _ := filepath.Rel(dc.workspaceRoot, path)
			
			doc := DocFile{
				Name:         info.Name(),
				Path:         relativePath,
				FullPath:     path,
				Size:         info.Size(),
				ModifiedAt:   info.ModTime(),
				Category:     categorizeDoc(info.Name()),
			}
			
			docs = append(docs, doc)
		}
		
		return nil
	})
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"documents": docs,
		"count":     len(docs),
	})
}

// GetDocContent returns the content of a specific document
// @Summary Get document content
// @Description Returns the markdown content of a document
// @Tags Documentation
// @Produce json
// @Param path query string true "Relative path to document"
// @Success 200 {object} map[string]interface{}
// @Router /docs/content [get]
func (dc *DocsController) GetDocContent(c *gin.Context) {
	relativePath := c.Query("path")
	if relativePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path parameter required"})
		return
	}
	
	// Security: prevent directory traversal
	if strings.Contains(relativePath, "..") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid path"})
		return
	}
	
	fullPath := filepath.Join(dc.workspaceRoot, relativePath)
	
	// Check file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	
	// Read file
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"name":        info.Name(),
		"path":        relativePath,
		"size":        info.Size(),
		"modified_at": info.ModTime(),
		"content":     string(content),
		"category":    categorizeDoc(info.Name()),
	})
}

// GetDocCategories returns documents grouped by category
// @Summary Get document categories
// @Description Returns documents organized by category
// @Tags Documentation
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /docs/categories [get]
func (dc *DocsController) GetDocCategories(c *gin.Context) {
	categories := make(map[string][]DocFile)
	
	// Scan workspace root
	err := filepath.Walk(dc.workspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			relativePath, _ := filepath.Rel(dc.workspaceRoot, path)
			category := categorizeDoc(info.Name())
			
			doc := DocFile{
				Name:       info.Name(),
				Path:       relativePath,
				FullPath:   path,
				Size:       info.Size(),
				ModifiedAt: info.ModTime(),
				Category:   category,
			}
			
			categories[category] = append(categories[category], doc)
		}
		
		return nil
	})
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// DocFile represents a documentation file
type DocFile struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	FullPath   string    `json:"full_path,omitempty"`
	Size       int64     `json:"size"`
	ModifiedAt interface{} `json:"modified_at"`
	Category   string    `json:"category"`
}

// categorizeDoc categorizes a document based on its name
func categorizeDoc(filename string) string {
	lower := strings.ToLower(filename)
	
	switch {
	case strings.Contains(lower, "masterplan"):
		return "Masterplan"
	case strings.Contains(lower, "gate") && strings.Contains(lower, "verification"):
		return "Gate Verification"
	case strings.Contains(lower, "phase") && strings.Contains(lower, "complete"):
		return "Phase Reports"
	case strings.Contains(lower, "implementation") || strings.Contains(lower, "status"):
		return "Implementation Status"
	case strings.Contains(lower, "architecture") || strings.Contains(lower, "compliance"):
		return "Architecture"
	case strings.Contains(lower, "ace") || strings.Contains(lower, "framework"):
		return "ACE Framework"
	case strings.Contains(lower, "solace") || strings.Contains(lower, "awakening"):
		return "SOLACE"
	case strings.Contains(lower, "trading") || strings.Contains(lower, "sandbox"):
		return "Trading System"
	case strings.Contains(lower, "memory") || strings.Contains(lower, "semantic"):
		return "Memory System"
	case strings.Contains(lower, "llm") || strings.Contains(lower, "deepseek"):
		return "LLM Infrastructure"
	case strings.Contains(lower, "ui") || strings.Contains(lower, "desktop"):
		return "UI/Desktop"
	case strings.Contains(lower, "guide") || strings.Contains(lower, "how_to"):
		return "Guides"
	case strings.Contains(lower, "readme"):
		return "README"
	case strings.Contains(lower, "session_summary"):
		return "Session Summaries"
	case strings.Contains(lower, "roadmap"):
		return "Roadmaps"
	case strings.Contains(lower, "security") || strings.Contains(lower, "stability"):
		return "Security & Stability"
	default:
		return "Other"
	}
}
