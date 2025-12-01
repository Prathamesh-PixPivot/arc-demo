package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateDSRRequest is alrready defined in types.go

// PublicCookieConsentRequest represents cookie consent from website visitors
type PublicCookieConsentRequest struct {
	VisitorID   string            `json:"visitor_id" validate:"required"`
	Domain      string            `json:"domain" validate:"required"`
	Consents    map[string]bool   `json:"consents" validate:"required"`
	CookieIDs   []string          `json:"cookie_ids"`
	Preferences map[string]string `json:"preferences"`
	IPAddress   string            `json:"ip_address"`
	UserAgent   string            `json:"user_agent"`
}

// CreateUserConsentRequest represents consent creation request
type CreateUserConsentRequest struct {
	ConsentFormID uuid.UUID              `json:"consent_form_id" validate:"required"`
	Purposes      []string               `json:"purposes" validate:"required,min=1"`
	DataObjects   []string               `json:"data_objects"`
	Channel       string                 `json:"channel" validate:"required"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ConsentResponse represents the response after consent creation
type ConsentResponse struct {
	ID            uuid.UUID `json:"id"`
	Status        string    `json:"status"`
	ReceiptNumber string    `json:"receipt_number,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// PrivacyPolicyResponse represents privacy policy data
type PrivacyPolicyResponse struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	Language     string    `json:"language"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	LastUpdated  time.Time `json:"last_updated"`
	Version      string    `json:"version"`
	EffectiveDate time.Time `json:"effective_date"`
}

// TermsOfServiceResponse represents terms of service data
type TermsOfServiceResponse struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	Language      string    `json:"language"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	LastUpdated   time.Time `json:"last_updated"`
	Version       string    `json:"version"`
	EffectiveDate time.Time `json:"effective_date"`
}

// CookieSettingsResponse represents cookie settings for a domain
type CookieSettingsResponse struct {
	TenantID   uuid.UUID                    `json:"tenant_id"`
	Domain     string                       `json:"domain"`
	Categories map[string]CookieCategoryInfo `json:"categories"`
	Cookies    []CookieInfo                 `json:"cookies"`
}

// CookieCategoryInfo represents information about a cookie category
type CookieCategoryInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Enabled     bool   `json:"enabled"`
}

// CookieInfo represents information about a specific cookie
type CookieInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Purpose     string    `json:"purpose"`
	Provider    string    `json:"provider"`
	ExpiryDays  int       `json:"expiry_days"`
	Description string    `json:"description"`
}

