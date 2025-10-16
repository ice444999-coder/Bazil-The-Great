package registry

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// ServiceInfo represents a service in the registry
type ServiceInfo struct {
	ID            int                    `json:"id" gorm:"primaryKey"`
	Name          string                 `json:"name" gorm:"unique;not null"`
	Version       string                 `json:"version" gorm:"not null"`
	Status        string                 `json:"status" gorm:"not null;default:offline"`
	Port          int                    `json:"port"`
	HealthURL     string                 `json:"health_url" gorm:"column:health_url"`
	LastHeartbeat *time.Time             `json:"last_heartbeat" gorm:"column:last_heartbeat"`
	Metadata      map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt     time.Time              `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time              `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for GORM
func (ServiceInfo) TableName() string {
	return "service_registry"
}

// RegisterService registers or updates a service in the registry
func RegisterService(db *gorm.DB, name, version string, port int, healthURL string) error {
	// Use raw SQL for upsert with PostgreSQL-specific syntax
	result := db.Exec(`
		INSERT INTO service_registry (name, version, status, port, health_url, last_heartbeat, created_at, updated_at)
		VALUES (?, ?, 'online', ?, ?, NOW(), NOW(), NOW())
		ON CONFLICT (name) 
		DO UPDATE SET 
			version = EXCLUDED.version,
			status = 'online',
			port = EXCLUDED.port,
			health_url = EXCLUDED.health_url,
			last_heartbeat = NOW(),
			updated_at = NOW()
	`, name, version, port, healthURL)

	if result.Error != nil {
		return result.Error
	}
	log.Printf("[REGISTRY] ✅ Registered %s v%s on port %d", name, version, port)
	return nil
}

// ServiceHeartbeat sends periodic heartbeats to update service status
func ServiceHeartbeat(db *gorm.DB, serviceName string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		result := db.Exec(`
			UPDATE service_registry 
			SET last_heartbeat = NOW(), status = 'online'
			WHERE name = ?
		`, serviceName)

		if result.Error != nil {
			log.Printf("[REGISTRY][ERROR] Heartbeat failed for %s: %v", serviceName, result.Error)
		}
	}
}

// GetAllServices retrieves all registered services
func GetAllServices(db *gorm.DB) ([]ServiceInfo, error) {
	var services []ServiceInfo
	result := db.Order("name").Find(&services)
	if result.Error != nil {
		return nil, result.Error
	}
	return services, nil
}

// GetService retrieves a specific service by name
func GetService(db *gorm.DB, name string) (*ServiceInfo, error) {
	var service ServiceInfo
	result := db.Where("name = ?", name).First(&service)
	if result.Error != nil {
		return nil, result.Error
	}
	return &service, nil
}

// MarkServiceOffline marks a service as offline
func MarkServiceOffline(db *gorm.DB, serviceName string) error {
	result := db.Model(&ServiceInfo{}).
		Where("name = ?", serviceName).
		Updates(map[string]interface{}{
			"status":     "offline",
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	log.Printf("[REGISTRY] Service %s marked as offline", serviceName)
	return nil
}
