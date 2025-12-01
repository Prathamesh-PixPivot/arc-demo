package handlers

import (
	"pixpivot/arc/internal/claims"
	"pixpivot/arc/internal/contextkeys"
	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/core/services"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type SDKHandler struct {
	service      *services.SDKGeneratorService
	auditService *services.AuditService
}

func NewSDKHandler(service *services.SDKGeneratorService, auditService *services.AuditService) *SDKHandler {
	return &SDKHandler{
		service:      service,
		auditService: auditService,
	}
}

// Public SDK endpoint - serves the generated JavaScript SDK
func (h *SDKHandler) GetSDK(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["tenantId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	formIDStr := strings.TrimSuffix(vars["formId"], ".js")
	formID, err := uuid.Parse(formIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	// Get SDK configuration
	config, err := h.service.GetSDKConfigByFormID(tenantID, formID)
	if err != nil {
		writeError(w, http.StatusNotFound, "SDK configuration not found")
		return
	}

	// Generate SDK
	sdk, err := h.service.GenerateSDK(tenantID, formID, config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate SDK")
		return
	}

	// Set appropriate headers for JavaScript
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Enable gzip compression
	w.Header().Set("Content-Encoding", "gzip")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(sdk))
}

// Get SDK configuration for a specific form
func (h *SDKHandler) GetSDKConfig(w http.ResponseWriter, r *http.Request) {
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
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	config, err := h.service.GetSDKConfigByFormID(tenantID, formID)
	if err != nil {
		writeError(w, http.StatusNotFound, "SDK configuration not found")
		return
	}

	// Convert to response DTO
	var theme dto.SDKTheme
	if err := json.Unmarshal(config.Theme, &theme); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse theme")
		return
	}

	response := &dto.SDKConfigResponse{
		ID:                   config.ID,
		TenantID:             config.TenantID,
		ConsentFormID:        config.ConsentFormID,
		Theme:                theme,
		Position:             config.Position,
		Language:             config.Language,
		ShowPreferenceCenter: config.ShowPreferenceCenter,
		AutoShow:             config.AutoShow,
		CookieExpiry:         config.CookieExpiry,
		CustomCSS:            config.CustomCSS,
		CreatedAt:            config.CreatedAt,
		UpdatedAt:            config.UpdatedAt,
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, formID, "sdk_config_accessed", "accessed", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":   formID,
			"config_id": config.ID,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

// Create SDK configuration
func (h *SDKHandler) CreateSDKConfig(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateSDKConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	config, err := h.service.CreateSDKConfig(tenantID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to response DTO
	var theme dto.SDKTheme
	if err := json.Unmarshal(config.Theme, &theme); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse theme")
		return
	}

	response := &dto.SDKConfigResponse{
		ID:                   config.ID,
		TenantID:             config.TenantID,
		ConsentFormID:        config.ConsentFormID,
		Theme:                theme,
		Position:             config.Position,
		Language:             config.Language,
		ShowPreferenceCenter: config.ShowPreferenceCenter,
		AutoShow:             config.AutoShow,
		CookieExpiry:         config.CookieExpiry,
		CustomCSS:            config.CustomCSS,
		CreatedAt:            config.CreatedAt,
		UpdatedAt:            config.UpdatedAt,
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, config.ConsentFormID, "sdk_config_created", "created", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":   config.ConsentFormID,
			"config_id": config.ID,
		})
	}

	writeJSON(w, http.StatusCreated, response)
}

// Update SDK configuration
func (h *SDKHandler) UpdateSDKConfig(w http.ResponseWriter, r *http.Request) {
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
	configID, err := uuid.Parse(vars["configId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	var req dto.UpdateSDKConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	config, err := h.service.UpdateSDKConfig(configID, tenantID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to response DTO
	var theme dto.SDKTheme
	if err := json.Unmarshal(config.Theme, &theme); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse theme")
		return
	}

	response := &dto.SDKConfigResponse{
		ID:                   config.ID,
		TenantID:             config.TenantID,
		ConsentFormID:        config.ConsentFormID,
		Theme:                theme,
		Position:             config.Position,
		Language:             config.Language,
		ShowPreferenceCenter: config.ShowPreferenceCenter,
		AutoShow:             config.AutoShow,
		CookieExpiry:         config.CookieExpiry,
		CustomCSS:            config.CustomCSS,
		CreatedAt:            config.CreatedAt,
		UpdatedAt:            config.UpdatedAt,
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, config.ConsentFormID, "sdk_config_updated", "updated", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":   config.ConsentFormID,
			"config_id": config.ID,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

// Get integration code for a form
func (h *SDKHandler) GetIntegrationCode(w http.ResponseWriter, r *http.Request) {
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
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	response, err := h.service.GenerateIntegrationCode(tenantID, formID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, formID, "integration_code_accessed", "accessed", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id": formID,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

// Delete SDK configuration
func (h *SDKHandler) DeleteSDKConfig(w http.ResponseWriter, r *http.Request) {
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
	configID, err := uuid.Parse(vars["configId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	if err := h.service.DeleteSDKConfig(configID, tenantID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging
	if h.auditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.auditService.Create(r.Context(), fiduciaryID, tenantID, uuid.Nil, "sdk_config_deleted", "deleted", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"config_id": configID,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

