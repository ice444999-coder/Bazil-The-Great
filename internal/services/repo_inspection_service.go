package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ares_api/internal/models"

	"gorm.io/gorm"
)

// RepoInspectionService handles repository file scanning and caching
type RepoInspectionService struct {
	db       *gorm.DB
	repoRoot string
}

// NewRepoInspectionService creates a new repo inspection service
func NewRepoInspectionService(db *gorm.DB, repoRoot string) *RepoInspectionService {
	return &RepoInspectionService{
		db:       db,
		repoRoot: repoRoot,
	}
}

// ScanRepository scans the entire repository and updates file cache
func (s *RepoInspectionService) ScanRepository() error {
	ignoreDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"bin":          true,
		"obj":          true,
		"dist":         true,
		".vs":          true,
		"venv":         true,
		".venv":        true,
	}

	return filepath.Walk(s.repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored directories
		if info.IsDir() {
			if ignoreDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(s.repoRoot, path)
		if err != nil {
			return err
		}

		// Only track code/config files
		ext := filepath.Ext(path)
		trackedExtensions := map[string]bool{
			".go": true, ".cs": true, ".axaml": true, ".sql": true,
			".py": true, ".js": true, ".ts": true, ".tsx": true,
			".html": true, ".css": true, ".json": true, ".md": true,
			".ps1": true, ".mod": true, ".sum": true,
		}

		if !trackedExtensions[ext] {
			return nil
		}

		// Read file content
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		// Calculate hash
		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:])

		// Count lines
		lineCount := strings.Count(string(content), "\n") + 1

		// Create or update cache entry
		fileCache := models.RepoFileCache{
			FilePath:      relPath,
			FileType:      ext,
			ContentHash:   hashStr,
			LineCount:     lineCount,
			LastInspected: time.Now(),
			LastModified:  info.ModTime(),
			SizeBytes:     info.Size(),
			IsTracked:     true,
			Metadata:      "{}", // Valid JSON string
		}

		// Upsert
		var existing models.RepoFileCache
		if err := s.db.Where("file_path = ?", relPath).First(&existing).Error; err != nil {
			// Create new record
			if err := s.db.Create(&fileCache).Error; err != nil {
				return fmt.Errorf("failed to cache file %s: %w", relPath, err)
			}
		} else {
			// Update existing record
			fileCache.ID = existing.ID
			fileCache.CreatedAt = existing.CreatedAt
			if err := s.db.Save(&fileCache).Error; err != nil {
				return fmt.Errorf("failed to update cache file %s: %w", relPath, err)
			}
		}

		return nil
	})
}

// GetFileInfo retrieves cached file information
func (s *RepoInspectionService) GetFileInfo(filePath string) (*models.RepoFileCache, error) {
	var fileCache models.RepoFileCache
	if err := s.db.Where("file_path = ?", filePath).First(&fileCache).Error; err != nil {
		return nil, err
	}
	return &fileCache, nil
}

// GetFilesByType retrieves all files of a specific type
func (s *RepoInspectionService) GetFilesByType(fileType string) ([]models.RepoFileCache, error) {
	var files []models.RepoFileCache
	if err := s.db.Where("file_type = ? AND is_tracked = true", fileType).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// GetRecentlyModified retrieves files modified within the last N hours
func (s *RepoInspectionService) GetRecentlyModified(hours int) ([]models.RepoFileCache, error) {
	var files []models.RepoFileCache
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	if err := s.db.Where("last_modified > ? AND is_tracked = true", since).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}
