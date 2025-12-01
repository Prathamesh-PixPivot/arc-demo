package handlers

import (
	"encoding/json"
	"net/http"

	"pixpivot/arc/internal/claims"
	contextKey "pixpivot/arc/internal/contextkeys"
	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/pkg/log"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type EnhancedBreachNotificationHandler struct {
	service *services.EnhancedBreachNotificationService
}

func NewEnhancedBreachNotificationHandler(
	service *services.EnhancedBreachNotificationService,
	auditService *services.AuditService,
) *EnhancedBreachNotificationHandler {
	return &EnhancedBreachNotificationHandler{
		service: service,
	}
}

// CreateBreachNotification creates a new breach with workflow
func (h *EnhancedBreachNotificationHandler) CreateBreachNotification(w http.ResponseWriter, r *http.Request) {
	var breach models.BreachNotification
	if err := json.NewDecoder(r.Body).Decode(&breach); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims := r.Context().Value(contextKey.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	tenantID, _ := uuid.Parse(claims.TenantID)
	createdBy, _ := uuid.Parse(claims.FiduciaryID)

	breach.TenantID = tenantID

	if err := h.service.CreateBreachWithWorkflow(&breach, createdBy); err != nil {
		log.Logger.Error().Err(err).Msg("failed to create breach")
		writeError(w, http.StatusInternalServerError, "failed to create breach notification")
		return
	}

	writeJSON(w, http.StatusCreated, breach)
}

// SubmitForVerification submits breach for verification
func (h *EnhancedBreachNotificationHandler) SubmitForVerification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	breachID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid breach ID")
		return
	}

	claims := r.Context().Value(contextKey.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	submittedBy, _ := uuid.Parse(claims.FiduciaryID)

	if err := h.service.SubmitForVerification(breachID, submittedBy); err != nil {
		log.Logger.Error().Err(err).Msg("failed to submit for verification")
		writeError(w, http.StatusInternalServerError, "failed to submit for verification")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "submitted for verification"})
}

// VerifyBreach verifies or rejects a breach
func (h *EnhancedBreachNotificationHandler) VerifyBreach(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	breachID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid breach ID")
		return
	}

	var req struct {
		Approved        bool   `json:"approved"`
		RejectionReason string `json:"rejection_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims := r.Context().Value(contextKey.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	verifiedBy, _ := uuid.Parse(claims.FiduciaryID)

	if err := h.service.VerifyBreach(breachID, verifiedBy, req.Approved, req.RejectionReason); err != nil {
		log.Logger.Error().Err(err).Msg("failed to verify breach")
		writeError(w, http.StatusInternalServerError, "failed to verify breach")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "breach verification completed"})
}

// ApproveDataPrincipalNotification approves notification to affected individuals
func (h *EnhancedBreachNotificationHandler) ApproveDataPrincipalNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	breachID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid breach ID")
		return
	}

	claims := r.Context().Value(contextKey.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	approvedBy, _ := uuid.Parse(claims.FiduciaryID)

	if err := h.service.ApproveDataPrincipalNotification(breachID, approvedBy); err != nil {
		log.Logger.Error().Err(err).Msg("failed to approve data principal notification")
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "data principal notification approved"})
}

// SendDPBNotification sends notification to DPB
func (h *EnhancedBreachNotificationHandler) SendDPBNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	breachID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid breach ID")
		return
	}

	claims := r.Context().Value(contextKey.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	sentBy, _ := uuid.Parse(claims.FiduciaryID)

	if err := h.service.SendDPBNotification(breachID, sentBy); err != nil {
		log.Logger.Error().Err(err).Msg("failed to send DPB notification")
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "DPB notification sent"})
}

// SendDataPrincipalNotifications sends notifications to affected individuals
func (h *EnhancedBreachNotificationHandler) SendDataPrincipalNotifications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	breachID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid breach ID")
		return
	}

	var req struct {
		AffectedEmails []string `json:"affected_emails"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims := r.Context().Value(contextKey.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	sentBy, _ := uuid.Parse(claims.FiduciaryID)

	if err := h.service.SendDataPrincipalNotifications(breachID, req.AffectedEmails, sentBy); err != nil {
		log.Logger.Error().Err(err).Msg("failed to send data principal notifications")
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "data principal notifications sent"})
}

// CheckSLACompliance checks SLA status
func (h *EnhancedBreachNotificationHandler) CheckSLACompliance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	breachID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid breach ID")
		return
	}

	slaStatus, err := h.service.CheckSLACompliance(breachID)
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to check SLA compliance")
		writeError(w, http.StatusInternalServerError, "failed to check SLA compliance")
		return
	}

	writeJSON(w, http.StatusOK, slaStatus)
}

// GetBreachRegister returns breach register for compliance
func (h *EnhancedBreachNotificationHandler) GetBreachRegister(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(contextKey.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
	tenantID, _ := uuid.Parse(claims.TenantID)

	breaches, err := h.service.GetBreachRegister(tenantID)
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to get breach register")
		writeError(w, http.StatusInternalServerError, "failed to get breach register")
		return
	}

	writeJSON(w, http.StatusOK, breaches)
}
