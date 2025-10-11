package services

import (
	"ares_api/internal/api/dto"
	repo "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ScannerServiceImpl struct {
	MemoryRepo repo.MemoryRepository
}

func NewScannerService(memoryRepo repo.MemoryRepository) *ScannerServiceImpl {
	return &ScannerServiceImpl{
		MemoryRepo: memoryRepo,
	}
}

// ScanFiles recursively scans filesystem for files containing search terms
func (s *ScannerServiceImpl) ScanFiles(req dto.FileScanRequest) (dto.FileScanResponse, error) {
	var filesFound []dto.FileFragment
	scannedPaths := 0

	// Default values
	if req.MaxDepth == 0 {
		req.MaxDepth = 5 // Limit depth for safety
	}
	if req.MaxResults == 0 {
		req.MaxResults = 100
	}
	if len(req.SearchTerms) == 0 {
		req.SearchTerms = []string{"Solace", "ARES", "consciousness", "recursive"}
	}
	if len(req.Extensions) == 0 {
		req.Extensions = []string{".md", ".txt", ".go", ".json", ".log"}
	}

	err := filepath.WalkDir(req.RootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Check depth limit
		relPath, _ := filepath.Rel(req.RootPath, path)
		depth := strings.Count(relPath, string(os.PathSeparator))
		if depth > req.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories and system folders
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" ||
			   name == "bin" || name == "obj" || name == "$RECYCLE.BIN" {
				return filepath.SkipDir
			}
			scannedPaths++
			return nil
		}

		// Check if we've hit max results
		if len(filesFound) >= req.MaxResults {
			return filepath.SkipAll
		}

		// Check file extension
		hasValidExt := false
		for _, ext := range req.Extensions {
			if strings.HasSuffix(strings.ToLower(d.Name()), strings.ToLower(ext)) {
				hasValidExt = true
				break
			}
		}
		if !hasValidExt {
			return nil
		}

		scannedPaths++

		// Check if file contains search terms (in filename or content)
		matchedTerms := []string{}
		lowerName := strings.ToLower(d.Name())
		lowerPath := strings.ToLower(path)

		// Check filename
		for _, term := range req.SearchTerms {
			lowerTerm := strings.ToLower(term)
			if strings.Contains(lowerName, lowerTerm) || strings.Contains(lowerPath, lowerTerm) {
				matchedTerms = append(matchedTerms, term)
			}
		}

		// Check content (first 10KB for performance)
		if len(matchedTerms) == 0 {
			content, err := os.ReadFile(path)
			if err == nil && len(content) > 0 {
				// Only check first 10KB
				searchContent := string(content)
				if len(searchContent) > 10240 {
					searchContent = searchContent[:10240]
				}
				lowerContent := strings.ToLower(searchContent)

				for _, term := range req.SearchTerms {
					if strings.Contains(lowerContent, strings.ToLower(term)) {
						matchedTerms = append(matchedTerms, term)
					}
				}
			}
		}

		// If matches found, add to results
		if len(matchedTerms) > 0 {
			info, _ := d.Info()
			preview := ""
			if content, err := os.ReadFile(path); err == nil {
				preview = string(content)
				if len(preview) > 500 {
					preview = preview[:500] + "..."
				}
			}

			filesFound = append(filesFound, dto.FileFragment{
				Path:         path,
				FileName:     d.Name(),
				Size:         info.Size(),
				Modified:     info.ModTime().Format(time.RFC3339),
				MatchedTerms: matchedTerms,
				Preview:      preview,
			})
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return dto.FileScanResponse{}, fmt.Errorf("scan error: %w", err)
	}

	return dto.FileScanResponse{
		FilesFound:   filesFound,
		TotalMatches: len(filesFound),
		ScannedPaths: scannedPaths,
	}, nil
}

// ImportFragments imports file fragments into memory_snapshots
func (s *ScannerServiceImpl) ImportFragments(userID uint, req dto.ImportFragmentsRequest) (dto.ImportFragmentsResponse, error) {
	imported := 0
	errors := []string{}
	sessionID := uuid.New()

	for _, filePath := range req.FilePaths {
		content, err := os.ReadFile(filePath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to read %s: %v", filePath, err))
			continue
		}

		// Create memory snapshot
		snapshot := &models.MemorySnapshot{
			Timestamp: time.Now(),
			EventType: req.EventType,
			Payload: models.JSONB{
				"file_path": filePath,
				"content":   string(content),
				"imported_at": time.Now().Unix(),
				"source":    "file_scanner",
			},
			UserID:    userID,
			SessionID: &sessionID,
		}

		if err := s.MemoryRepo.SaveSnapshot(snapshot); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to save %s: %v", filePath, err))
			continue
		}

		imported++
	}

	return dto.ImportFragmentsResponse{
		ImportedCount: imported,
		Errors:        errors,
	}, nil
}

// ImportSolaceData scans C:\ProgramData\Solace\State and imports all files
func (s *ScannerServiceImpl) ImportSolaceData(userID uint) (dto.ImportFragmentsResponse, error) {
	solacePath := `C:\ProgramData\Solace\State`

	// Check if directory exists
	if _, err := os.Stat(solacePath); os.IsNotExist(err) {
		return dto.ImportFragmentsResponse{
			ImportedCount: 0,
			Errors:        []string{"Solace directory not found at " + solacePath},
		}, nil
	}

	// Scan for all files in Solace directory
	var filePaths []string
	err := filepath.WalkDir(solacePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Accept all files (md, txt, json, log, etc.)
		filePaths = append(filePaths, path)
		return nil
	})

	if err != nil {
		return dto.ImportFragmentsResponse{
			ImportedCount: 0,
			Errors:        []string{fmt.Sprintf("Scan error: %v", err)},
		}, err
	}

	// Import all found files with special Solace tag
	importReq := dto.ImportFragmentsRequest{
		FilePaths: filePaths,
		EventType: "solace_delta_3_1",
	}

	return s.ImportFragments(userID, importReq)
}
