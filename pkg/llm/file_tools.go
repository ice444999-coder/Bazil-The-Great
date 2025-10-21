/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package llm

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FileAccessTools provides file system operations for the LLM
type FileAccessTools struct {
	WorkspaceRoot string
	AllowedPaths  []string // Whitelist of allowed directories
}

// NewFileAccessTools creates a new file access tools instance
func NewFileAccessTools(workspaceRoot string) *FileAccessTools {
	// Default allowed paths
	allowedPaths := []string{
		workspaceRoot,
		filepath.Join(workspaceRoot, "ARES_API"),
		filepath.Join(workspaceRoot, "ARES_UI"),
		filepath.Join(workspaceRoot, "ARES_Desktop_App"),
	}

	return &FileAccessTools{
		WorkspaceRoot: workspaceRoot,
		AllowedPaths:  allowedPaths,
	}
}

// ReadFileResult represents the result of reading a file
type ReadFileResult struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Lines    int    `json:"lines"`
	SizeKB   int    `json:"size_kb"`
	Error    string `json:"error,omitempty"`
}

// ListDirectoryResult represents a directory listing
type ListDirectoryResult struct {
	Path    string   `json:"path"`
	Files   []string `json:"files"`
	Folders []string `json:"folders"`
	Error   string   `json:"error,omitempty"`
}

// SearchCodeResult represents search results
type SearchCodeResult struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Content string `json:"content"`
	Match   string `json:"match"`
}

// isPathAllowed checks if a path is within allowed directories
func (f *FileAccessTools) isPathAllowed(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, allowed := range f.AllowedPaths {
		absAllowed, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		
		// Check if path is within allowed directory
		if strings.HasPrefix(absPath, absAllowed) {
			return true
		}
	}

	return false
}

// ReadFile reads a file and returns its content
func (f *FileAccessTools) ReadFile(ctx context.Context, path string, maxLines int) (*ReadFileResult, error) {
	if !f.isPathAllowed(path) {
		return &ReadFileResult{
			Path:  path,
			Error: "Path not allowed - outside workspace",
		}, fmt.Errorf("path not allowed: %s", path)
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return &ReadFileResult{
			Path:  path,
			Error: err.Error(),
		}, err
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Limit lines if requested
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
		contentStr = strings.Join(lines, "\n") + fmt.Sprintf("\n... (truncated, showing first %d lines)", maxLines)
	}

	return &ReadFileResult{
		Path:    path,
		Content: contentStr,
		Lines:   len(lines),
		SizeKB:  len(content) / 1024,
	}, nil
}

// ListDirectory lists files and folders in a directory
func (f *FileAccessTools) ListDirectory(ctx context.Context, path string) (*ListDirectoryResult, error) {
	if !f.isPathAllowed(path) {
		return &ListDirectoryResult{
			Path:  path,
			Error: "Path not allowed - outside workspace",
		}, fmt.Errorf("path not allowed: %s", path)
	}

	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return &ListDirectoryResult{
			Path:  path,
			Error: err.Error(),
		}, err
	}

	var files []string
	var folders []string

	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
		} else {
			files = append(files, entry.Name())
		}
	}

	return &ListDirectoryResult{
		Path:    path,
		Files:   files,
		Folders: folders,
	}, nil
}

// SearchCode searches for a pattern in files within a directory
func (f *FileAccessTools) SearchCode(ctx context.Context, pattern string, directory string, fileExtensions []string, maxResults int) ([]*SearchCodeResult, error) {
	if !f.isPathAllowed(directory) {
		return nil, fmt.Errorf("path not allowed: %s", directory)
	}

	var results []*SearchCodeResult
	resultCount := 0

	// Walk directory tree
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check file extension filter
		if len(fileExtensions) > 0 {
			ext := filepath.Ext(path)
			matched := false
			for _, allowedExt := range fileExtensions {
				if ext == allowedExt {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		// Read file
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		// Search for pattern in each line
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
				results = append(results, &SearchCodeResult{
					Path:    path,
					Line:    i + 1,
					Content: strings.TrimSpace(line),
					Match:   pattern,
				})

				resultCount++
				if maxResults > 0 && resultCount >= maxResults {
					return filepath.SkipDir // Stop searching
				}
			}
		}

		return nil
	})

	return results, err
}
