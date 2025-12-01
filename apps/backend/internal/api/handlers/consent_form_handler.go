package handlers

import (
	"pixpivot/arc/internal/claims"
	"pixpivot/arc/internal/contextkeys"
	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/core/services"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ConsentFormHandler struct {
	service      *services.ConsentFormService
	AuditService *services.AuditService
}

func NewConsentFormHandler(service *services.ConsentFormService, auditService *services.AuditService) *ConsentFormHandler {
	return &ConsentFormHandler{service: service, AuditService: auditService}
}

func (h *ConsentFormHandler) CreateConsentForm(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateConsentFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	form, err := h.service.CreateConsentForm(tenantID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for consent form creation
	if h.AuditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, form.ID, "consent_form_created", "created", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"title":       form.Title,
			"description": form.Description,
		})
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(form)
}

func (h *ConsentFormHandler) UpdateConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	var req dto.UpdateConsentFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	form, err := h.service.UpdateConsentForm(formID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for consent form update
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_updated", "updated", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"title":       form.Title,
			"description": form.Description,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(form)
}

func (h *ConsentFormHandler) DeleteConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	if err := h.service.DeleteConsentForm(formID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for consent form deletion
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_deleted", "deleted", claims.FiduciaryID, r.RemoteAddr, "", "", nil)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConsentFormHandler) GetConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	form, err := h.service.GetConsentFormByID(formID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for consent form access
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_accessed", "accessed", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"title": form.Title,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(form)
}

func (h *ConsentFormHandler) ListConsentForms(w http.ResponseWriter, r *http.Request) {
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

	forms, err := h.service.ListConsentForms(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for consent forms list access
	if h.AuditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, tenantID, "consent_forms_list_accessed", "accessed", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"count": len(forms),
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(forms)
}

func (h *ConsentFormHandler) AddPurposeToConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	var req dto.AddPurposeToConsentFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	formPurpose, err := h.service.AddPurposeToConsentForm(formID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for adding purpose to consent form
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "purpose_added_to_consent_form", "updated", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"purpose_id": req.PurposeID,
		})
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(formPurpose)
}

func (h *ConsentFormHandler) UpdatePurposeInConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}
	purposeID, err := uuid.Parse(vars["purposeId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid purpose ID")
		return
	}

	var req dto.UpdatePurposeInConsentFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	formPurpose, err := h.service.UpdatePurposeInConsentForm(formID, purposeID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for updating purpose in consent form
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "purpose_updated_in_consent_form", "updated", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"purpose_id": purposeID,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(formPurpose)
}

func (h *ConsentFormHandler) RemovePurposeFromConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}
	purposeID, err := uuid.Parse(vars["purposeId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid purpose ID")
		return
	}

	if err := h.service.RemovePurposeFromConsentForm(formID, purposeID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for removing purpose from consent form
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "purpose_removed_from_consent_form", "updated", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"purpose_id": purposeID,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConsentFormHandler) GetIntegrationScript(w http.ResponseWriter, r *http.Request) {
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

	// Redirect to SDK integration endpoint
	// This maintains backward compatibility while using the new SDK system
	script := map[string]string{
		"message": "Please use the new SDK endpoints: /api/v1/fiduciary/sdk-config/{formId} for configuration and /api/v1/fiduciary/integration-code/{formId} for integration code",
		"sdkConfigEndpoint": "/api/v1/fiduciary/sdk-config/" + formID.String(),
		"integrationCodeEndpoint": "/api/v1/fiduciary/integration-code/" + formID.String(),
	}

	// Audit logging for getting integration script
	if h.AuditService != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "integration_script_accessed", "accessed", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id": formID,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(script)
}

func (h *ConsentFormHandler) PublishConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	var req dto.PublishConsentFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	fiduciaryID, err := uuid.Parse(claims.FiduciaryID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid fiduciary ID in claims")
		return
	}

	if err := h.service.PublishConsentFormWithValidation(formID, fiduciaryID, req.ChangeLog); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for publishing consent form
	if h.AuditService != nil {
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_published", "published", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":    formID,
			"change_log": req.ChangeLog,
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "form published successfully"})
}

func (h *ConsentFormHandler) ValidateConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	validation, err := h.service.ValidateForPublish(formID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for validation
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_validated", "validated", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":  formID,
			"is_valid": validation.IsValid,
			"errors":   len(validation.Errors),
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(validation)
}

func (h *ConsentFormHandler) SubmitForReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	var req dto.SubmitForReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.SubmitForReview(formID, req.ReviewNotes); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for submit for review
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_submitted_for_review", "submitted_for_review", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":      formID,
			"review_notes": req.ReviewNotes,
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "form submitted for review successfully"})
}

func (h *ConsentFormHandler) GetVersionHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	versions, err := h.service.GetVersionHistory(formID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for version history access
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_version_history_accessed", "accessed", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":        formID,
			"version_count": len(versions),
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(versions)
}

func (h *ConsentFormHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}
	versionID, err := uuid.Parse(vars["versionId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid version ID")
		return
	}

	version, err := h.service.GetVersion(versionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Verify the version belongs to the form
	if version.ConsentFormID != formID.String() {
		writeError(w, http.StatusBadRequest, "Version does not belong to the specified form")
		return
	}

	// Audit logging for version access
	claims, _ := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if h.AuditService != nil && claims != nil {
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_version_accessed", "accessed", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":        formID,
			"version_id":     versionID,
			"version_number": version.VersionNumber,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(version)
}

func (h *ConsentFormHandler) RollbackToVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}
	versionID, err := uuid.Parse(vars["versionId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid version ID")
		return
	}

	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	var req dto.RollbackConsentFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	fiduciaryID, err := uuid.Parse(claims.FiduciaryID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid fiduciary ID in claims")
		return
	}

	if err := h.service.RollbackToVersion(formID, versionID, fiduciaryID, req.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for rollback
	if h.AuditService != nil {
		tenantID, _ := uuid.Parse(claims.TenantID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_rolled_back", "rolled_back", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":    formID,
			"version_id": versionID,
			"reason":     req.Reason,
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "form rolled back successfully"})
}

func (h *ConsentFormHandler) TranslateConsentForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	formID, err := uuid.Parse(vars["formId"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form ID")
		return
	}

	claims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	if !ok {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	var req struct {
		Languages []string `json:"languages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(req.Languages) == 0 {
		writeError(w, http.StatusBadRequest, "at least one language is required")
		return
	}

	if err := h.service.AutoTranslateConsentForm(r.Context(), formID, req.Languages); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit logging for translation
	if h.AuditService != nil {
		tenantID, _ := uuid.Parse(claims.TenantID)
		fiduciaryID, _ := uuid.Parse(claims.FiduciaryID)
		go h.AuditService.Create(r.Context(), fiduciaryID, tenantID, formID, "consent_form_translated", "updated", claims.FiduciaryID, r.RemoteAddr, "", "", map[string]interface{}{
			"form_id":   formID,
			"languages": req.Languages,
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "form translation initiated"})
}

