package models

import (
	"time"

	"github.com/google/uuid"

)

// DataSource represents a connected database for scanning
type DataSource struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID      `gorm:"type:uuid;index"`
	Name        string         `gorm:"type:text"`
	Type        string         `gorm:"type:varchar(20)"` // postgres, mysql, mongodb
	Host        string         `gorm:"type:text"`
	Port        int            `gorm:"default:5432"`
	Database    string         `gorm:"type:text"`
	Username    string         `gorm:"type:text"`
	Password    string         `gorm:"type:text"` // Encrypted
	Description string         `gorm:"type:text"`
	IsActive    bool           `gorm:"default:true"`
	LastScanned *time.Time
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
}

// DiscoveryJob represents a single execution of a scan
type DiscoveryJob struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID     uuid.UUID      `gorm:"type:uuid;index"`
	DataSourceID uuid.UUID      `gorm:"type:uuid;index"`
	Status       string         `gorm:"type:varchar(20);default:'pending'"` // pending, running, completed, failed
	StartTime    *time.Time
	EndTime      *time.Time
	TotalTables  int            `gorm:"default:0"`
	TotalColumns int            `gorm:"default:0"`
	PIIFound     int            `gorm:"default:0"`
	ErrorMessage string         `gorm:"type:text"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
}

// DiscoveryResult represents a specific PII finding
type DiscoveryResult struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey"`
	JobID          uuid.UUID      `gorm:"type:uuid;index"`
	TenantID       uuid.UUID      `gorm:"type:uuid;index"`
	DataSourceID   uuid.UUID      `gorm:"type:uuid;index"`
	TableName      string         `gorm:"type:text"`
	ColumnName     string         `gorm:"type:text"`
	DataType       string         `gorm:"type:text"`
	PIIType        string         `gorm:"type:varchar(50)"` // aadhaar, pan, email, phone, etc.
	Confidence     float64        `gorm:"type:decimal(5,2)"`
	SampleData     string         `gorm:"type:text"` // Masked
	IsVerified     bool           `gorm:"default:false"`
	Classification string         `gorm:"type:varchar(20)"` // public, internal, confidential, restricted
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
}

// DataClassificationStats for dashboard
type DataClassificationStats struct {
	TotalDataSources int            `json:"total_data_sources"`
	TotalPIIColumns  int            `json:"total_pii_columns"`
	ByPIIType        map[string]int `json:"by_pii_type"`
	BySeverity       map[string]int `json:"by_severity"`
}

