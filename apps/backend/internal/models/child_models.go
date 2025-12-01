package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ChildProfile represents a child user linked to a parent
type ChildProfile struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	ParentID     uuid.UUID `gorm:"type:uuid;index;not null"` // The DataPrincipal ID of the parent
	TenantID     uuid.UUID `gorm:"type:uuid;index;not null"`
	Name         string    `gorm:"type:text;not null"`
	DateOfBirth  time.Time `gorm:"type:date;not null"`
	Gender       string    `gorm:"type:varchar(20)"`
	Relationship string    `gorm:"type:varchar(50)"` // e.g., "Parent", "Guardian"
	IsActive     bool      `gorm:"default:true"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`

	// Parent reference (optional, for GORM)
	// Parent DataPrincipal `gorm:"foreignKey:ParentID"`
}

func (ChildProfile) TableName() string {
	return "child_profiles"
}

// ParentalConsentRequest represents a request for a parent to approve an action for a child
type ParentalConsentRequest struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ChildID        uuid.UUID      `gorm:"type:uuid;index;not null"`
	ParentID       uuid.UUID      `gorm:"type:uuid;index;not null"`
	TenantID       uuid.UUID      `gorm:"type:uuid;index;not null"`
	RequestType    string         `gorm:"type:varchar(50);not null"`          // e.g., "app_access", "data_sharing"
	ResourceName   string         `gorm:"type:text"`                          // e.g., "Math Learning App"
	PurposeID      *uuid.UUID     `gorm:"type:uuid"`                          // If related to a specific purpose
	Status         string         `gorm:"type:varchar(20);default:'pending'"` // pending, approved, rejected, expired
	RequestDetails datatypes.JSON `gorm:"type:jsonb"`
	ExpiresAt      time.Time
	ApprovedAt     *time.Time
	RejectedAt     *time.Time
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func (ParentalConsentRequest) TableName() string {
	return "parental_consent_requests"
}
