package models

import (
	"time"

	"github.com/google/uuid"
)

// PlatformSession represents user sessions per platform
type PlatformSession struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Platform     string    `json:"platform" gorm:"type:varchar(20);not null;check:platform IN ('web', 'desktop')"`
	SessionToken string    `json:"session_token" gorm:"type:varchar(255);unique;not null"`
	IPAddress    string    `json:"ip_address" gorm:"type:inet"`
	UserAgent    string    `json:"user_agent" gorm:"type:text"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Note: APIKey is already defined in models.go

// PublicConsentSubmission represents consent submissions from public websites
type PublicConsentSubmission struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TenantID    uuid.UUID              `json:"tenant_id" gorm:"type:uuid;not null;index"`
	WebsiteID   *uuid.UUID             `json:"website_id" gorm:"type:uuid;index"`
	VisitorID   string                 `json:"visitor_id" gorm:"type:varchar(255);index"`
	ConsentData map[string]interface{} `json:"consent_data" gorm:"type:jsonb"`
	IPAddress   string                 `json:"ip_address" gorm:"type:inet"`
	UserAgent   string                 `json:"user_agent" gorm:"type:text"`
	Referrer    string                 `json:"referrer" gorm:"type:text"`
	SubmittedAt time.Time              `json:"submitted_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Tenant  Tenant  `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Website Website `json:"website,omitempty" gorm:"foreignKey:WebsiteID"`
}

// Website represents client websites that integrate with the consent manager
type Website struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TenantID    uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Domain      string    `json:"domain" gorm:"type:varchar(255);not null;index"`
	Description string    `json:"description" gorm:"type:text"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Tenant Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// SecurityAuditLog represents security-related audit events
type SecurityAuditLog struct {
	ID        uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TenantID  uuid.UUID              `json:"tenant_id" gorm:"type:uuid;not null;index"`
	UserID    *uuid.UUID             `json:"user_id" gorm:"type:uuid;index"`
	SessionID *uuid.UUID             `json:"session_id" gorm:"type:uuid;index"`
	EventType string                 `json:"event_type" gorm:"type:varchar(50);not null;index"`
	Severity  string                 `json:"severity" gorm:"type:varchar(20);not null;check:severity IN ('low', 'medium', 'high', 'critical')"`
	IPAddress string                 `json:"ip_address" gorm:"type:inet"`
	UserAgent string                 `json:"user_agent" gorm:"type:text"`
	Resource  string                 `json:"resource" gorm:"type:varchar(255)"`
	Action    string                 `json:"action" gorm:"type:varchar(100)"`
	Result    string                 `json:"result" gorm:"type:varchar(50)"`
	Details   map[string]interface{} `json:"details" gorm:"type:jsonb"`
	Timestamp time.Time              `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP;index"`

	// Relationships
	Tenant  Tenant          `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	User    User            `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Session PlatformSession `json:"session,omitempty" gorm:"foreignKey:SessionID"`
}

// RateLimitEntry represents rate limiting data
type RateLimitEntry struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Identifier  string    `json:"identifier" gorm:"type:varchar(255);not null;index"` // IP, user ID, API key
	Resource    string    `json:"resource" gorm:"type:varchar(255);not null"`
	Count       int       `json:"count" gorm:"default:1"`
	WindowStart time.Time `json:"window_start" gorm:"index"`
	ExpiresAt   time.Time `json:"expires_at" gorm:"index"`
}

// OAuthClient represents OAuth 2.0 clients
type OAuthClient struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TenantID     uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	ClientID     string    `json:"client_id" gorm:"type:varchar(255);unique;not null"`
	ClientSecret string    `json:"-" gorm:"type:varchar(255);not null"` // Never expose in JSON
	Name         string    `json:"name" gorm:"type:varchar(255);not null"`
	RedirectURIs []string  `json:"redirect_uris" gorm:"type:jsonb"`
	Scopes       []string  `json:"scopes" gorm:"type:jsonb"`
	GrantTypes   []string  `json:"grant_types" gorm:"type:jsonb"`
	IsPublic     bool      `json:"is_public" gorm:"default:false"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Tenant Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// OAuthToken represents OAuth 2.0 tokens
type OAuthToken struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClientID     string     `json:"client_id" gorm:"type:varchar(255);not null;index"`
	UserID       *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	TokenType    string     `json:"token_type" gorm:"type:varchar(50);not null"`
	AccessToken  string     `json:"-" gorm:"type:text;not null"` // Never expose in JSON
	RefreshToken string     `json:"-" gorm:"type:text"`
	Scopes       []string   `json:"scopes" gorm:"type:jsonb"`
	ExpiresAt    time.Time  `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at"`

	// Relationships
	Client OAuthClient `json:"client,omitempty" gorm:"foreignKey:ClientID;references:ClientID"`
	User   User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// SecurityConfiguration represents tenant-specific security settings
type SecurityConfiguration struct {
	ID                   uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TenantID             uuid.UUID              `json:"tenant_id" gorm:"type:uuid;unique;not null"`
	PasswordPolicy       map[string]interface{} `json:"password_policy" gorm:"type:jsonb"`
	SessionTimeout       int                    `json:"session_timeout" gorm:"default:3600"` // seconds
	MaxLoginAttempts     int                    `json:"max_login_attempts" gorm:"default:5"`
	LockoutDuration      int                    `json:"lockout_duration" gorm:"default:900"` // seconds
	RequireMFA           bool                   `json:"require_mfa" gorm:"default:false"`
	AllowedIPRanges      []string               `json:"allowed_ip_ranges" gorm:"type:jsonb"`
	BlockedIPRanges      []string               `json:"blocked_ip_ranges" gorm:"type:jsonb"`
	EnableRequestSigning bool                   `json:"enable_request_signing" gorm:"default:false"`
	EnableAuditLogging   bool                   `json:"enable_audit_logging" gorm:"default:true"`
	DataRetentionDays    int                    `json:"data_retention_days" gorm:"default:2555"` // 7 years
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`

	// Relationships
	Tenant Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// Constants for security events
const (
	// Authentication events
	SecurityEventLogin         = "login"
	SecurityEventLoginFailed   = "login_failed"
	SecurityEventLogout        = "logout"
	SecurityEventTokenRefresh  = "token_refresh"
	SecurityEventPasswordReset = "password_reset"

	// Authorization events
	SecurityEventAccessDenied    = "access_denied"
	SecurityEventRoleChanged     = "role_changed"
	SecurityEventPermissionCheck = "permission_check"

	// Data access events
	SecurityEventDataAccess   = "data_access"
	SecurityEventDataExport   = "data_export"
	SecurityEventDataImport   = "data_import"
	SecurityEventDataDeletion = "data_deletion"

	// API events
	SecurityEventAPIKeyUsed    = "api_key_used"
	SecurityEventAPIKeyCreated = "api_key_created"
	SecurityEventAPIKeyRevoked = "api_key_revoked"
	SecurityEventRateLimited   = "rate_limited"

	// Security events
	SecurityEventSuspiciousActivity = "suspicious_activity"
	SecurityEventBruteForce         = "brute_force"
	SecurityEventIPBlocked          = "ip_blocked"
	SecurityEventMaliciousRequest   = "malicious_request"

	// Severity levels
	SecuritySeverityLow      = "low"
	SecuritySeverityMedium   = "medium"
	SecuritySeverityHigh     = "high"
	SecuritySeverityCritical = "critical"
)

// Helper methods for SecurityAuditLog
func (s *SecurityAuditLog) IsCritical() bool {
	return s.Severity == SecuritySeverityCritical
}

func (s *SecurityAuditLog) IsSecurityThreat() bool {
	threatEvents := []string{
		SecurityEventLoginFailed,
		SecurityEventAccessDenied,
		SecurityEventSuspiciousActivity,
		SecurityEventBruteForce,
		SecurityEventMaliciousRequest,
	}

	for _, event := range threatEvents {
		if s.EventType == event {
			return true
		}
	}
	return false
}

// Helper methods for PlatformSession
func (p *PlatformSession) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

func (p *PlatformSession) IsWebSession() bool {
	return p.Platform == "web"
}

func (p *PlatformSession) IsDesktopSession() bool {
	return p.Platform == "desktop"
}

