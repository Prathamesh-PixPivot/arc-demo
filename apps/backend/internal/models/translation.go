package models

import (
	"time"

	"github.com/google/uuid"
)

// Translation represents a language translation entry
type Translation struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	LanguageCode string                 `json:"language_code" gorm:"type:varchar(10);not null;index:idx_translation_lang_key"`
	LanguageName string                 `json:"language_name" gorm:"type:varchar(50);not null"`
	Key          string                 `json:"key" gorm:"type:varchar(255);not null;index:idx_translation_lang_key"`
	Value        string                 `json:"value" gorm:"type:text;not null"`
	Context      string                 `json:"context,omitempty" gorm:"type:varchar(100)"` // e.g., "consent_form", "dsr", "notification"
	Metadata     map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	IsActive     bool                   `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Language represents a supported language
type Language struct {
	Code        string    `json:"code" gorm:"type:varchar(10);primary_key"`
	Name        string    `json:"name" gorm:"type:varchar(50);not null"`
	NativeName  string    `json:"native_name" gorm:"type:varchar(50);not null"`
	Direction   string    `json:"direction" gorm:"type:varchar(3);default:'ltr'"` // ltr or rtl
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TranslationTemplate represents a template for translations with placeholders
type TranslationTemplate struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Key          string                 `json:"key" gorm:"type:varchar(255);unique;not null"`
	Description  string                 `json:"description" gorm:"type:text"`
	Placeholders []string               `json:"placeholders" gorm:"type:jsonb"` // e.g., ["{{name}}", "{{date}}"]
	Context      string                 `json:"context" gorm:"type:varchar(100)"`
	Category     string                 `json:"category" gorm:"type:varchar(50)"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// TenantTranslation represents tenant-specific translation overrides
type TenantTranslation struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TenantID     uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index:idx_tenant_translation"`
	LanguageCode string    `json:"language_code" gorm:"type:varchar(10);not null;index:idx_tenant_translation"`
	Key          string    `json:"key" gorm:"type:varchar(255);not null;index:idx_tenant_translation"`
	Value        string    `json:"value" gorm:"type:text;not null"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserLanguagePreference represents a user's language preference
type UserLanguagePreference struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;unique;not null"`
	LanguageCode string    `json:"language_code" gorm:"type:varchar(10);not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DPDPA 2023 mandated languages (22 official languages of India)
var DPDPALanguages = []Language{
	{Code: "en", Name: "English", NativeName: "English", Direction: "ltr", IsDefault: true, SortOrder: 1},
	{Code: "hi", Name: "Hindi", NativeName: "हिन्दी", Direction: "ltr", SortOrder: 2},
	{Code: "bn", Name: "Bengali", NativeName: "বাংলা", Direction: "ltr", SortOrder: 3},
	{Code: "te", Name: "Telugu", NativeName: "తెలుగు", Direction: "ltr", SortOrder: 4},
	{Code: "mr", Name: "Marathi", NativeName: "मराठी", Direction: "ltr", SortOrder: 5},
	{Code: "ta", Name: "Tamil", NativeName: "தமிழ்", Direction: "ltr", SortOrder: 6},
	{Code: "ur", Name: "Urdu", NativeName: "اردو", Direction: "rtl", SortOrder: 7},
	{Code: "gu", Name: "Gujarati", NativeName: "ગુજરાતી", Direction: "ltr", SortOrder: 8},
	{Code: "kn", Name: "Kannada", NativeName: "ಕನ್ನಡ", Direction: "ltr", SortOrder: 9},
	{Code: "ml", Name: "Malayalam", NativeName: "മലയാളം", Direction: "ltr", SortOrder: 10},
	{Code: "or", Name: "Odia", NativeName: "ଓଡ଼ିଆ", Direction: "ltr", SortOrder: 11},
	{Code: "pa", Name: "Punjabi", NativeName: "ਪੰਜਾਬੀ", Direction: "ltr", SortOrder: 12},
	{Code: "as", Name: "Assamese", NativeName: "অসমীয়া", Direction: "ltr", SortOrder: 13},
	{Code: "mai", Name: "Maithili", NativeName: "मैथिली", Direction: "ltr", SortOrder: 14},
	{Code: "sat", Name: "Santali", NativeName: "ᱥᱟᱱᱛᱟᱲᱤ", Direction: "ltr", SortOrder: 15},
	{Code: "ks", Name: "Kashmiri", NativeName: "कॉशुर", Direction: "rtl", SortOrder: 16},
	{Code: "ne", Name: "Nepali", NativeName: "नेपाली", Direction: "ltr", SortOrder: 17},
	{Code: "sd", Name: "Sindhi", NativeName: "سنڌي", Direction: "rtl", SortOrder: 18},
	{Code: "kok", Name: "Konkani", NativeName: "कोंकणी", Direction: "ltr", SortOrder: 19},
	{Code: "mni", Name: "Manipuri", NativeName: "মৈতৈলোন্", Direction: "ltr", SortOrder: 20},
	{Code: "doi", Name: "Dogri", NativeName: "डोगरी", Direction: "ltr", SortOrder: 21},
	{Code: "bodo", Name: "Bodo", NativeName: "बर'", Direction: "ltr", SortOrder: 22},
}

// TranslationKey constants for common keys
const (
	// Consent related
	TransKeyConsentTitle      = "consent.title"
	TransKeyConsentAgree      = "consent.agree"
	TransKeyConsentDisagree   = "consent.disagree"
	TransKeyConsentWithdraw   = "consent.withdraw"
	TransKeyConsentPurpose    = "consent.purpose"
	TransKeyConsentDataObject = "consent.data_object"
	TransKeyConsentExpiry     = "consent.expiry"
	
	// DSR related
	TransKeyDSRTitle         = "dsr.title"
	TransKeyDSRAccess        = "dsr.access"
	TransKeyDSRCorrection    = "dsr.correction"
	TransKeyDSRErasure       = "dsr.erasure"
	TransKeyDSRPortability   = "dsr.portability"
	TransKeyDSRObjection     = "dsr.objection"
	
	// Notification related
	TransKeyNotificationEmail = "notification.email"
	TransKeyNotificationSMS   = "notification.sms"
	TransKeyNotificationPush  = "notification.push"
	
	// Common UI elements
	TransKeyUISubmit         = "ui.submit"
	TransKeyUICancel         = "ui.cancel"
	TransKeyUISave           = "ui.save"
	TransKeyUIDelete         = "ui.delete"
	TransKeyUIEdit           = "ui.edit"
	TransKeyUIView           = "ui.view"
	TransKeyUISearch         = "ui.search"
	TransKeyUIFilter         = "ui.filter"
	TransKeyUIExport         = "ui.export"
	TransKeyUIImport         = "ui.import"
	
	// DPDPA specific
	TransKeyDPDPANotice      = "dpdpa.notice"
	TransKeyDPDPARights      = "dpdpa.rights"
	TransKeyDPDPAGrievance   = "dpdpa.grievance"
	TransKeyDPDPADPO         = "dpdpa.dpo_contact"
)

