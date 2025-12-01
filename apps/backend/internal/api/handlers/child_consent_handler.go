package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"pixpivot/arc/internal/api/middleware"
	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/pkg/response"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ChildConsentHandler struct {
	service *services.ChildConsentService
}

func NewChildConsentHandler(service *services.ChildConsentService) *ChildConsentHandler {
	return &ChildConsentHandler{service: service}
}

func (h *ChildConsentHandler) AddChild(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	parentID, err := uuid.Parse(claims.PrincipalID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}
	
	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid tenant ID in token")
		return
	}

	var req struct {
		Name         string `json:"name"`
		DateOfBirth  string `json:"date_of_birth"` // YYYY-MM-DD
		Relationship string `json:"relationship"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid date of birth format (YYYY-MM-DD)")
		return
	}

	child, err := h.service.AddChild(parentID, tenantID, req.Name, dob, req.Relationship)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, child)
}

func (h *ChildConsentHandler) ListChildren(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	parentID, err := uuid.Parse(claims.PrincipalID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}
	
	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid tenant ID in token")
		return
	}

	children, err := h.service.ListChildren(parentID, tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, children)
}

func (h *ChildConsentHandler) CreateConsentRequest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid tenant ID in token")
		return
	}

	vars := mux.Vars(r)
	childID, err := uuid.Parse(vars["childId"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid child ID")
		return
	}

	var req struct {
		RequestType  string `json:"request_type"`
		ResourceName string `json:"resource_name"`
		PurposeID    string `json:"purpose_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var purposeID *uuid.UUID
	if req.PurposeID != "" {
		pID, err := uuid.Parse(req.PurposeID)
		if err == nil {
			purposeID = &pID
		}
	}

	consentReq, err := h.service.CreateConsentRequest(childID, tenantID, req.RequestType, req.ResourceName, purposeID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, consentReq)
}

func (h *ChildConsentHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	parentID, err := uuid.Parse(claims.PrincipalID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}
	
	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid tenant ID in token")
		return
	}

	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["requestId"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	if err := h.service.ApproveRequest(requestID, parentID, tenantID); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func (h *ChildConsentHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	parentID, err := uuid.Parse(claims.PrincipalID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}
	
	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid tenant ID in token")
		return
	}

	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["requestId"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	if err := h.service.RejectRequest(requestID, parentID, tenantID); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}
