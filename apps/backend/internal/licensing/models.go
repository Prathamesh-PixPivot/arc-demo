package licensing

import (
	"time"

	"github.com/google/uuid"
)

type LicenseType string
type PlanTier string

const (
	LicenseTypeSaaS   LicenseType = "SAAS"
	LicenseTypeOnPrem LicenseType = "ON_PREM"

	PlanTierStarter    PlanTier = "STARTER"
	PlanTierPro        PlanTier = "PRO"
	PlanTierEnterprise PlanTier = "ENTERPRISE"
)

// License represents the decoded license payload
type License struct {
	LicenseID    uuid.UUID      `json:"license_id"`
	CustomerID   uuid.UUID      `json:"customer_id"`
	CustomerName string         `json:"customer_name"`
	Issuer       string         `json:"issuer"`
	IssuedAt     time.Time      `json:"issued_at"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"` // Null for perpetual
	Type         LicenseType    `json:"type"`
	PlanTier     PlanTier       `json:"plan_tier"`
	Features     []string       `json:"features"`
	Limits       LicenseLimits  `json:"limits"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// LicenseLimits defines the quotas and limits
type LicenseLimits struct {
	// SaaS Limits
	MonthlyAPIRequests  int64 `json:"monthly_api_requests,omitempty"`
	PIIRecordsProcessed int64 `json:"pii_records_processed,omitempty"`

	// On-Prem Limits
	MaxUsers   int `json:"max_users,omitempty"`
	MaxDomains int `json:"max_domains,omitempty"`
	MaxCores   int `json:"max_cores,omitempty"`
}

// IsExpired checks if the license is expired
func (l *License) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*l.ExpiresAt)
}

// HasFeature checks if a feature is enabled
func (l *License) HasFeature(feature string) bool {
	for _, f := range l.Features {
		if f == feature {
			return true
		}
	}
	return false
}
