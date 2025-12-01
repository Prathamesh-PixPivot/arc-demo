package handlers

import (
	"encoding/json"
	"net/http"
	"pixpivot/arc/internal/claims"
	"pixpivot/arc/internal/contextkeys"
	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/models"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type CookieHandler struct {
	cookieService        *services.CookieService
	cookieScannerService *services.CookieScannerService
	auditService         *services.AuditService
}

func NewCookieHandler(cookieService *services.CookieService, cookieScannerService *services.CookieScannerService, auditService *services.AuditService) *CookieHandler {
	return &CookieHandler{
		cookieService:        cookieService,
		cookieScannerService: cookieScannerService,
		auditService:         auditService,
	}
}

// Cookie CRUD endpoints
func (h *CookieHandler) CreateCookie(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	var req dto.CreateCookieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if validationErrors := h.cookieService.ValidateCookieData(&req); len(validationErrors) > 0 {
		writeError(w, http.StatusBadRequest, "validation failed: "+validationErrors[0])
		return
	}

	cookie, err := h.cookieService.CreateCookie(tenantID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, cookie.ID, "cookie_created", "created", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"cookie_name":   cookie.Name,
			"cookie_domain": cookie.Domain,
			"category":      cookie.Category,
		})
	}

	writeJSON(w, http.StatusCreated, h.convertToResponse(cookie))
}

func (h *CookieHandler) GetCookie(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	vars := mux.Vars(r)
	cookieID, err := uuid.Parse(vars["cookieId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cookie ID")
		return
	}

	cookie, err := h.cookieService.GetCookie(cookieID, tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Cookie not found")
		return
	}

	writeJSON(w, http.StatusOK, h.convertToResponse(cookie))
}

func (h *CookieHandler) UpdateCookie(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	vars := mux.Vars(r)
	cookieID, err := uuid.Parse(vars["cookieId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cookie ID")
		return
	}

	var req dto.UpdateCookieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cookie, err := h.cookieService.UpdateCookie(cookieID, tenantID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, cookie.ID, "cookie_updated", "updated", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"cookie_name":   cookie.Name,
			"cookie_domain": cookie.Domain,
			"category":      cookie.Category,
		})
	}

	writeJSON(w, http.StatusOK, h.convertToResponse(cookie))
}

func (h *CookieHandler) DeleteCookie(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	vars := mux.Vars(r)
	cookieID, err := uuid.Parse(vars["cookieId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid cookie ID")
		return
	}

	if err := h.cookieService.DeleteCookie(cookieID, tenantID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, cookieID, "cookie_deleted", "deleted", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"cookie_id": cookieID,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CookieHandler) ListCookies(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	// Parse query parameters
	category := r.URL.Query().Get("category")
	searchTerm := r.URL.Query().Get("search")
	provider := r.URL.Query().Get("provider")

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	var cookies []*models.Cookie
	var total int64

	if searchTerm != "" || provider != "" {
		cookies, total, err = h.cookieService.SearchCookies(tenantID, searchTerm, category, provider, limit, offset)
	} else {
		var isActive *bool
		if activeStr := r.URL.Query().Get("active"); activeStr != "" {
			if active, err := strconv.ParseBool(activeStr); err == nil {
				isActive = &active
			}
		}

		allCookies, err := h.cookieService.ListCookies(tenantID, category, isActive)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		total = int64(len(allCookies))

		// Apply pagination
		start := offset
		end := offset + limit
		if start > len(allCookies) {
			cookies = []*models.Cookie{}
		} else {
			if end > len(allCookies) {
				end = len(allCookies)
			}
			cookies = allCookies[start:end]
		}
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to response format
	responses := make([]dto.CookieResponse, len(cookies))
	for i, cookie := range cookies {
		responses[i] = h.convertToResponse(cookie)
	}

	result := map[string]interface{}{
		"cookies": responses,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *CookieHandler) BulkCategorizeCookies(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	var req dto.BulkCategorizeCookiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.cookieService.BulkUpdateCategories(tenantID, &req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, uuid.Nil, "cookies_bulk_categorized", "updated", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"cookie_count": len(req.CookieIDs),
			"category":     req.Category,
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "cookies categorized successfully"})
}

func (h *CookieHandler) GetCookieStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	stats, err := h.cookieService.GetCookieStats(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Cookie scanning endpoints
func (h *CookieHandler) ScanWebsite(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	var req dto.ScanWebsiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	scan, err := h.cookieScannerService.ScanWebsite(tenantID, req.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, scan.ID, "website_scan_started", "created", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"scan_url": req.URL,
			"scan_id":  scan.ID,
		})
	}

	response := dto.ScanWebsiteResponse{
		ScanID:    scan.ID,
		URL:       scan.URL,
		Status:    scan.Status,
		CreatedAt: scan.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *CookieHandler) GetScanResults(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	vars := mux.Vars(r)
	scanID, err := uuid.Parse(vars["scanId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid scan ID")
		return
	}

	scan, err := h.cookieScannerService.GetScanResults(scanID, tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Scan not found")
		return
	}

	response := h.convertScanToResponse(scan)
	writeJSON(w, http.StatusOK, response)
}

func (h *CookieHandler) GetScanHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID in claims")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	scans, err := h.cookieScannerService.GetScanHistory(tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]dto.CookieScanResponse, len(scans))
	for i, scan := range scans {
		responses[i] = h.convertScanToResponse(scan)
	}

	writeJSON(w, http.StatusOK, responses)
}

// Public endpoint for SDK to get allowed cookies
func (h *CookieHandler) GetAllowedCookies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["tenantId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	cookies, err := h.cookieService.GetAllowedCookiesForTenant(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to simplified format for SDK
	simplifiedCookies := make([]map[string]interface{}, len(cookies))
	for i, cookie := range cookies {
		simplifiedCookies[i] = map[string]interface{}{
			"name":     cookie.Name,
			"domain":   cookie.Domain,
			"category": cookie.Category,
			"required": cookie.Category == models.CookieCategoryNecessary,
		}
	}

	// Set CORS headers for public access
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cookies": simplifiedCookies,
	})
}

// Helper functions
func (h *CookieHandler) convertToResponse(cookie *models.Cookie) dto.CookieResponse {
	return dto.CookieResponse{
		ID:            cookie.ID,
		TenantID:      cookie.TenantID,
		Name:          cookie.Name,
		Domain:        cookie.Domain,
		Path:          cookie.Path,
		Category:      cookie.Category,
		Purpose:       cookie.Purpose,
		Provider:      cookie.Provider,
		ExpiryDays:    cookie.ExpiryDays,
		IsFirstParty:  cookie.IsFirstParty,
		IsSecure:      cookie.IsSecure,
		IsHttpOnly:    cookie.IsHttpOnly,
		SameSite:      cookie.SameSite,
		Description:   cookie.Description,
		DataCollected: cookie.DataCollected,
		IsActive:      cookie.IsActive,
		CreatedAt:     cookie.CreatedAt,
		UpdatedAt:     cookie.UpdatedAt,
	}
}

func (h *CookieHandler) convertScanToResponse(scan *models.CookieScan) dto.CookieScanResponse {
	response := dto.CookieScanResponse{
		ID:           scan.ID,
		TenantID:     scan.TenantID,
		URL:          scan.URL,
		ScanDate:     scan.ScanDate,
		CookiesFound: scan.CookiesFound,
		NewCookies:   scan.NewCookies,
		Status:       scan.Status,
		ScanDuration: scan.ScanDuration,
		ErrorMessage: scan.ErrorMessage,
		CreatedAt:    scan.CreatedAt,
		UpdatedAt:    scan.UpdatedAt,
	}

	// Parse results if available
	if len(scan.Results) > 0 {
		var result services.ScanResult
		if err := json.Unmarshal(scan.Results, &result); err == nil {
			response.Cookies = result.Cookies
		}
	}

	return response
}
