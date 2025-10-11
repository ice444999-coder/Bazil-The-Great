package dto

// EditorFileRequest - Request to read a file
type EditorFileRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

// EditorFileResponse - File content response
type EditorFileResponse struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
	Language string `json:"language"` // File language/extension
	Size     int64  `json:"size"`
}

// EditorSaveRequest - Request to save a file
type EditorSaveRequest struct {
	FilePath string `json:"file_path" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// EditorSaveResponse - Save confirmation
type EditorSaveResponse struct {
	FilePath string `json:"file_path"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// EditorListRequest - Request to list files in directory
type EditorListRequest struct {
	DirectoryPath string `json:"directory_path" binding:"required"`
	Recursive     bool   `json:"recursive"`
	MaxDepth      int    `json:"max_depth"`
}

// EditorFileInfo - File information
type EditorFileInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// EditorListResponse - Directory listing response
type EditorListResponse struct {
	DirectoryPath string           `json:"directory_path"`
	Files         []EditorFileInfo `json:"files"`
	TotalFiles    int              `json:"total_files"`
}

// EditorCreateRequest - Request to create new file/directory
type EditorCreateRequest struct {
	Path  string `json:"path" binding:"required"`
	IsDir bool   `json:"is_dir"`
}

// EditorDeleteRequest - Request to delete file/directory
type EditorDeleteRequest struct {
	Path string `json:"path" binding:"required"`
}

// EditorRenameRequest - Request to rename file/directory
type EditorRenameRequest struct {
	OldPath string `json:"old_path" binding:"required"`
	NewPath string `json:"new_path" binding:"required"`
}
