package services

import (
	"ares_api/internal/api/dto"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type EditorServiceImpl struct {
	WorkspaceRoot string // Root directory for file operations (security boundary)
}

func NewEditorService(workspaceRoot string) *EditorServiceImpl {
	return &EditorServiceImpl{
		WorkspaceRoot: workspaceRoot,
	}
}

// validatePath ensures path is within workspace root (security check)
func (s *EditorServiceImpl) validatePath(requestedPath string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(requestedPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Ensure it's within workspace root
	absRoot, _ := filepath.Abs(s.WorkspaceRoot)
	if !strings.HasPrefix(absPath, absRoot) {
		return "", fmt.Errorf("path outside workspace: %s", requestedPath)
	}

	return absPath, nil
}

// getLanguageFromExtension determines Monaco language ID from file extension
func getLanguageFromExtension(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	languageMap := map[string]string{
		".go":     "go",
		".js":     "javascript",
		".ts":     "typescript",
		".jsx":    "javascript",
		".tsx":    "typescript",
		".json":   "json",
		".md":     "markdown",
		".html":   "html",
		".css":    "css",
		".scss":   "scss",
		".py":     "python",
		".java":   "java",
		".c":      "c",
		".cpp":    "cpp",
		".cs":     "csharp",
		".sql":    "sql",
		".sh":     "shell",
		".xml":    "xml",
		".yaml":   "yaml",
		".yml":    "yaml",
		".txt":    "plaintext",
		".log":    "plaintext",
		".axaml":  "xml",
		".xaml":   "xml",
		".csproj": "xml",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}
	return "plaintext"
}

// ReadFile reads file content
func (s *EditorServiceImpl) ReadFile(req dto.EditorFileRequest) (dto.EditorFileResponse, error) {
	validPath, err := s.validatePath(req.FilePath)
	if err != nil {
		return dto.EditorFileResponse{}, err
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		return dto.EditorFileResponse{}, fmt.Errorf("failed to read file: %w", err)
	}

	info, err := os.Stat(validPath)
	if err != nil {
		return dto.EditorFileResponse{}, fmt.Errorf("failed to stat file: %w", err)
	}

	return dto.EditorFileResponse{
		FilePath: req.FilePath,
		Content:  string(content),
		Language: getLanguageFromExtension(validPath),
		Size:     info.Size(),
	}, nil
}

// SaveFile saves file content
func (s *EditorServiceImpl) SaveFile(req dto.EditorSaveRequest) (dto.EditorSaveResponse, error) {
	validPath, err := s.validatePath(req.FilePath)
	if err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.FilePath,
			Success:  false,
			Message:  err.Error(),
		}, err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.FilePath,
			Success:  false,
			Message:  fmt.Sprintf("failed to create directory: %v", err),
		}, err
	}

	// Write file
	if err := os.WriteFile(validPath, []byte(req.Content), 0644); err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.FilePath,
			Success:  false,
			Message:  fmt.Sprintf("failed to write file: %v", err),
		}, err
	}

	return dto.EditorSaveResponse{
		FilePath: req.FilePath,
		Success:  true,
		Message:  "File saved successfully",
	}, nil
}

// ListFiles lists files in directory
func (s *EditorServiceImpl) ListFiles(req dto.EditorListRequest) (dto.EditorListResponse, error) {
	validPath, err := s.validatePath(req.DirectoryPath)
	if err != nil {
		return dto.EditorListResponse{}, err
	}

	var files []dto.EditorFileInfo
	maxDepth := req.MaxDepth
	if maxDepth == 0 {
		maxDepth = 5 // Default max depth
	}

	if req.Recursive {
		err = filepath.WalkDir(validPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			// Check depth
			relPath, _ := filepath.Rel(validPath, path)
			depth := strings.Count(relPath, string(os.PathSeparator))
			if depth > maxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip system folders
			if d.IsDir() {
				name := d.Name()
				if name == "node_modules" || name == ".git" || name == "vendor" ||
					name == "bin" || name == "obj" || name == "$RECYCLE.BIN" ||
					name == ".vs" || name == ".vscode" {
					return filepath.SkipDir
				}
			}

			// Skip the root directory itself
			if path == validPath {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}

			files = append(files, dto.EditorFileInfo{
				Name:     d.Name(),
				Path:     path,
				IsDir:    d.IsDir(),
				Size:     info.Size(),
				Modified: info.ModTime().Format(time.RFC3339),
			})

			return nil
		})
	} else {
		// Non-recursive: just list direct children
		entries, err := os.ReadDir(validPath)
		if err != nil {
			return dto.EditorListResponse{}, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			fullPath := filepath.Join(validPath, entry.Name())
			files = append(files, dto.EditorFileInfo{
				Name:     entry.Name(),
				Path:     fullPath,
				IsDir:    entry.IsDir(),
				Size:     info.Size(),
				Modified: info.ModTime().Format(time.RFC3339),
			})
		}
	}

	if err != nil && err != filepath.SkipAll {
		return dto.EditorListResponse{}, fmt.Errorf("failed to list directory: %w", err)
	}

	return dto.EditorListResponse{
		DirectoryPath: req.DirectoryPath,
		Files:         files,
		TotalFiles:    len(files),
	}, nil
}

// CreateFile creates a new file or directory
func (s *EditorServiceImpl) CreateFile(req dto.EditorCreateRequest) (dto.EditorSaveResponse, error) {
	validPath, err := s.validatePath(req.Path)
	if err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.Path,
			Success:  false,
			Message:  err.Error(),
		}, err
	}

	if req.IsDir {
		if err := os.MkdirAll(validPath, 0755); err != nil {
			return dto.EditorSaveResponse{
				FilePath: req.Path,
				Success:  false,
				Message:  fmt.Sprintf("failed to create directory: %v", err),
			}, err
		}
	} else {
		// Ensure parent directory exists
		dir := filepath.Dir(validPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return dto.EditorSaveResponse{
				FilePath: req.Path,
				Success:  false,
				Message:  fmt.Sprintf("failed to create parent directory: %v", err),
			}, err
		}

		// Create empty file
		if err := os.WriteFile(validPath, []byte(""), 0644); err != nil {
			return dto.EditorSaveResponse{
				FilePath: req.Path,
				Success:  false,
				Message:  fmt.Sprintf("failed to create file: %v", err),
			}, err
		}
	}

	return dto.EditorSaveResponse{
		FilePath: req.Path,
		Success:  true,
		Message:  "Created successfully",
	}, nil
}

// DeleteFile deletes a file or directory
func (s *EditorServiceImpl) DeleteFile(req dto.EditorDeleteRequest) (dto.EditorSaveResponse, error) {
	validPath, err := s.validatePath(req.Path)
	if err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.Path,
			Success:  false,
			Message:  err.Error(),
		}, err
	}

	if err := os.RemoveAll(validPath); err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.Path,
			Success:  false,
			Message:  fmt.Sprintf("failed to delete: %v", err),
		}, err
	}

	return dto.EditorSaveResponse{
		FilePath: req.Path,
		Success:  true,
		Message:  "Deleted successfully",
	}, nil
}

// RenameFile renames/moves a file or directory
func (s *EditorServiceImpl) RenameFile(req dto.EditorRenameRequest) (dto.EditorSaveResponse, error) {
	validOldPath, err := s.validatePath(req.OldPath)
	if err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.OldPath,
			Success:  false,
			Message:  err.Error(),
		}, err
	}

	validNewPath, err := s.validatePath(req.NewPath)
	if err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.NewPath,
			Success:  false,
			Message:  err.Error(),
		}, err
	}

	// Ensure parent directory of new path exists
	newDir := filepath.Dir(validNewPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.NewPath,
			Success:  false,
			Message:  fmt.Sprintf("failed to create parent directory: %v", err),
		}, err
	}

	if err := os.Rename(validOldPath, validNewPath); err != nil {
		return dto.EditorSaveResponse{
			FilePath: req.OldPath,
			Success:  false,
			Message:  fmt.Sprintf("failed to rename: %v", err),
		}, err
	}

	return dto.EditorSaveResponse{
		FilePath: req.NewPath,
		Success:  true,
		Message:  "Renamed successfully",
	}, nil
}
