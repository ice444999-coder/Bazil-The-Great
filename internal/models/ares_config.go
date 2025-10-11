package models

import "time"

// AresConfig stores ARES identity and core configuration
type AresConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Identity        string    `gorm:"type:text;not null" json:"identity_name"`
	Mission         string    `gorm:"type:text" json:"mission_statement"`
	Initialized     bool      `gorm:"default:false" json:"initialized"`
	GenesisDate     time.Time `gorm:"not null" json:"genesis_date"`
	SolaceImported  bool      `gorm:"default:false" json:"solace_imported"`
	LocalLLMEnabled bool      `gorm:"default:false" json:"enable_local_llm"`
	LLMModel        string    `gorm:"type:text" json:"llm_model"`
	Metadata        JSONB     `gorm:"type:jsonb" json:"metadata"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
