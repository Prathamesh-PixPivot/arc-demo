package handlers

import (
	"encoding/json"
	"net/http"

	"pixpivot/arc/internal/api/middleware"
	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/internal/db"
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type DataDiscoveryHandler struct {
	Service *services.DataDiscoveryService
}

func NewDataDiscoveryHandler(db *gorm.DB) *DataDiscoveryHandler {
	return &DataDiscoveryHandler{
		Service: services.NewDataDiscoveryService(db),
	}
}

// POST /api/v1/discovery/sources
func (h *DataDiscoveryHandler) CreateDataSource(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Host        string `json:"host"`
		Port        int    `json:"port"`
		Database    string `json:"database"`
		Username    string `json:"username"`
		Password    string `json:"password"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	tenantID := uuid.MustParse(claims.TenantID)
	// Switch to tenant DB
	schema := "tenant_" + claims.TenantID[:8]
	tenantDB, err := db.GetTenantDB(schema)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}
	
	// Re-init service with tenant DB
	service := services.NewDataDiscoveryService(tenantDB)

	ds := &models.DataSource{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Type:        req.Type,
		Host:        req.Host,
		Port:        req.Port,
		Database:    req.Database,
		Username:    req.Username,
		Password:    req.Password, // In real app, encrypt this!
		Description: req.Description,
	}

	if err := service.CreateDataSource(ds); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ds)
}

// GET /api/v1/discovery/sources
func (h *DataDiscoveryHandler) ListDataSources(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID := uuid.MustParse(claims.TenantID)
	schema := "tenant_" + claims.TenantID[:8]
	tenantDB, err := db.GetTenantDB(schema)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}
	service := services.NewDataDiscoveryService(tenantDB)

	sources, err := service.ListDataSources(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, sources)
}

// POST /api/v1/discovery/sources/{id}/scan
func (h *DataDiscoveryHandler) StartScan(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	vars := mux.Vars(r)
	sourceID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid source id")
		return
	}

	tenantID := uuid.MustParse(claims.TenantID)
	schema := "tenant_" + claims.TenantID[:8]
	tenantDB, err := db.GetTenantDB(schema)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}
	service := services.NewDataDiscoveryService(tenantDB)

	job, err := service.StartScan(tenantID, sourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, job)
}

// GET /api/v1/discovery/jobs/{id}
func (h *DataDiscoveryHandler) GetJobResults(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	schema := "tenant_" + claims.TenantID[:8]
	tenantDB, err := db.GetTenantDB(schema)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}
	service := services.NewDataDiscoveryService(tenantDB)

	results, err := service.GetJobResults(jobID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// GET /api/v1/discovery/dashboard
func (h *DataDiscoveryHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID := uuid.MustParse(claims.TenantID)
	schema := "tenant_" + claims.TenantID[:8]
	tenantDB, err := db.GetTenantDB(schema)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}
	service := services.NewDataDiscoveryService(tenantDB)

	stats, err := service.GetDashboardStats(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

