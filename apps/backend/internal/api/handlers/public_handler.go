package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// PublicHandler handles public-facing endpoints (web-only features)
type PublicHandler struct {
	consentService *services.UserConsentService
	receiptService *services.ReceiptService
	cookieService  *services.CookieService
	auditService   *services.AuditService
}

// NewPublicHandler creates a new public handler
func NewPublicHandler(
	consentService *services.UserConsentService,
	receiptService *services.ReceiptService,
	cookieService *services.CookieService,
) *PublicHandler {
	return &PublicHandler{
		consentService: consentService,
		receiptService: receiptService,
		cookieService:  cookieService,
	}
}

// SubmitPublicConsent handles public consent submissions from websites
// POST /api/v1/public/consent
func (h *PublicHandler) SubmitPublicConsent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TenantID      string                 `json:"tenant_id"`
		WebsiteID     string                 `json:"website_id,omitempty"`
		VisitorID     string                 `json:"visitor_id,omitempty"`
		Email         string                 `json:"email"`
		Name          string                 `json:"name,omitempty"`
		Phone         string                 `json:"phone,omitempty"`
		ConsentFormID string                 `json:"consent_form_id"`
		Purposes      []string               `json:"purposes"`
		DataObjects   []string               `json:"data_objects,omitempty"`
		Channel       string                 `json:"channel"`
		Metadata      map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if request.TenantID == "" || request.Email == "" || request.ConsentFormID == "" {
		writeError(w, http.StatusBadRequest, "Missing required fields: tenant_id, email, consent_form_id")
		return
	}

	tenantID, err := uuid.Parse(request.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	consentFormID, err := uuid.Parse(request.ConsentFormID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid consent form ID")
		return
	}

	// Get client IP and user agent
	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	referrer := r.Header.Get("Referer")

	// Create or find data principal
	principal := &models.DataPrincipal{
		Email:     request.Email,
		FirstName: request.Name, // Using Name as FirstName for simplicity
		Phone:     request.Phone,
	}

	// Create consent request
	consentReq := &dto.CreateUserConsentRequest{
		ConsentFormID: consentFormID,
		Purposes:      request.Purposes,
		DataObjects:   request.DataObjects,
		Channel:       request.Channel,
		IPAddress:     clientIP,
		UserAgent:     userAgent,
		Metadata:      request.Metadata,
	}

	// Add referrer to metadata
	if consentReq.Metadata == nil {
		consentReq.Metadata = make(map[string]interface{})
	}
	consentReq.Metadata["referrer"] = referrer
	consentReq.Metadata["visitor_id"] = request.VisitorID
	consentReq.Metadata["website_id"] = request.WebsiteID

	// Submit consent
	consent, err := h.consentService.CreatePublicConsent(r.Context(), tenantID, principal, consentReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create consent")
		return
	}

	// Generate receipt asynchronously
	go func() {
		_, err := h.receiptService.GenerateReceipt(consent.ID)
		if err != nil {
			// Log error but don't fail the consent submission
			// TODO: Add proper logging
		}
	}()

	// Return success response
	response := map[string]interface{}{
		"consent_id": consent.ID,
		"status":     "success",
		"message":    "Consent recorded successfully",
		"timestamp":  time.Now(),
	}

	writeJSON(w, http.StatusCreated, response)
}

// GetPublicConsentForm handles fetching public consent forms
// GET /api/v1/public/consent/{formId}
func (h *PublicHandler) GetPublicConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formIDStr := vars["formId"]

	formID, err := uuid.Parse(formIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	// Get tenant ID from query parameter or header
	tenantIDStr := r.URL.Query().Get("tenant_id")
	if tenantIDStr == "" {
		tenantIDStr = r.Header.Get("X-Tenant-ID")
	}

	if tenantIDStr == "" {
		writeError(w, http.StatusBadRequest, "Tenant ID required")
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// TODO: Implement GetPublicConsentForm in consent service
	form, err := h.consentService.GetPublicConsentForm(r.Context(), tenantID, formID)

	writeJSON(w, http.StatusOK, form)
}

// GetCookieSettings handles fetching cookie settings for websites
// GET /api/v1/public/cookies/{tenantId}
func (h *PublicHandler) GetCookieSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		domain = r.Header.Get("Origin")
		if domain != "" {
			// Remove protocol from origin
			if strings.HasPrefix(domain, "http://") {
				domain = domain[7:]
			} else if strings.HasPrefix(domain, "https://") {
				domain = domain[8:]
			}
		}
	}

	// Get cookie settings for the domain
	cookies, err := h.cookieService.GetPublicCookieSettings(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get cookie settings")
		return
	}

	writeJSON(w, http.StatusOK, cookies)
}

// SubmitCookieConsent handles cookie consent submissions
// POST /api/v1/public/cookie-consent
func (h *PublicHandler) SubmitCookieConsent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TenantID    string            `json:"tenant_id"`
		VisitorID   string            `json:"visitor_id"`
		Domain      string            `json:"domain"`
		Consents    map[string]bool   `json:"consents"` // category -> allowed
		CookieIDs   []string          `json:"cookie_ids,omitempty"`
		Preferences map[string]string `json:"preferences,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tenantID, err := uuid.Parse(request.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	// Get client information
	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Submit cookie consent
	consentData := &dto.PublicCookieConsentRequest{
		VisitorID:   request.VisitorID,
		Domain:      request.Domain,
		Consents:    request.Consents,
		CookieIDs:   request.CookieIDs,
		Preferences: request.Preferences,
		IPAddress:   clientIP,
		UserAgent:   userAgent,
	}
	err = h.cookieService.SubmitPublicCookieConsent(tenantID, consentData)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to submit cookie consent")
		return
	}

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Cookie consent recorded",
		"timestamp": time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// VerifyReceipt handles public receipt verification
// GET /api/v1/public/receipts/verify/{receiptNumber}
func (h *PublicHandler) VerifyReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	receiptNumber := vars["receiptNumber"]

	if receiptNumber == "" {
		writeError(w, http.StatusBadRequest, "Receipt number required")
		return
	}

	// Verify receipt
	verification, err := h.receiptService.VerifyReceipt(receiptNumber)
	if err != nil {
		writeError(w, http.StatusNotFound, "Receipt not found or invalid")
		return
	}

	writeJSON(w, http.StatusOK, verification)
}

// GetPrivacyPolicy handles fetching privacy policy
// GET /api/v1/public/privacy-policy/{tenantId}
func (h *PublicHandler) GetPrivacyPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	language := r.URL.Query().Get("lang")
	if language == "" {
		language = "en"
	}

	// TODO: Implement GetPrivacyPolicy in a service
	// For now, return a mock response
	policy := map[string]interface{}{
		"tenant_id":    tenantID,
		"language":     language,
		"title":        "Privacy Policy",
		"content":      "This is the privacy policy content...",
		"last_updated": time.Now().AddDate(0, -1, 0),
		"version":      "1.0",
	}

	writeJSON(w, http.StatusOK, policy)
}

// GetTermsOfService handles fetching terms of service
// GET /api/v1/public/terms/{tenantId}
func (h *PublicHandler) GetTermsOfService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	language := r.URL.Query().Get("lang")
	if language == "" {
		language = "en"
	}

	// TODO: Implement GetTermsOfService in a service
	terms := map[string]interface{}{
		"tenant_id":    tenantID,
		"language":     language,
		"title":        "Terms of Service",
		"content":      "These are the terms of service...",
		"last_updated": time.Now().AddDate(0, -1, 0),
		"version":      "1.0",
	}

	writeJSON(w, http.StatusOK, terms)
}

// SubmitDSRRequest handles public DSR request submissions
// POST /api/v1/public/dsr/request
func (h *PublicHandler) SubmitDSRRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TenantID     string `json:"tenant_id"`
		Email        string `json:"email"`
		Name         string `json:"name,omitempty"`
		Phone        string `json:"phone,omitempty"`
		RequestType  string `json:"request_type"`
		Description  string `json:"description,omitempty"`
		Verification string `json:"verification,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if request.TenantID == "" || request.Email == "" || request.RequestType == "" {
		writeError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	tenantID, err := uuid.Parse(request.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}
	_ = tenantID

	// TODO: Implement SubmitPublicDSRRequest in DSR service
	/*
		dsrRequest := &dto.CreateDSRRequest{
			UserID:       request.Email,
			TenantID:     tenantID.String(),
			Type:         request.RequestType,
		}
		_ = dsrRequest
	*/
	requestID := uuid.New()

	response := map[string]interface{}{
		"request_id": requestID,
		"status":     "submitted",
		"message":    "DSR request submitted successfully",
		"reference":  fmt.Sprintf("DSR-%s", requestID.String()[:8]),
		"timestamp":  time.Now(),
	}

	writeJSON(w, http.StatusCreated, response)
}

