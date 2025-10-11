package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FileSystemReader provides read access to repository files
type FileSystemReader struct {
	RepoPath string
}

// NewFileSystemReader creates a new file system reader
func NewFileSystemReader(repoPath string) *FileSystemReader {
	return &FileSystemReader{RepoPath: repoPath}
}

// ReadFile reads a file from the repository
func (fsr *FileSystemReader) ReadFile(relativePath string) (string, error) {
	fullPath := filepath.Join(fsr.RepoPath, relativePath)

	// Security check: ensure path is within repo
	if !strings.HasPrefix(fullPath, fsr.RepoPath) {
		return "", fmt.Errorf("path traversal attempt blocked")
	}

	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// ListFiles lists files in a directory (non-recursive)
func (fsr *FileSystemReader) ListFiles(relativePath string) ([]string, error) {
	fullPath := filepath.Join(fsr.RepoPath, relativePath)

	// Security check
	if !strings.HasPrefix(fullPath, fsr.RepoPath) {
		return nil, fmt.Errorf("path traversal attempt blocked")
	}

	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}

// FindFiles searches for files matching a pattern (recursive)
func (fsr *FileSystemReader) FindFiles(pattern string) ([]string, error) {
	var matches []string

	err := filepath.Walk(fsr.RepoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common excludes
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == ".env" {
				return filepath.SkipDir
			}
			return nil
		}

		// Match pattern
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}

		if matched {
			// Return relative path
			relPath, err := filepath.Rel(fsr.RepoPath, path)
			if err != nil {
				return err
			}
			matches = append(matches, relPath)
		}

		return nil
	})

	return matches, err
}

// GetFileInfo returns info about a file
func (fsr *FileSystemReader) GetFileInfo(relativePath string) (os.FileInfo, error) {
	fullPath := filepath.Join(fsr.RepoPath, relativePath)

	// Security check
	if !strings.HasPrefix(fullPath, fsr.RepoPath) {
		return nil, fmt.Errorf("path traversal attempt blocked")
	}

	return os.Stat(fullPath)
}
