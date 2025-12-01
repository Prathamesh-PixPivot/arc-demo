package dto

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AuditLogRequest struct {
	//UserID uuid.UUID, tenantID uuid.UUID, purposeID uuid.UUID, actionType string, consentStatus string, initiator string, sourceIP string, geoRegion string, jurisdiction string, details map[string]interface{}
	UserID        string                 `json:"userId"`
	TenantID      string                 `json:"tenantId"`
	PurposeID     string                 `json:"purposeId"`
	ActionType    string                 `json:"actionType"`
	ConsentStatus string                 `json:"consentStatus"`
	Initiator     string                 `json:"initiator"`
	SourceIP      string                 `json:"sourceIp"`
	GeoRegion     string                 `json:"geoRegion"`
	Jurisdiction  string                 `json:"jurisdiction"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

type Purpose struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Consented bool      `json:"consented"`
	Version   string    `json:"version,omitempty"`
	Language  string    `json:"language,omitempty"`
}

type ConsentPurpose struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Status      bool       `json:"status"` // e.g., "active", "withdrawn"
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

type ConsentPurposes struct {
	Purposes []ConsentPurpose `json:"purposes"`
}

type CreateConsentRequest struct {
	Input    string    `json:"input"` // email or phone
	Purposes []Purpose `json:"purposes"`
	TenantID string    `json:"tenantId"`
}

// Implement the driver.Valuer interface
func (c ConsentPurposes) Value() (driver.Value, error) {
	return json.Marshal(c.Purposes)
}

// Implement the sql.Scanner interface
func (c *ConsentPurposes) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte for ConsentPurposes, got %T", value)
	}
	return json.Unmarshal(bytes, &c.Purposes)
}

type VendorConsentRequest struct {
	Input    string    `json:"input"`    // email/phone
	TenantID string    `json:"tenantId"` // tenant context
	Purposes []Purpose `json:"purposes"` // list of purpose names and statuses
}

type AdminConsentOverrideRequest struct {
	UID      string    `json:"uid"`
	TenantID string    `json:"tenantId"`
	Purposes []Purpose `json:"purposes"`
}

type CreateGrievanceRequest struct {
	UserID               string `json:"userId"`
	TenantID             string `json:"tenantId,omitempty"` // filled by server, not client
	GrievanceType        string `json:"grievanceType"`
	GrievanceSubject     string `json:"grievanceSubject"`
	GrievanceDescription string `json:"grievanceDescription"`
	Category             string `json:"category,omitempty"`
	Priority             string `json:"priority,omitempty"`
}

type UpdateGrievanceRequest struct {
	Status     string `json:"status" binding:"required,oneof=open in_progress escalated closed"`
	AssignedTo string `json:"assignedTo,omitempty" binding:"omitempty,uuid"`
}

type UpdateGrievanceDetailsRequest struct {
	GrievanceType        string `json:"grievanceType,omitempty"`
	GrievanceSubject     string `json:"grievanceSubject,omitempty"`
	GrievanceDescription string `json:"grievanceDescription,omitempty"`
	Category             string `json:"category,omitempty"`
	Priority             string `json:"priority,omitempty"`
}

type ListGrievanceRequest struct {
	UserID   string `json:"userId,omitempty" binding:"omitempty,uuid"`
	TenantID string `json:"tenantId" binding:"required,uuid"`
	Status   string `json:"status,omitempty" binding:"omitempty,oneof=open in_progress escalated closed"`
	Page     int    `json:"page" binding:"required,gte=1"`
	Limit    int    `json:"limit" binding:"required,gte=1,lte=100"`
}

type CreateGrievanceCommentRequest struct {
	GrievanceID string `json:"grievanceId" binding:"required,uuid"`
	UserID      string `json:"userId" binding:"required,uuid"`
	AdminID     string `json:"adminId,omitempty" binding:"omitempty,uuid"`
	Comment     string `json:"comment" binding:"required"`
}

type ReviewPageData struct {
	UID      string           `json:"uid"`
	TenantID uuid.UUID        `json:"tenantId"`
	Purposes []ConsentPurpose `json:"purposes"`
}

type ConsentHistoryEntry struct {
	Action       string    `json:"action"`
	Purposes     []Purpose `json:"purposes"`
	Timestamp    time.Time `json:"timestamp"`
	ChangedBy    string    `json:"changedBy"`
	GeoRegion    string    `json:"geoRegion,omitempty"`
	Jurisdiction string    `json:"jurisdiction,omitempty"`
}

type NotificationResponse struct {
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	Body    string    `json:"body"`
	Unread  bool      `json:"unread"`
	Created time.Time `json:"createdAt"`
}

type CreateDSRRequest struct {
	UserID   string `json:"userId" binding:"required,uuid"`
	TenantID string `json:"tenantId" binding:"required,uuid"`
	Type     string `json:"type" binding:"required,oneof=access delete rectify port restrict object"`
}

// Consent Form DTOs
type CreateConsentFormRequest struct {
	Name                 string                 `json:"name" binding:"required"`
	Title                string                 `json:"title" binding:"required"`
	Description          *string                `json:"description,omitempty"`
	Department           *string                `json:"department,omitempty"`
	Project              *string                `json:"project,omitempty"`
	OrganizationEntityID *string                `json:"organizationEntityId,omitempty"`
	DataRetentionPeriod  *string                `json:"dataRetentionPeriod,omitempty"`
	UserRightsSummary    *string                `json:"userRightsSummary,omitempty"`
	TermsAndConditions   *string                `json:"termsAndConditions,omitempty"`
	PrivacyPolicy        *string                `json:"privacyPolicy,omitempty"`
	Translations         map[string]interface{} `json:"translations,omitempty"`
	Regions              []string               `json:"regions,omitempty"`
}

type UpdateConsentFormRequest struct {
	Name                 *string                `json:"name,omitempty"`
	Title                *string                `json:"title,omitempty"`
	Description          *string                `json:"description,omitempty"`
	Department           *string                `json:"department,omitempty"`
	Project              *string                `json:"project,omitempty"`
	OrganizationEntityID *string                `json:"organizationEntityId,omitempty"`
	DataRetentionPeriod  *string                `json:"dataRetentionPeriod,omitempty"`
	UserRightsSummary    *string                `json:"userRightsSummary,omitempty"`
	TermsAndConditions   *string                `json:"termsAndConditions,omitempty"`
	PrivacyPolicy        *string                `json:"privacyPolicy,omitempty"`
	Translations         map[string]interface{} `json:"translations,omitempty"`
	Regions              []string               `json:"regions,omitempty"`
}

type AddPurposeToConsentFormRequest struct {
	PurposeID    string   `json:"purposeId" binding:"required,uuid"`
	DataObjects  []string `json:"dataObjects"`
	VendorIDs    []string `json:"vendorIds"`
	ExpiryInDays int      `json:"expiryInDays"`
}

type UpdatePurposeInConsentFormRequest struct {
	DataObjects  []string `json:"dataObjects"`
	VendorIDs    []string `json:"vendorIds"`
	ExpiryInDays int      `json:"expiryInDays"`
}

type ConsentFormPurposeResponse struct {
	PurposeID    string   `json:"purposeId"`
	PurposeName  string   `json:"purposeName"`
	DataObjects  []string `json:"dataObjects"`
	VendorIDs    []string `json:"vendorIds"`
	ExpiryInDays int      `json:"expiryInDays"`
}

type ConsentFormResponse struct {
	ID                      string                       `json:"id"`
	Name                    string                       `json:"name"`
	Title                   string                       `json:"title"`
	Description             string                       `json:"description"`
	DataCollectionAndUsage  string                       `json:"dataCollectionAndUsage"`
	DataSharingAndTransfers string                       `json:"dataSharingAndTransfers"`
	DataRetentionPeriod     string                       `json:"dataRetentionPeriod"`
	UserRightsSummary       string                       `json:"userRightsSummary"`
	TermsAndConditions      string                       `json:"termsAndConditions"`
	PrivacyPolicy           string                       `json:"privacyPolicy"`
	Purposes                []ConsentFormPurposeResponse `json:"purposes"`
	Translations            map[string]interface{}       `json:"translations,omitempty"`
	Regions                 []string                     `json:"regions,omitempty"`
	CreatedAt               time.Time                    `json:"createdAt"`
	UpdatedAt               time.Time                    `json:"updatedAt"`
}

type SubmitConsentRequest struct {
	UserID        string           `json:"userId" binding:"required,uuid"`
	ConsentFormID string           `json:"consentFormId" binding:"required,uuid"`
	Purposes      []PurposeConsent `json:"purposes"`
}

type PurposeConsent struct {
	PurposeID string `json:"purposeId"`
	Consented bool   `json:"consented"`
}

type IntegrationScriptResponse struct {
	Script string `json:"script"`
}

// Consent Form Validation and Versioning DTOs
type ValidateConsentFormResponse struct {
	IsValid bool              `json:"isValid"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Summary ValidationSummary `json:"summary"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type ValidationSummary struct {
	RequiredFieldsComplete bool `json:"requiredFieldsComplete"`
	PurposesAssigned       bool `json:"purposesAssigned"`
	DataObjectsValid       bool `json:"dataObjectsValid"`
	ExpirySettingsValid    bool `json:"expirySettingsValid"`
	NoDuplicatePurposes    bool `json:"noDuplicatePurposes"`
}

type PublishConsentFormRequest struct {
	ChangeLog string `json:"changeLog" binding:"required"`
}

type ConsentFormVersionResponse struct {
	ID            string                 `json:"id"`
	ConsentFormID string                 `json:"consentFormId"`
	VersionNumber int                    `json:"versionNumber"`
	Snapshot      map[string]interface{} `json:"snapshot"`
	PublishedAt   time.Time              `json:"publishedAt"`
	PublishedBy   string                 `json:"publishedBy"`
	Status        string                 `json:"status"`
	ChangeLog     string                 `json:"changeLog"`
	CreatedAt     time.Time              `json:"createdAt"`
}

type RollbackConsentFormRequest struct {
	VersionID string `json:"versionId" binding:"required,uuid"`
	Reason    string `json:"reason" binding:"required"`
}

type SubmitForReviewRequest struct {
	ReviewNotes string `json:"reviewNotes,omitempty"`
}

// Breach Notification DTOs
type CreateBreachNotificationRequest struct {
	Description        string    `json:"description" binding:"required"`
	BreachDate         time.Time `json:"breachDate" binding:"required"`
	DetectionDate      time.Time `json:"detectionDate" binding:"required"`
	AffectedUsersCount int       `json:"affectedUsersCount" binding:"required"`
	Status             string    `json:"status" binding:"required"`
}

type UpdateBreachNotificationRequest struct {
	Description          string    `json:"description"`
	BreachDate           time.Time `json:"breachDate"`
	DetectionDate        time.Time `json:"detectionDate"`
	AffectedUsersCount   int       `json:"affectedUsersCount"`
	Severity             string    `json:"severity"`
	BreachType           string    `json:"breachType"`
	Status               string    `json:"status"`
	RequiresDPBReporting bool      `json:"requiresDPBReporting"`
	RemedialActions      string    `json:"remedialActions"`
	PreventiveMeasures   string    `json:"preventiveMeasures"`
}

type BreachNotificationResponse struct {
	ID                 uuid.UUID  `json:"id"`
	TenantID           uuid.UUID  `json:"tenantId"`
	Description        string     `json:"description"`
	BreachDate         time.Time  `json:"breachDate"`
	DetectionDate      time.Time  `json:"detectionDate"`
	AffectedUsersCount int        `json:"affectedUsersCount"`
	NotifiedUsersCount int        `json:"notifiedUsersCount"`
	Status             string     `json:"status"`
	ReportedToDPB      bool       `json:"reportedToDpb"`
	ReportedToDPBDate  *time.Time `json:"reportedToDpbDate,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// SDK Configuration DTOs
type SDKTheme struct {
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	FontFamily     string `json:"fontFamily"`
	BorderRadius   int    `json:"borderRadius"`
	ButtonStyle    string `json:"buttonStyle"`
}

type CreateSDKConfigRequest struct {
	ConsentFormID        uuid.UUID `json:"consentFormId" binding:"required"`
	Theme                SDKTheme  `json:"theme"`
	Position             string    `json:"position"`
	Language             string    `json:"language"`
	ShowPreferenceCenter bool      `json:"showPreferenceCenter"`
	AutoShow             bool      `json:"autoShow"`
	CookieExpiry         int       `json:"cookieExpiry"`
	CustomCSS            string    `json:"customCss"`
}

type UpdateSDKConfigRequest struct {
	Theme                SDKTheme `json:"theme"`
	Position             string   `json:"position"`
	Language             string   `json:"language"`
	ShowPreferenceCenter bool     `json:"showPreferenceCenter"`
	AutoShow             bool     `json:"autoShow"`
	CookieExpiry         int      `json:"cookieExpiry"`
	CustomCSS            string   `json:"customCss"`
}

type SDKConfigResponse struct {
	ID                   uuid.UUID `json:"id"`
	TenantID             uuid.UUID `json:"tenantId"`
	ConsentFormID        uuid.UUID `json:"consentFormId"`
	Theme                SDKTheme  `json:"theme"`
	Position             string    `json:"position"`
	Language             string    `json:"language"`
	ShowPreferenceCenter bool      `json:"showPreferenceCenter"`
	AutoShow             bool      `json:"autoShow"`
	CookieExpiry         int       `json:"cookieExpiry"`
	CustomCSS            string    `json:"customCss"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type SDKGenerationResponse struct {
	SDKURL          string `json:"sdkUrl"`
	IntegrationCode string `json:"integrationCode"`
	PreviewURL      string `json:"previewUrl"`
}

// Cookie Management DTOs
type DetectedCookie struct {
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	Path         string `json:"path"`
	Value        string `json:"value"`
	ExpiryDays   int    `json:"expiryDays"`
	IsFirstParty bool   `json:"isFirstParty"`
	IsSecure     bool   `json:"isSecure"`
	IsHttpOnly   bool   `json:"isHttpOnly"`
	SameSite     string `json:"sameSite"`
	IsNew        bool   `json:"isNew"`
	Category     string `json:"category"`
	Purpose      string `json:"purpose"`
	Provider     string `json:"provider"`
}

type ScanWebsiteRequest struct {
	URL string `json:"url" binding:"required"`
}

type ScanWebsiteResponse struct {
	ScanID       uuid.UUID        `json:"scanId"`
	URL          string           `json:"url"`
	Status       string           `json:"status"`
	CookiesFound int              `json:"cookiesFound"`
	NewCookies   int              `json:"newCookies"`
	ScanDuration int              `json:"scanDuration"`
	Cookies      []DetectedCookie `json:"cookies"`
	CreatedAt    time.Time        `json:"createdAt"`
}

type CookieResponse struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenantId"`
	Name          string    `json:"name"`
	Domain        string    `json:"domain"`
	Path          string    `json:"path"`
	Category      string    `json:"category"`
	Purpose       string    `json:"purpose"`
	Provider      string    `json:"provider"`
	ExpiryDays    int       `json:"expiryDays"`
	IsFirstParty  bool      `json:"isFirstParty"`
	IsSecure      bool      `json:"isSecure"`
	IsHttpOnly    bool      `json:"isHttpOnly"`
	SameSite      string    `json:"sameSite"`
	Description   string    `json:"description"`
	DataCollected string    `json:"dataCollected"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type CreateCookieRequest struct {
	Name          string `json:"name" binding:"required"`
	Domain        string `json:"domain" binding:"required"`
	Path          string `json:"path"`
	Category      string `json:"category" binding:"required"`
	Purpose       string `json:"purpose"`
	Provider      string `json:"provider"`
	ExpiryDays    int    `json:"expiryDays"`
	IsFirstParty  bool   `json:"isFirstParty"`
	IsSecure      bool   `json:"isSecure"`
	IsHttpOnly    bool   `json:"isHttpOnly"`
	SameSite      string `json:"sameSite"`
	Description   string `json:"description"`
	DataCollected string `json:"dataCollected"`
}

type UpdateCookieRequest struct {
	Name          string `json:"name"`
	Domain        string `json:"domain"`
	Path          string `json:"path"`
	Category      string `json:"category"`
	Purpose       string `json:"purpose"`
	Provider      string `json:"provider"`
	ExpiryDays    int    `json:"expiryDays"`
	IsFirstParty  bool   `json:"isFirstParty"`
	IsSecure      bool   `json:"isSecure"`
	IsHttpOnly    bool   `json:"isHttpOnly"`
	SameSite      string `json:"sameSite"`
	Description   string `json:"description"`
	DataCollected string `json:"dataCollected"`
	IsActive      bool   `json:"isActive"`
}

type BulkCategorizeCookiesRequest struct {
	CookieIDs []uuid.UUID `json:"cookieIds" binding:"required"`
	Category  string      `json:"category" binding:"required"`
}

type CookieScanResponse struct {
	ID           uuid.UUID        `json:"id"`
	TenantID     uuid.UUID        `json:"tenantId"`
	URL          string           `json:"url"`
	ScanDate     time.Time        `json:"scanDate"`
	CookiesFound int              `json:"cookiesFound"`
	NewCookies   int              `json:"newCookies"`
	Status       string           `json:"status"`
	ScanDuration int              `json:"scanDuration"`
	ErrorMessage string           `json:"errorMessage,omitempty"`
	Cookies      []DetectedCookie `json:"cookies,omitempty"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

