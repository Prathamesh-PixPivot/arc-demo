package models

import (
	"time"

	"github.com/google/uuid"

)

// DPATemplate represents a standard Data Processing Agreement template
type DPATemplate struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID `gorm:"type:uuid;index"`
	Name      string    `gorm:"type:text;not null"`
	Content   string    `gorm:"type:text"` // HTML or Markdown content
	Version   string    `gorm:"type:varchar(20)"`
	IsActive  bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// DPAAgreement represents a signed agreement with a vendor
type DPAAgreement struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID     uuid.UUID `gorm:"type:uuid;index"`
	VendorID     uuid.UUID `gorm:"type:uuid;index"`
	TemplateID   *uuid.UUID `gorm:"type:uuid;index"`
	Status       string    `gorm:"type:varchar(20);default:'draft'"` // draft, sent, signed, expired
	SignedURL    string    `gorm:"type:text"` // Path to signed PDF
	ValidFrom    *time.Time
	ValidUntil   *time.Time
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// AuditChecklist represents a set of questions for an assessment
type AuditChecklist struct {
	ID          string         `json:"id"` // e.g., "dpdpa-v1"
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Categories  []AuditCategory `json:"categories"`
}

type AuditCategory struct {
	Name      string          `json:"name"`
	Questions []AuditQuestion `json:"questions"`
}

type AuditQuestion struct {
	ID               string `json:"id"`
	Text             string `json:"text"`
	ReferenceSection string `json:"reference_section"`
	RiskWeight       int    `json:"risk_weight"` // 1-10
}

// AuditResponse stores the answers for an assessment
type AuditResponse struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	AssessmentID uuid.UUID      `gorm:"type:uuid;index"`
	QuestionID   string         `gorm:"type:varchar(50);index"`
	Response     string         `gorm:"type:varchar(20)"` // yes, no, na
	Comments     string         `gorm:"type:text"`
	EvidenceID   *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
}

