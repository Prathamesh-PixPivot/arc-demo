package models

import (
	"time"

	"github.com/google/uuid"
)

// IssuedLicense represents a license issued to a customer
type IssuedLicense struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	CustomerID   uuid.UUID `gorm:"index"`
	CustomerName string    `gorm:"type:text"`
	PlanTier     string    `gorm:"type:varchar(50)"` // e.g., ENTERPRISE, PRO
	Type         string    `gorm:"type:varchar(50)"` // e.g., SAAS, ON_PREM
	IssuedAt     time.Time
	ExpiresAt    *time.Time
	LicenseKey   string `gorm:"type:text"` // The full signed JWT
	IsActive     bool   `gorm:"default:true"`
	RevokedAt    *time.Time
	Metadata     string    `gorm:"type:jsonb"` // JSON string of limits/features
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
