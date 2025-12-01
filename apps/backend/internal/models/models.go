package models

import (
	"pixpivot/arc/internal/claims"
	"pixpivot/arc/internal/dto"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Permission represents an atomic action a user can perform.
type Permission struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"type:varchar(100);uniqueIndex;not null"` // e.g., "users:create", "roles:manage"
	Description string `gorm:"type:text"`
}

// Role is a collection of permissions that can be assigned to a user.
type Role struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID        `gorm:"type:uuid;index:idx_tenant_role_name,unique"`
	Name        string           `gorm:"type:varchar(100);index:idx_tenant_role_name,unique"`
	Description string           `gorm:"type:text"`
	Permissions []*Permission    `gorm:"many2many:role_permissions;"`
	Users       []*FiduciaryUser `gorm:"many2many:fiduciary_user_roles;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// -------------------------------
// Admin and Organizational Models
// -------------------------------
type FiduciaryUser struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID            uuid.UUID `gorm:"type:uuid;index"`
	Email               string    `gorm:"type:text;uniqueIndex"`
	Phone               string    `gorm:"type:text;uniqueIndex"`
	Name                string    `gorm:"type:text"`
	PasswordHash        string    `gorm:"type:text"`
	IsVerified          bool      `gorm:"default:false"`
	VerificationToken   string    `gorm:"type:text;index"`
	VerificationExpiry  time.Time
	PasswordResetToken  string `gorm:"type:text"`
	PasswordResetExpiry time.Time
	AuthProvider        string    `gorm:"type:varchar(50);default:'email'"` // email, google, microsoft
	ProviderID          string    `gorm:"type:text;index"`                  // External ID from provider
	CreatedAt           time.Time `gorm:"autoCreateTime"`
	LastSeen            time.Time `gorm:"autoUpdateTime"`
	Roles               []*Role   `gorm:"many2many:fiduciary_user_roles;"`

	// DEPRECATED: These will be replaced by the new RBAC system.
	// Kept for a transitional period if needed, but new logic should not use them.
	Role                  string `gorm:"type:varchar(20);default:'viewer'"` // E.g., superadmin, admin, dpo, viewer
	CanManageConsent      bool   `gorm:"default:false"`
	CanManageGrievance    bool   `gorm:"default:false"`
	CanManagePurposes     bool   `gorm:"default:false"`
	CanManageAuditLogs    bool   `gorm:"default:false"`
	CanManageConsentForms bool   `gorm:"default:false"`

	// Superadmin
	IsSuperAdmin bool `gorm:"default:false"`
}

// User represents a generic user (alias for FiduciaryUser for compatibility)
type User = FiduciaryUser

// HasPermission checks if a user has a specific permission through their roles.
func (u *FiduciaryUser) HasPermission(permissionName string) bool {
	for _, role := range u.Roles {
		for _, p := range role.Permissions {
			if p.Name == permissionName {
				return true
			}
		}
	}
	return false
}

// ToClaims converts a FiduciaryUser to FiduciaryClaims for JWT generation.
func (u *FiduciaryUser) ToClaims() *claims.FiduciaryClaims {
	permissions := make(map[string]bool)
	roleNames := []string{}
	for _, role := range u.Roles {
		roleNames = append(roleNames, role.Name)
		for _, p := range role.Permissions {
			permissions[p.Name] = true
		}
	}

	return &claims.FiduciaryClaims{
		FiduciaryID:  u.ID.String(),
		TenantID:     u.TenantID.String(),
		Roles:        roleNames,
		Permissions:  permissions,
		Role:         u.Role, // Deprecated
		Type:         "fiduciary",
		IsSuperAdmin: u.IsSuperAdmin,
	}
}

type OrganizationEntity struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID `gorm:"type:uuid;index"`
	Name        string    `gorm:"type:text"`
	TaxID       string    `gorm:"type:text"`
	Website     string    `gorm:"type:text"`
	Email       string    `gorm:"type:text;uniqueIndex"`
	Phone       string    `gorm:"type:text"`
	CompanySize string    `gorm:"type:text"`
	Industry    string    `gorm:"type:text"`
	Address     string    `gorm:"type:text"`
	Country     string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// -------------------------------
// Data Principal (End-User) Models
// -------------------------------

// DataPrincipal represents the end-user (the data subject).
type DataPrincipal struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID           uuid.UUID `gorm:"type:uuid;index"` // Link to the DF's tenant
	ExternalID         string    `gorm:"type:text;index"` // ID from the fiduciary's system
	Email              string    `gorm:"type:text;uniqueIndex"`
	Phone              string    `gorm:"type:text;index"`
	FirstName          string    `gorm:"type:text"`
	LastName           string    `gorm:"type:text"`
	Age                int       `gorm:"type:int"`
	Location           string    `gorm:"type:text"`
	IsVerified         bool      `gorm:"default:false"`
	VerificationToken  string    `gorm:"type:text;index"`
	VerificationExpiry time.Time

	// Password-related fields for user authentication
	PasswordHash        string `gorm:"type:text"`
	PasswordResetToken  string `gorm:"type:text"`
	PasswordResetExpiry time.Time

	// Guardian-related fields for minors
	GuardianEmail              string `gorm:"type:text;index"`
	IsGuardianVerified         bool   `gorm:"default:false"`
	GuardianVerificationToken  string `gorm:"type:text;index"`
	GuardianVerificationExpiry time.Time

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// EncryptedDataPrincipal represents a DataPrincipal with encrypted sensitive fields
type EncryptedDataPrincipal struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID           uuid.UUID `gorm:"type:uuid;index"` // Link to the DF's tenant
	ExternalID         string    `gorm:"type:text;index"` // ID from the fiduciary's system
	Email              string    `gorm:"type:text;uniqueIndex"`
	Phone              string    `gorm:"type:text;index"`
	FirstName          string    `gorm:"type:text"`
	LastName           string    `gorm:"type:text"`
	Age                int       `gorm:"type:int"`
	Location           string    `gorm:"type:text"`
	IsVerified         bool      `gorm:"default:false"`
	VerificationToken  string    `gorm:"type:text;index"`
	VerificationExpiry time.Time

	// Password-related fields for user authentication
	PasswordHash        string `gorm:"type:text"`
	PasswordResetToken  string `gorm:"type:text"`
	PasswordResetExpiry time.Time

	// Guardian-related fields for minors
	GuardianEmail              string `gorm:"type:text;index"`
	IsGuardianVerified         bool   `gorm:"default:false"`
	GuardianVerificationToken  string `gorm:"type:text;index"`
	GuardianVerificationExpiry time.Time

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type UserTenantLink struct {
	ID             uuid.UUID `gorm:"primaryKey"`
	UserID         uuid.UUID `gorm:"type:uuid;index"`
	TenantID       uuid.UUID `gorm:"type:uuid;index"`
	TenantName     string
	FirstGrantedAt time.Time
	LastUpdatedAt  time.Time
}

// -------------------------------
// Consent & Purpose Models
// -------------------------------
type Purpose struct {
	ID                uuid.UUID `gorm:"primaryKey"`
	Name              string
	Description       string
	Required          bool
	Active            bool
	ReviewCycleMonths int
	LegalBasis        string
	Version           string
	Language          string
	TenantID          uuid.UUID      `gorm:"index"`
	Vendors           pq.StringArray `gorm:"type:text[]" json:"vendors"`
	IsThirdParty      bool           `gorm:"default:false"`
	// New hierarchy and compliance fields
	ParentPurposeID     *uuid.UUID `gorm:"type:uuid;index"` // For hierarchy
	TemplateID          *uuid.UUID `gorm:"type:uuid;index"` // Link to template
	RetentionPeriodDays int        `gorm:"default:0"`
	ComplianceStatus    string     `gorm:"type:varchar(20);default:'needs_review'"` // compliant, needs_review, non_compliant
	LastComplianceCheck *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastUsedAt          *time.Time
	TotalGranted        int
	TotalRevoked        int
}

func (Purpose) TableName() string {
	return "purposes"
}

// PurposeTemplate represents a regulatory-compliant purpose template
type PurposeTemplate struct {
	ID                     uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name                   string         `gorm:"type:text;not null"`
	Description            string         `gorm:"type:text;not null"`
	Category               string         `gorm:"type:varchar(20);not null"` // marketing, analytics, functional, necessary
	LegalBasis             string         `gorm:"type:varchar(30);not null"` // consent, contract, legal_obligation, legitimate_interest
	RegulatoryFramework    string         `gorm:"type:varchar(10);not null"` // dpdp, gdpr, ccpa
	RequiredDataObjects    pq.StringArray `gorm:"type:text[]" json:"required_data_objects"`
	SuggestedRetentionDays int            `gorm:"default:365"`
	ComplianceNotes        string         `gorm:"type:text"`
	IsActive               bool           `gorm:"default:true"`
	CreatedAt              time.Time      `gorm:"autoCreateTime"`
	UpdatedAt              time.Time      `gorm:"autoUpdateTime"`
}

func (PurposeTemplate) TableName() string {
	return "purpose_templates"
}

// ComplianceReport represents the result of a compliance check
type ComplianceReport struct {
	PurposeID       uuid.UUID `json:"purpose_id"`
	Status          string    `json:"status"` // compliant, needs_review, non_compliant
	Issues          []string  `json:"issues"`
	Recommendations []string  `json:"recommendations"`
	LastChecked     time.Time `json:"last_checked"`
}

// PurposeTree represents a hierarchical structure of purposes
type PurposeTree struct {
	Purpose  Purpose        `json:"purpose"`
	Children []*PurposeTree `json:"children,omitempty"`
}

// UsageStats represents analytics for purpose usage
type UsageStats struct {
	PurposeID        uuid.UUID  `json:"purpose_id"`
	ActiveConsents   int        `json:"active_consents"`
	ConsentForms     int        `json:"consent_forms"`
	ExpiringConsents int        `json:"expiring_consents"`
	TotalGranted     int        `json:"total_granted"`
	TotalRevoked     int        `json:"total_revoked"`
	LastUsed         *time.Time `json:"last_used"`
}

type PurposeStatus struct {
	Name     string `json:"name"`
	Status   bool   `json:"status"`
	Version  string `json:"version"`
	Language string `json:"language"`
}

type PendingConsent struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	MinorUserID    uuid.UUID  `gorm:"type:uuid"`
	GuardianUserID *uuid.UUID `gorm:"type:uuid;default:null"` // dashboard flow, null for DigiLocker
	Updates        []byte     `gorm:"type:jsonb"`             // json.Marshal([]ConsentUpdateRequest)
	Status         string     `gorm:"type:varchar(32)"`
	Token          string     `gorm:"type:text"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ConsentHistory struct {
	ID             uuid.UUID `gorm:"primaryKey"`
	ConsentID      uuid.UUID `gorm:"index"`
	UserID         uuid.UUID `gorm:"index"`
	TenantID       uuid.UUID `gorm:"index"`
	Action         string
	Purposes       datatypes.JSON `gorm:"type:jsonb" json:"purposes"`
	ChangedBy      string
	PolicySnapshot datatypes.JSON `gorm:"type:jsonb"`
	Timestamp      time.Time      `gorm:"autoCreateTime"`
	ReviewTokenID  *uuid.UUID     `gorm:"index"`
}

func (ConsentHistory) TableName() string {
	return "consent_histories"
}

type ReviewToken struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	Token     string    `gorm:"uniqueIndex"`
	UserID    uuid.UUID `gorm:"index"`
	TenantID  uuid.UUID `gorm:"index"`
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Consent struct {
	ID             uuid.UUID           `gorm:"primaryKey"`
	UserID         uuid.UUID           `gorm:"index"`
	Purposes       dto.ConsentPurposes `gorm:"type:jsonb" json:"purposes"`
	PolicySnapshot datatypes.JSON      `gorm:"type:jsonb" json:"policy_snapshot"`
	Signature      string
	GeoRegion      string
	Jurisdiction   string
	TenantID       uuid.UUID `gorm:"index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type EncryptedConsent struct {
	ID             uuid.UUID      `gorm:"primaryKey"`
	UserID         uuid.UUID      `gorm:"index"`
	Purposes       datatypes.JSON `gorm:"type:jsonb" json:"purposes"`
	PolicySnapshot datatypes.JSON `gorm:"type:jsonb" json:"policy_snapshot"`
	Signature      string
	GeoRegion      string
	Jurisdiction   string
	TenantID       uuid.UUID `gorm:"index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserConsent struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID `gorm:"type:uuid;index"`
	PurposeID     uuid.UUID `gorm:"type:uuid;index"`
	TenantID      uuid.UUID `gorm:"type:uuid;index"`
	ConsentFormID uuid.UUID `gorm:"type:uuid;index"`
	Status        bool      // true for granted, false for withdrawn
	ExpiresAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ConsentLink struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Link            string    `gorm:"type:text;not null;unique"`
	TenantID        uuid.UUID `gorm:"type:uuid;not null;index"`
	Name            string
	SubmissionCount int    `gorm:"default:0"`
	Metadata        []byte `gorm:"type:jsonb;default:null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (ConsentLink) TableName() string {
	return "consent_links"
}

// -------------------------------
// Privacy & Compliance Models
// -------------------------------
type DSRRequest struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	UserID      uuid.UUID `gorm:"index"`
	TenantID    uuid.UUID `gorm:"index"`
	Type        string    `gorm:"type:varchar(50)"`                    // access, rectification, erasure, portability, restriction, objection
	Status      string    `gorm:"type:varchar(50);default:'pending'"`  // pending, verified, in_progress, review, approved, rejected, completed
	Priority    string    `gorm:"type:varchar(20);default:'medium'"`   // low, medium, high, critical
	SubjectType string    `gorm:"type:varchar(50);default:'customer'"` // customer, employee, partner

	// SLA & Workflow
	RequestedAt time.Time
	DueDate     time.Time  `gorm:"index"`
	AssignedTo  *uuid.UUID `gorm:"type:uuid;index"`
	VerifiedAt  *time.Time
	ProcessedAt *time.Time

	// Data
	RequestDetails datatypes.JSON `gorm:"type:jsonb"` // Specifics of what is requested
	ResultData     datatypes.JSON `gorm:"type:jsonb"` // Link to export or summary of action
	ResolutionNote string         `gorm:"type:text"`
	InternalNotes  string         `gorm:"type:text"`

	ResolvedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
}

type DSRComment struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	RequestID  uuid.UUID `gorm:"index"`
	AuthorID   uuid.UUID `gorm:"index"`
	Content    string    `gorm:"type:text"`
	IsInternal bool      `gorm:"default:true"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

type AuditLog struct {
	LogID         uuid.UUID `gorm:"primaryKey"`
	UserID        uuid.UUID `gorm:"index"`
	TenantID      uuid.UUID `gorm:"index"`
	PurposeID     uuid.UUID `gorm:"index"`
	ActionType    string
	Timestamp     time.Time `gorm:"autoCreateTime"`
	ConsentStatus string
	Initiator     string
	SourceIP      string
	GeoRegion     string
	Jurisdiction  string
	AuditHash     string
	PreviousHash  string
	Details       datatypes.JSON `gorm:"type:jsonb"`
}

type NotificationPreferences struct {
	UserID                     uuid.UUID `gorm:"primaryKey"`
	OnNewGrievance             bool      `gorm:"default:true"`
	OnGrievanceUpdate          bool      `gorm:"default:true"`
	OnConsentUpdate            bool      `gorm:"default:true"`
	OnNewConsentRequest        bool      `gorm:"default:true"`
	OnDataSubjectRequest       bool      `gorm:"default:true"`
	OnDataSubjectRequestUpdate bool      `gorm:"default:true"`
}

// -------------------------------
// Engagement & Notification Models
// -------------------------------
type Grievance struct {
	ID                   uuid.UUID  `gorm:"primaryKey" json:"id"`
	UserID               uuid.UUID  `gorm:"index" json:"userId"`
	TenantID             uuid.UUID  `gorm:"index" json:"tenantId"`
	GrievanceType        string     `json:"grievanceType"`
	GrievanceSubject     string     `json:"grievanceSubject"`
	GrievanceDescription string     `json:"grievanceDescription"`
	Status               string     `json:"status"` // e.g., open, in_progress, resolved, closed
	AssignedTo           *uuid.UUID `gorm:"index" json:"assignedTo,omitempty"`
	Category             string     `json:"category"` // e.g., billing, technical, general
	Priority             string     `json:"priority"` // e.g., low, medium, high, urgent
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

// ===================== GrievanceComment (Chat) =====================
type GrievanceComment struct {
	ID          uuid.UUID  `gorm:"primaryKey" json:"id"`
	GrievanceID uuid.UUID  `gorm:"index" json:"grievanceId"`
	UserID      uuid.UUID  `gorm:"index" json:"userId"`
	AdminId     *uuid.UUID `gorm:"index" json:"adminId,omitempty"`
	Comment     string     `json:"comment"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// ===================== Notification =====================
type Notification struct {
	ID        uuid.UUID `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"index" json:"userId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Icon      string    `json:"icon"`
	Link      string    `json:"link,omitempty"`
	Unread    bool      `gorm:"index" json:"unread"`
	CreatedAt time.Time `json:"createdAt"`
}

// -------------------------------
// API & Webhook Infrastructure
// -------------------------------
type APIKey struct {
	KeyID          uuid.UUID `gorm:"primaryKey"`
	TenantID       uuid.UUID `gorm:"index"`
	UserID         uuid.UUID `gorm:"index"`
	Label          string
	HashedKey      string
	Scopes         datatypes.JSON `gorm:"type:jsonb" json:"scopes"`
	CreatedAt      time.Time
	LastUsedAt     *time.Time
	Revoked        bool
	RevokedAt      *time.Time
	ExpiresAt      *time.Time
	WhitelistedIPs datatypes.JSON `gorm:"type:jsonb" json:"whitelisted_ips"`
}

// -------------------------------
// Tenant Infrastructure
// -------------------------------
type Tenant struct {
	TenantID              uuid.UUID `gorm:"primaryKey"`
	Name                  string
	Cluster               string
	Industry              string
	CompanySize           string
	Config                datatypes.JSON
	ReviewFrequencyMonths int `gorm:"default:6"`
	CreatedAt             time.Time
}

// -------------------------------
// Data Processor
// -------------------------------
type Vendor struct {
	VendorID uuid.UUID `gorm:"primaryKey" json:"id"`
	Company  string
	Email    string `gorm:"type:text;uniqueIndex"`
	Address  string
	// DPA-related fields
	DPAAgreementID         *uuid.UUID `gorm:"type:uuid" json:"dpaAgreementId,omitempty"`
	ProcessingLocation     string     `gorm:"type:text" json:"processingLocation,omitempty"`
	SecurityCertifications string     `gorm:"type:text" json:"securityCertifications,omitempty"`
	LastComplianceCheck    *time.Time `gorm:"type:timestamp" json:"lastComplianceCheck,omitempty"`
	ComplianceStatus       string     `gorm:"type:text" json:"complianceStatus,omitempty"`
	// TPRM-related fields
	RiskScore float64   `gorm:"type:decimal(5,2);default:0" json:"riskScore,omitempty"`
	RiskLevel string    `gorm:"type:varchar(20);default:'low'" json:"riskLevel,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// -------------------------------
// Third-Party Risk Management (TPRM) Models
// -------------------------------

// TPRMAssessment represents a structured risk assessment performed on a vendor
type TPRMAssessment struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID      uuid.UUID  `gorm:"type:uuid;index" json:"tenantId"`
	VendorID      uuid.UUID  `gorm:"type:uuid;index" json:"vendorId"`
	Title         string     `gorm:"type:text" json:"title"`
	Framework     string     `gorm:"type:varchar(50)" json:"framework"`                // e.g., ISO27001, SOC2
	Status        string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, in_progress, completed, failed
	DueDate       *time.Time `gorm:"type:timestamp" json:"dueDate,omitempty"`
	CompletedAt   *time.Time `gorm:"type:timestamp" json:"completedAt,omitempty"`
	RiskScore     float64    `gorm:"type:decimal(5,2);default:0" json:"riskScore"`
	FindingsCount int        `gorm:"default:0" json:"findingsCount"`
	EvidenceCount int        `gorm:"default:0" json:"evidenceCount"`
	AssessorID    *uuid.UUID `gorm:"type:uuid" json:"assessorId,omitempty"`
	Notes         string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TPRMEvidence stores uploaded evidence files for an assessment
type TPRMEvidence struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AssessmentID uuid.UUID `gorm:"type:uuid;index" json:"assessmentId"`
	TenantID     uuid.UUID `gorm:"type:uuid;index" json:"tenantId"`
	VendorID     uuid.UUID `gorm:"type:uuid;index" json:"vendorId"`
	FilePath     string    `gorm:"type:text" json:"filePath"`
	ContentType  string    `gorm:"type:varchar(100)" json:"contentType"`
	SizeBytes    int64     `gorm:"type:bigint" json:"sizeBytes"`
	UploadedAt   time.Time `gorm:"autoCreateTime" json:"uploadedAt"`
	UploadedBy   uuid.UUID `gorm:"type:uuid;index" json:"uploadedBy"`
}

// TPRMFinding captures a risk finding from an assessment
type TPRMFinding struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AssessmentID uuid.UUID `gorm:"type:uuid;index" json:"assessmentId"`
	TenantID     uuid.UUID `gorm:"type:uuid;index" json:"tenantId"`
	VendorID     uuid.UUID `gorm:"type:uuid;index" json:"vendorId"`
	Severity     string    `gorm:"type:varchar(20)" json:"severity"` // low, medium, high, critical
	Title        string    `gorm:"type:text" json:"title"`
	Description  string    `gorm:"type:text" json:"description"`
	Remediation  string    `gorm:"type:text" json:"remediation"`
	Status       string    `gorm:"type:varchar(20);default:'open'" json:"status"` // open, in_progress, resolved
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// -------------------------------
// Consent Form Models
// -------------------------------
type ConsentForm struct {
	ID                   uuid.UUID            `gorm:"type:uuid;primaryKey"`
	TenantID             uuid.UUID            `gorm:"type:uuid;index"`
	Name                 string               `gorm:"type:text"`
	Title                string               `gorm:"type:text"`
	Description          string               `gorm:"type:text"`
	Department           string               `gorm:"type:text"`
	Project              string               `gorm:"type:text"`
	OrganizationEntityID uuid.UUID            `gorm:"type:uuid;index"`
	Purposes             []ConsentFormPurpose `gorm:"foreignKey:ConsentFormID"`
	DataRetentionPeriod  string               `gorm:"type:text"`
	UserRightsSummary    string               `gorm:"type:text"`
	TermsAndConditions   string               `gorm:"type:text"`
	PrivacyPolicy        string               `gorm:"type:text"`
	Published            bool                 `gorm:"default:false"`
	FormLink             string               `gorm:"type:text;uniqueIndex"`
	FormSDK              string               `gorm:"type:text"`
	// Enterprise Features
	Translations datatypes.JSON `gorm:"type:jsonb" json:"translations"` // Map[langCode]Content
	Regions      pq.StringArray `gorm:"type:text[]" json:"regions"`     // e.g. ["EU", "US-CA"]
	// Versioning and status fields
	CurrentVersion  int    `gorm:"default:1"`
	Status          string `gorm:"type:varchar(20);default:'draft'"` // draft, review, published, archived
	LastPublishedAt *time.Time
	LastPublishedBy *uuid.UUID `gorm:"type:uuid"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
}

type ConsentFormVersion struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ConsentFormID uuid.UUID      `gorm:"type:uuid;index"`
	VersionNumber int            `gorm:"not null"`
	Snapshot      datatypes.JSON `gorm:"type:jsonb"` // Full form data snapshot
	PublishedAt   time.Time      `gorm:"autoCreateTime"`
	PublishedBy   uuid.UUID      `gorm:"type:uuid"`        // Fiduciary user ID
	Status        string         `gorm:"type:varchar(20)"` // draft, review, published, archived
	ChangeLog     string         `gorm:"type:text"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
}

type ConsentFormPurpose struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ConsentFormID uuid.UUID      `gorm:"type:uuid;index"`
	PurposeID     uuid.UUID      `gorm:"type:uuid;index"`
	Purpose       Purpose        `gorm:"foreignKey:PurposeID"`
	DataObjects   pq.StringArray `gorm:"type:text[]"`
	VendorIDs     pq.StringArray `gorm:"type:text[]"`
	ExpiryInDays  int
}

// -------------------------------
// Consent Receipt Models
// -------------------------------

// ConsentReceipt represents a DPDP-compliant consent receipt with PDF generation and verification
type ConsentReceipt struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserConsentID uuid.UUID      `gorm:"type:uuid;index;not null"` // Link to UserConsent
	TenantID      uuid.UUID      `gorm:"type:uuid;index;not null"`
	ReceiptNumber string         `gorm:"type:varchar(50);uniqueIndex;not null"` // Format: RCP-{YYYY}{MM}{DD}-{RANDOM}
	PDFPath       string         `gorm:"type:text"`                             // Path to stored PDF file
	QRCodeData    string         `gorm:"type:text"`                             // QR code verification data
	GeneratedAt   time.Time      `gorm:"autoCreateTime"`
	EmailedAt     *time.Time     `gorm:"type:timestamp"`
	DownloadCount int            `gorm:"default:0"`
	IsValid       bool           `gorm:"default:true"`
	ExpiresAt     *time.Time     `gorm:"type:timestamp"`
	Metadata      datatypes.JSON `gorm:"type:jsonb"` // Additional receipt metadata
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
}

// -------------------------------
// Breach Notification Models
// -------------------------------

type BreachNotification struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID           uuid.UUID `gorm:"type:uuid;index"`
	Title              string    `gorm:"type:varchar(255)"` // Brief title of breach
	Description        string    `gorm:"type:text"`
	BreachDate         time.Time
	DetectionDate      time.Time
	ContainmentDate    *time.Time // When breach was contained
	NotificationDate   *time.Time
	AffectedUsersCount int
	NotifiedUsersCount int
	Severity           string `gorm:"type:varchar(20)"` // low, medium, high, critical
	BreachType         string `gorm:"type:varchar(50)"` // unauthorized_access, data_theft, data_loss, ransomware, etc.
	Status             string `gorm:"type:varchar(50)"` // draft, pending_verification, verified, notifying, notified, resolved, closed

	// DPDP-specific fields
	RequiresDPBReporting    bool `gorm:"default:false"`
	DPBReported             bool `gorm:"default:false"`
	DPBReportedDate         *time.Time
	DPBReportReference      *string    `gorm:"type:text"`
	DPBNotificationDeadline *time.Time // Calculated deadline for DPB notification

	// Data categories affected (DPDP requirement)
	DataCategoriesAffected datatypes.JSON `gorm:"type:jsonb"` // ["personal_identifiable", "financial", "health", "biometric", etc.]

	// Affected individuals notification
	RequiresDataPrincipalNotification   bool       `gorm:"default:false"`
	DataPrincipalNotificationApproved   bool       `gorm:"default:false"`
	DataPrincipalNotificationApprovedBy *uuid.UUID `gorm:"type:uuid"`
	DataPrincipalNotificationApprovedAt *time.Time
	DataPrincipalNotificationSentAt     *time.Time
	DataPrincipalNotificationDeadline   *time.Time

	// Workflow tracking
	CurrentWorkflowStage string     `gorm:"type:varchar(50)"` // detection, assessment, containment, verification, notification, resolution
	VerifiedBy           *uuid.UUID `gorm:"type:uuid"`
	VerifiedAt           *time.Time

	// Risk assessment
	LikelihoodOfHarm       string `gorm:"type:varchar(20)"` // low, medium, high
	ImpactOnDataPrincipals string `gorm:"type:text"`

	// Remediation details
	RemedialActions    datatypes.JSON `gorm:"type:jsonb"`
	PreventiveMeasures datatypes.JSON `gorm:"type:jsonb"`
	LessonsLearned     *string        `gorm:"type:text"`

	// Investigation details
	InvestigationSummary *string `gorm:"type:text"`
	InvestigatedBy       *string `gorm:"type:text"`
	InvestigationDate    *time.Time
	RootCause            *string `gorm:"type:text"`

	// Compliance details
	ComplianceStatus string  `gorm:"type:varchar(50)"`
	ComplianceNotes  *string `gorm:"type:text"`
	IsOverdue        bool    `gorm:"default:false"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type EncryptedBreachNotification struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID           uuid.UUID `gorm:"type:uuid;index"`
	Description        string    `gorm:"type:text"`
	BreachDate         time.Time
	DetectionDate      time.Time
	NotificationDate   *time.Time
	AffectedUsersCount int
	NotifiedUsersCount int
	Severity           string `gorm:"type:varchar(20)"` // low, medium, high, critical
	BreachType         string `gorm:"type:varchar(50)"` // unauthorized_access, data_theft, etc.
	Status             string `gorm:"type:varchar(50)"` // e.g., Investigating, Notifying, Resolved
	// DPDP-specific fields
	RequiresDPBReporting bool `gorm:"default:false"`
	DPBReported          bool `gorm:"default:false"`
	DPBReportedDate      *time.Time
	DPBReportReference   *string `gorm:"type:text"`
	// Remediation details
	RemedialActions    datatypes.JSON `gorm:"type:jsonb"`
	PreventiveMeasures datatypes.JSON `gorm:"type:jsonb"`
	// Investigation details
	InvestigationSummary *string `gorm:"type:text"`
	InvestigatedBy       *string `gorm:"type:text"`
	InvestigationDate    *time.Time
	// Compliance details
	ComplianceStatus string    `gorm:"type:varchar(50)"`
	ComplianceNotes  *string   `gorm:"type:text"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

func (b BreachNotification) BreachSeverity(severity string) string {
	switch severity {
	case "low":
		return "Low"
	case "medium":
		return "Medium"
	case "high":
		return "High"
	case "critical":
		return "Critical"
	default:
		return "Unknown"
	}
}

// -------------------------------
// SDK Configuration Models
// -------------------------------

// SDKConfig stores configuration for JavaScript SDK generation
type SDKConfig struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID             uuid.UUID      `gorm:"type:uuid;index"`
	ConsentFormID        uuid.UUID      `gorm:"type:uuid;index"`
	Theme                datatypes.JSON `gorm:"type:jsonb"`                        // colors, fonts, layout
	Position             string         `gorm:"type:varchar(20);default:'bottom'"` // top, bottom, left, right, center
	Language             string         `gorm:"type:varchar(10);default:'en'"`
	ShowPreferenceCenter bool           `gorm:"default:true"`
	AutoShow             bool           `gorm:"default:true"`
	CookieExpiry         int            `gorm:"default:365"` // days
	CustomCSS            string         `gorm:"type:text"`
	CreatedAt            time.Time      `gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime"`
}

// -------------------------------
// Cookie Management Models
// -------------------------------

// Cookie represents a detected or managed cookie
type Cookie struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID      uuid.UUID `gorm:"type:uuid;index"`
	Name          string    `gorm:"type:varchar(255);not null;index"`
	Domain        string    `gorm:"type:varchar(255);not null"`
	Path          string    `gorm:"type:varchar(255);default:'/'"`
	Category      string    `gorm:"type:varchar(50);not null;index"` // necessary, functional, analytics, marketing
	Purpose       string    `gorm:"type:text"`
	Provider      string    `gorm:"type:varchar(255)"`
	ExpiryDays    int       `gorm:"default:0"` // 0 = session cookie
	IsFirstParty  bool      `gorm:"default:true"`
	IsSecure      bool      `gorm:"default:false"`
	IsHttpOnly    bool      `gorm:"default:false"`
	SameSite      string    `gorm:"type:varchar(20)"` // None, Lax, Strict
	Description   string    `gorm:"type:text"`
	DataCollected string    `gorm:"type:text"`
	IsActive      bool      `gorm:"default:true"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

// CookieScan represents a website cookie scanning session
type CookieScan struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID     uuid.UUID      `gorm:"type:uuid;index"`
	URL          string         `gorm:"type:text;not null"`
	ScanDate     time.Time      `gorm:"autoCreateTime"`
	CookiesFound int            `gorm:"default:0"`
	NewCookies   int            `gorm:"default:0"`
	Status       string         `gorm:"type:varchar(50);default:'pending'"` // pending, running, completed, failed
	ScanDuration int            `gorm:"default:0"`                          // in milliseconds
	Results      datatypes.JSON `gorm:"type:jsonb"`
	ErrorMessage string         `gorm:"type:text"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
}

// CookieConsent links user consent to specific cookies
type CookieConsent struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserConsentID uuid.UUID `gorm:"type:uuid;index"`
	CookieID      uuid.UUID `gorm:"type:uuid;index"`
	Allowed       bool      `gorm:"default:false"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

// CookieCategory enum values
const (
	CookieCategoryNecessary  = "necessary"
	CookieCategoryFunctional = "functional"
	CookieCategoryAnalytics  = "analytics"
	CookieCategoryMarketing  = "marketing"
)

// CookieScanStatus enum values
const (
	CookieScanStatusPending   = "pending"
	CookieScanStatusRunning   = "running"
	CookieScanStatusCompleted = "completed"
	CookieScanStatusFailed    = "failed"
)

// -------------------------------
// Webhook Models
// -------------------------------

// Webhook stores configuration for a single webhook endpoint.
type Webhook struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID   uuid.UUID      `gorm:"type:uuid;index"`
	URL        string         `gorm:"type:text;not null"`
	Secret     string         `gorm:"type:text;not null"` // Used to sign outgoing payloads
	EventTypes pq.StringArray `gorm:"type:text[]"`        // e.g., ["consent.updated", "dsr.created"]
	IsActive   bool           `gorm:"default:true"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
}

// WebhookEvent logs an attempt to send a webhook.
type WebhookEvent struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	WebhookID   uuid.UUID `gorm:"type:uuid;index"`
	EventType   string    `gorm:"type:varchar(100)"`
	Payload     datatypes.JSON
	Success     bool
	Response    string `gorm:"type:text"`
	AttemptedAt time.Time
}
