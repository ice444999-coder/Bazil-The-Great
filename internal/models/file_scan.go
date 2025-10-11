package models

import "time"

// FileScanResult stores results from file scanner
type FileScanResult struct {
	ID           uint      `gorm:"primaryKey"`
	ScanPath     string    `gorm:"type:text;not null"`
	FoundFiles   []string  `gorm:"type:text[]"`
	ScanDuration int64     `gorm:"default:0"` // milliseconds
	FilesImported int      `gorm:"default:0"`
	ScannedAt    time.Time `gorm:"autoCreateTime;not null;index"`
	UserID       uint      `gorm:"index"`
}
