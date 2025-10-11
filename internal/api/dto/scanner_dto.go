package dto

type FileScanRequest struct {
	RootPath     string   `json:"root_path"`
	SearchTerms  []string `json:"search_terms"`
	Extensions   []string `json:"extensions"`   // e.g., [".md", ".txt", ".go"]
	MaxDepth     int      `json:"max_depth"`     // Max directory depth
	MaxResults   int      `json:"max_results"`   // Max files to return
}

type FileFragment struct {
	Path         string `json:"path"`
	FileName     string `json:"file_name"`
	Size         int64  `json:"size"`
	Modified     string `json:"modified"`
	MatchedTerms []string `json:"matched_terms"`
	Preview      string `json:"preview"` // First 500 chars
}

type FileScanResponse struct {
	FilesFound    []FileFragment `json:"files_found"`
	TotalMatches  int            `json:"total_matches"`
	ScannedPaths  int            `json:"scanned_paths"`
	ImportedCount int            `json:"imported_count"`
}

type ImportFragmentsRequest struct {
	FilePaths []string `json:"file_paths"`
	EventType string   `json:"event_type"` // e.g., "solace_fragment"
}

type ImportFragmentsResponse struct {
	ImportedCount int      `json:"imported_count"`
	Errors        []string `json:"errors"`
}
