package controllers

import (
	"ares_api/internal/common"
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BackupController struct {
	DB *gorm.DB
}

func NewBackupController(db *gorm.DB) *BackupController {
	return &BackupController{DB: db}
}

// @Summary Export full database backup
// @Description Exports all tables to JSON and creates a downloadable zip file
// @Tags Backup
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /backup/export [get]
func (bc *BackupController) Export(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Create backup directory
	backupDir := "C:\\ARES_Workspace\\backups"
	os.MkdirAll(backupDir, 0755)

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("ares_backup_%d_%s", userID, timestamp)
	zipPath := filepath.Join(backupDir, backupName+".zip")

	// Create zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": "Failed to create backup file"})
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Export user data
	tables := []string{
		"users", "chats", "trades", "settings", "ledgers", "balances",
		"memory_snapshots", "chat_messages", "conversation_imports",
		"file_scan_results", "ares_configs",
	}

	for _, table := range tables {
		if err := bc.exportTable(zipWriter, table, userID); err != nil {
			fmt.Printf("Warning: Failed to export table %s: %v\n", table, err)
		}
	}

	// Add metadata file
	metadata := map[string]interface{}{
		"user_id":     userID,
		"timestamp":   timestamp,
		"version":     "1.0",
		"tables":      tables,
		"backup_type": "full",
	}

	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	metadataFile, _ := zipWriter.Create("metadata.json")
	metadataFile.Write(metadataJSON)

	common.JSON(c, http.StatusOK, gin.H{
		"message":     "Backup created successfully",
		"backup_file": backupName + ".zip",
		"path":        zipPath,
		"size_bytes":  getFileSize(zipPath),
	})
}

func (bc *BackupController) exportTable(zipWriter *zip.Writer, table string, userID uint) error {
	var data []map[string]interface{}

	// Query table with user filter where applicable
	query := bc.DB.Table(table)
	if table != "ares_configs" { // Skip user filter for config table
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&data).Error; err != nil {
		return err
	}

	// Write to zip
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s.json", table)
	fileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}

	_, err = fileWriter.Write(jsonData)
	return err
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// @Summary Import database backup
// @Description Restore data from a backup zip file
// @Tags Backup
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Backup zip file"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /backup/import [post]
func (bc *BackupController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Save uploaded file
	uploadPath := filepath.Join("C:\\ARES_Workspace\\backups", "temp_"+file.Filename)
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer os.Remove(uploadPath)

	// Open zip
	reader, err := zip.OpenReader(uploadPath)
	if err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": "Invalid backup file"})
		return
	}
	defer reader.Close()

	imported := 0
	for _, f := range reader.File {
		if filepath.Ext(f.Name) == ".json" && f.Name != "metadata.json" {
			if err := bc.importTable(f); err != nil {
				fmt.Printf("Warning: Failed to import %s: %v\n", f.Name, err)
			} else {
				imported++
			}
		}
	}

	common.JSON(c, http.StatusOK, gin.H{
		"message":        "Backup imported successfully",
		"tables_imported": imported,
	})
}

func (bc *BackupController) importTable(f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	tableName := f.Name[:len(f.Name)-5] // Remove .json extension

	// Insert records
	for _, record := range records {
		bc.DB.Table(tableName).Create(&record)
	}

	return nil
}
