package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"
	"pixpivot/arc/pkg/log"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type SuperAdminHandler struct {
	service *services.SuperAdminService
}

func NewSuperAdminHandler(service *services.SuperAdminService) *SuperAdminHandler {
	return &SuperAdminHandler{service: service}
}

// Licenses

func (h *SuperAdminHandler) GenerateLicense(w http.ResponseWriter, r *http.Request) {
	var req services.GenerateLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	license, err := h.service.GenerateLicense(req)
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to generate license")
		writeError(w, http.StatusInternalServerError, "failed to generate license")
		return
	}

	writeJSON(w, http.StatusCreated, license)
}

func (h *SuperAdminHandler) ListLicenses(w http.ResponseWriter, r *http.Request) {
	params := repository.LicenseListParams{
		Page:   1,
		Limit:  20,
		Search: r.URL.Query().Get("search"),
		Type:   r.URL.Query().Get("type"),
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
		}
	}
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		active := activeStr == "true"
		params.IsActive = &active
	}

	resp, err := h.service.ListLicenses(params)
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to list licenses")
		writeError(w, http.StatusInternalServerError, "failed to list licenses")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SuperAdminHandler) RevokeLicense(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid license ID")
		return
	}

	if err := h.service.RevokeLicense(id); err != nil {
		log.Logger.Error().Err(err).Msg("failed to revoke license")
		writeError(w, http.StatusInternalServerError, "failed to revoke license")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Tenants

func (h *SuperAdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	params := repository.TenantListParams{
		Page:   1,
		Limit:  20,
		Search: r.URL.Query().Get("search"),
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
		}
	}

	resp, err := h.service.ListTenants(params)
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to list tenants")
		writeError(w, http.StatusInternalServerError, "failed to list tenants")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SuperAdminHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var tenant models.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.CreateTenant(&tenant); err != nil {
		log.Logger.Error().Err(err).Msg("failed to create tenant")
		writeError(w, http.StatusInternalServerError, "failed to create tenant")
		return
	}

	writeJSON(w, http.StatusCreated, tenant)
}

func (h *SuperAdminHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	var tenant models.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tenant.TenantID = id

	if err := h.service.UpdateTenant(&tenant); err != nil {
		log.Logger.Error().Err(err).Msg("failed to update tenant")
		writeError(w, http.StatusInternalServerError, "failed to update tenant")
		return
	}

	writeJSON(w, http.StatusOK, tenant)
}

func (h *SuperAdminHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	if err := h.service.DeleteTenant(id); err != nil {
		log.Logger.Error().Err(err).Msg("failed to delete tenant")
		writeError(w, http.StatusInternalServerError, "failed to delete tenant")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
