package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// BreachImpactAssessment represents risk assessment for a breach
type BreachImpactAssessment struct {
	ID                              uuid.UUID      `gorm:"type:uuid;primaryKey"`
	BreachID                        uuid.UUID      `gorm:"type:uuid;index;not null"`
	TenantID                        uuid.UUID      `gorm:"type:uuid;index"`
	LikelihoodOfHarm                string         `gorm:"type:varchar(20)"` // high, medium, low
	ImpactLevel                     string         `gorm:"type:varchar(20)"` // severe, significant, minor
	RiskToRightsLevel               string         `gorm:"type:varchar(20)"` // high, medium, low
	DataCategoriesAffected          datatypes.JSON `gorm:"type:jsonb"`       // ["PII", "financial", "health", "biometric"]
	NumberOfAffected                int            `gorm:"default:0"`
	GeographicScope                 datatypes.JSON `gorm:"type:jsonb"` // ["India", "EU", "US"]
	RequiresAuthorityNotification   bool           `gorm:"default:false"`
	RequiresDataSubjectNotification bool           `gorm:"default:false"`
	AssessmentNotes                 string         `gorm:"type:text"`
	AssessedBy                      uuid.UUID      `gorm:"type:uuid"`
	AssessedAt                      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt                       time.Time      `gorm:"autoUpdateTime"`
}

// BreachStakeholder tracks notifications sent to various parties
type BreachStakeholder struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	BreachID           uuid.UUID `gorm:"type:uuid;index;not null"`
	TenantID           uuid.UUID `gorm:"type:uuid;index"`
	StakeholderType    string    `gorm:"type:varchar(50)"` // dpb, affected_individual, media, business_associate, regulator
	ContactName        string    `gorm:"type:varchar(255)"`
	ContactEmail       string    `gorm:"type:varchar(255)"`
	ContactPhone       string    `gorm:"type:varchar(50)"`
	NotificationSent   bool      `gorm:"default:false"`
	NotifiedAt         *time.Time
	NotificationMethod string `gorm:"type:varchar(50)"` // email, postal, sms, portal
	NotificationStatus string `gorm:"type:varchar(50)"` // pending, sent, delivered, failed, acknowledged
	AcknowledgedAt     *time.Time
	AcknowledgementRef string    `gorm:"type:text"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

// BreachCommunication tracks all communications related to a breach
type BreachCommunication struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	BreachID          uuid.UUID `gorm:"type:uuid;index;not null"`
	TenantID          uuid.UUID `gorm:"type:uuid;index"`
	CommunicationType string    `gorm:"type:varchar(50)"` // initial_notification, update, resolution, follow_up
	Recipient         string    `gorm:"type:varchar(255)"`
	RecipientType     string    `gorm:"type:varchar(50)"` // dpb, data_principal, media
	Subject           string    `gorm:"type:text"`
	Content           string    `gorm:"type:text"`
	TemplateName      string    `gorm:"type:varchar(100)"`
	SendMethod        string    `gorm:"type:varchar(50)"` // email, sms, postal, portal
	SentAt            *time.Time
	DeliveredAt       *time.Time
	Status            string    `gorm:"type:varchar(50)"` // draft, queued, sent, delivered, failed
	ErrorMessage      *string   `gorm:"type:text"`
	CreatedBy         uuid.UUID `gorm:"type:uuid"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

// BreachWorkflowStage tracks the workflow progression with approval gates
type BreachWorkflowStage struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	BreachID         uuid.UUID  `gorm:"type:uuid;index;not null"`
	TenantID         uuid.UUID  `gorm:"type:uuid;index"`
	Stage            string     `gorm:"type:varchar(50)"` // detection, assessment, containment, verification, notification, resolution
	Status           string     `gorm:"type:varchar(50)"` // pending, in_progress, completed, approved, rejected, overdue
	RequiresApproval bool       `gorm:"default:false"`
	ApprovedBy       *uuid.UUID `gorm:"type:uuid"`
	ApprovedAt       *time.Time
	RejectedBy       *uuid.UUID `gorm:"type:uuid"`
	RejectedAt       *time.Time
	RejectionReason  *string `gorm:"type:text"`
	DueDate          *time.Time
	CompletedAt      *time.Time
	AssignedTo       *uuid.UUID `gorm:"type:uuid"`
	Notes            string     `gorm:"type:text"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime"`
}

// BreachEvidence stores evidence and documentation related to a breach
type BreachEvidence struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	BreachID       uuid.UUID `gorm:"type:uuid;index;not null"`
	TenantID       uuid.UUID `gorm:"type:uuid;index"`
	EvidenceType   string    `gorm:"type:varchar(50)"` // logs, screenshots, forensic_report, email, document, video
	Title          string    `gorm:"type:varchar(255)"`
	Description    string    `gorm:"type:text"`
	FilePath       string    `gorm:"type:text"`
	FileSize       int64     `gorm:"default:0"`
	FileHash       string    `gorm:"type:varchar(255)"` // For integrity verification
	CollectedBy    uuid.UUID `gorm:"type:uuid"`
	CollectedAt    time.Time
	ChainOfCustody string    `gorm:"type:text"` // JSON array of custody transfers
	IsConfidential bool      `gorm:"default:false"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

// BreachTimeline tracks all significant events in breach lifecycle
type BreachTimeline struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	BreachID    uuid.UUID      `gorm:"type:uuid;index;not null"`
	TenantID    uuid.UUID      `gorm:"type:uuid;index"`
	EventType   string         `gorm:"type:varchar(50)"` // detection, containment, notification_sent, status_change, etc.
	Description string         `gorm:"type:text"`
	PerformedBy *uuid.UUID     `gorm:"type:uuid"`
	EventData   datatypes.JSON `gorm:"type:jsonb"` // Additional context
	OccurredAt  time.Time      `gorm:"autoCreateTime"`
}

// BreachNotificationTemplate for DPDP-compliant templates
type BreachNotificationTemplate struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID         *uuid.UUID `gorm:"type:uuid;index"` // Null for system templates
	TemplateName     string     `gorm:"type:varchar(100);not null"`
	RecipientType    string     `gorm:"type:varchar(50)"` // dpb, data_principal
	TemplateType     string     `gorm:"type:varchar(50)"` // email, sms, letter
	Subject          string     `gorm:"type:text"`
	Body             string     `gorm:"type:text"` // Supports template variables like {{breach_title}}
	Language         string     `gorm:"type:varchar(10);default:'en'"`
	IsActive         bool       `gorm:"default:true"`
	IsSystemTemplate bool       `gorm:"default:false"` // Cannot be modified
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime"`
}
