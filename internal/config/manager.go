/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ServiceConfig represents a configuration entry
type ServiceConfig struct {
	ID          int             `json:"id" gorm:"primaryKey"`
	ServiceName string          `json:"service_name" gorm:"not null"`
	ConfigKey   string          `json:"config_key" gorm:"not null"`
	ConfigValue json.RawMessage `json:"config_value" gorm:"type:jsonb;not null"`
	Description string          `json:"description"`
	IsEncrypted bool            `json:"is_encrypted" gorm:"default:false"`
	LastUpdated time.Time       `json:"last_updated" gorm:"column:last_updated"`
	UpdatedBy   string          `json:"updated_by"`
	Version     int             `json:"version" gorm:"default:1"`
	CreatedAt   time.Time       `json:"created_at"`
}

// TableName specifies the table name for GORM
func (ServiceConfig) TableName() string {
	return "service_config"
}

// ConfigHistory tracks config changes
type ConfigHistory struct {
	ID           int             `json:"id" gorm:"primaryKey"`
	ServiceName  string          `json:"service_name"`
	ConfigKey    string          `json:"config_key"`
	OldValue     json.RawMessage `json:"old_value" gorm:"type:jsonb"`
	NewValue     json.RawMessage `json:"new_value" gorm:"type:jsonb"`
	ChangedBy    string          `json:"changed_by"`
	ChangeReason string          `json:"change_reason"`
	ChangedAt    time.Time       `json:"changed_at"`
}

func (ConfigHistory) TableName() string {
	return "service_config_history"
}

// Manager handles dynamic configuration with hot-reload
type Manager struct {
	db          *gorm.DB
	serviceName string
	cache       map[string]interface{}
	mu          sync.RWMutex
	stopCh      chan struct{}
}

// GetServiceName returns the service name
func (m *Manager) GetServiceName() string {
	return m.serviceName
}

// NewManager creates a new config manager
func NewManager(db *gorm.DB, serviceName string) *Manager {
	m := &Manager{
		db:          db,
		serviceName: serviceName,
		cache:       make(map[string]interface{}),
		stopCh:      make(chan struct{}),
	}

	// Initial load
	if err := m.Reload(); err != nil {
		log.Printf("[CONFIG] Warning: Initial config load failed: %v", err)
	}

	// Start hot-reload goroutine
	go m.startHotReload()

	return m
}

// Reload reloads configuration from database
func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var configs []ServiceConfig
	err := m.db.Where("service_name = ?", m.serviceName).Find(&configs).Error
	if err != nil {
		return fmt.Errorf("failed to load configs: %w", err)
	}

	// Update cache
	newCache := make(map[string]interface{})
	for _, cfg := range configs {
		var value interface{}
		if err := json.Unmarshal(cfg.ConfigValue, &value); err != nil {
			log.Printf("[CONFIG] Warning: Failed to unmarshal config %s: %v", cfg.ConfigKey, err)
			continue
		}
		newCache[cfg.ConfigKey] = value
	}

	// Only log if configs actually changed
	configsChanged := len(newCache) != len(m.cache)
	if !configsChanged {
		// Check if any values changed
		for key, newVal := range newCache {
			oldVal, exists := m.cache[key]
			if !exists || fmt.Sprintf("%v", newVal) != fmt.Sprintf("%v", oldVal) {
				configsChanged = true
				break
			}
		}
	}

	m.cache = newCache

	if configsChanged {
		log.Printf("[CONFIG] ✅ Reloaded %d configs for %s (changes detected)", len(newCache), m.serviceName)
	}
	// No log spam if nothing changed
	return nil
}

// Get retrieves a config value
func (m *Manager) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.cache[key]
	return val, ok
}

// GetString retrieves a string config value with default
func (m *Manager) GetString(key, defaultValue string) string {
	val, ok := m.Get(key)
	if !ok {
		return defaultValue
	}
	if str, ok := val.(string); ok {
		return str
	}
	return defaultValue
}

// GetInt retrieves an int config value with default
func (m *Manager) GetInt(key string, defaultValue int) int {
	val, ok := m.Get(key)
	if !ok {
		return defaultValue
	}
	// JSON numbers are float64
	if f, ok := val.(float64); ok {
		return int(f)
	}
	return defaultValue
}

// GetBool retrieves a bool config value with default
func (m *Manager) GetBool(key string, defaultValue bool) bool {
	val, ok := m.Get(key)
	if !ok {
		return defaultValue
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

// Set updates a config value (with history tracking)
func (m *Manager) Set(key string, value interface{}, updatedBy, reason string) error {
	// Marshal new value
	newValueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Get old value for history
	var oldConfig ServiceConfig
	err = m.db.Where("service_name = ? AND config_key = ?", m.serviceName, key).First(&oldConfig).Error
	oldExists := err == nil

	// Update or insert config
	config := ServiceConfig{
		ServiceName: m.serviceName,
		ConfigKey:   key,
		ConfigValue: newValueBytes,
		UpdatedBy:   updatedBy,
		LastUpdated: time.Now(),
	}

	if oldExists {
		// Update existing
		config.ID = oldConfig.ID
		config.Version = oldConfig.Version + 1
		err = m.db.Save(&config).Error
	} else {
		// Insert new
		err = m.db.Create(&config).Error
	}

	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Record history
	history := ConfigHistory{
		ServiceName:  m.serviceName,
		ConfigKey:    key,
		NewValue:     newValueBytes,
		ChangedBy:    updatedBy,
		ChangeReason: reason,
		ChangedAt:    time.Now(),
	}
	if oldExists {
		history.OldValue = oldConfig.ConfigValue
	}
	m.db.Create(&history)

	// Update cache
	m.mu.Lock()
	m.cache[key] = value
	m.mu.Unlock()

	log.Printf("[CONFIG] ✅ Updated %s.%s (version %d)", m.serviceName, key, config.Version)
	return nil
}

// GetAll returns all configs for this service
func (m *Manager) GetAll() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]interface{}, len(m.cache))
	for k, v := range m.cache {
		result[k] = v
	}
	return result
}

// startHotReload starts the hot-reload goroutine
func (m *Manager) startHotReload() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.Reload(); err != nil {
				log.Printf("[CONFIG] ⚠️  Hot-reload failed: %v", err)
			}
		case <-m.stopCh:
			log.Printf("[CONFIG] Stopping hot-reload for %s", m.serviceName)
			return
		}
	}
}

// Close stops the hot-reload goroutine
func (m *Manager) Close() {
	close(m.stopCh)
}

// GetHistory returns config change history
func (m *Manager) GetHistory(key string, limit int) ([]ConfigHistory, error) {
	var history []ConfigHistory
	query := m.db.Where("service_name = ?", m.serviceName)
	if key != "" {
		query = query.Where("config_key = ?", key)
	}
	err := query.Order("changed_at DESC").Limit(limit).Find(&history).Error
	return history, err
}
