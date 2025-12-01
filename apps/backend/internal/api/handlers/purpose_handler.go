package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"pixpivot/arc/internal/api/middleware"
	"pixpivot/arc/internal/claims"
	"pixpivot/arc/internal/contextkeys"
	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/internal/db"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type CreatePurposeRequest struct {
	Name                string         `json:"name"`
	Description         string         `json:"description"`
	DataObjects         pq.StringArray `json:"data_objects"`
	ReviewCycleMonths   int            `json:"review_cycle_months"`
	Vendors             []string       `json:"vendors"`
	IsThirdParty        bool           `json:"is_third_party"`
	Required            bool           `json:"required"`
	LegalBasis          string         `json:"legal_basis"`
	RetentionPeriodDays int            `json:"retention_period_days"`
	ParentPurposeID     *string        `json:"parent_purpose_id,omitempty"`
	TemplateID          *string        `json:"template_id,omitempty"`
}

type CreateFromTemplateRequest struct {
	TemplateID     string                 `json:"template_id"`
	Customizations map[string]interface{} `json:"customizations,omitempty"`
}

// PurposeHandler handles purpose-related requests
type PurposeHandler struct {
	DB                     *gorm.DB
	PurposeService         *services.PurposeService
	PurposeTemplateService *services.PurposeTemplateService
	PurposeRepository      *repository.PurposeRepository
}

func NewPurposeHandler(db *gorm.DB) *PurposeHandler {
	return &PurposeHandler{
		DB:                     db,
		PurposeService:         services.NewPurposeService(db),
		PurposeTemplateService: services.NewPurposeTemplateService(db),
		PurposeRepository:      repository.NewPurposeRepository(db),
	}
}

func CreatePurposeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetFiduciaryAuthClaims(r.Context())
		if claims == nil {
			writeError(w, http.StatusForbidden, "fiduciary access required")
			return
		}
		tenantID := claims.TenantID
		// Now, get your tenant db etc using tenantID
		schema := "tenant_" + tenantID[:8]
		dbConn, err := db.GetTenantDB(schema)
		if err != nil || dbConn == nil {
			writeError(w, http.StatusInternalServerError, "tenant db not found")
			return
		}

		var req CreatePurposeRequest
		// IsThirdParty is optional, default to false
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid payload")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		if req.Description == "" {
			writeError(w, http.StatusBadRequest, "description is required")
			return
		}

		//if IsThirdParty is true then vendors must be provided
		if req.IsThirdParty && len(req.Vendors) == 0 {
			writeError(w, http.StatusBadRequest, "vendors are required for third-party purposes")
			return
		}

		// if IsThirdParty is false then vendors must be empty
		if !req.IsThirdParty && len(req.Vendors) > 0 {
			writeError(w, http.StatusBadRequest, "vendors must be empty for non-third-party purposes")
			return
		}

		purpose := &models.Purpose{
			ID:                  uuid.New(),
			Name:                req.Name,
			Description:         req.Description,
			Vendors:             req.Vendors,
			ReviewCycleMonths:   req.ReviewCycleMonths,
			IsThirdParty:        req.IsThirdParty,
			Required:            req.Required,
			LegalBasis:          req.LegalBasis,
			RetentionPeriodDays: req.RetentionPeriodDays,
			Active:              true,
			TenantID:            uuid.MustParse(tenantID),
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		// Set parent purpose if provided
		if req.ParentPurposeID != nil && *req.ParentPurposeID != "" {
			parentID, err := uuid.Parse(*req.ParentPurposeID)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid parent purpose ID")
				return
			}
			purpose.ParentPurposeID = &parentID
		}

		// Set template if provided
		if req.TemplateID != nil && *req.TemplateID != "" {
			templateID, err := uuid.Parse(*req.TemplateID)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid template ID")
				return
			}
			purpose.TemplateID = &templateID
		}

		if err := dbConn.Create(purpose).Error; err != nil {
			log.Printf("[ERROR] Failed to create purpose: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to create purpose")
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(purpose)
	}
}

func (h *PurposeHandler) ToggleActive(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}
	tenantID := claims.TenantID
	// Now, get your tenant db etc using tenantID
	schema := "tenant_" + tenantID[:8]
	dbConn, err := db.GetTenantDB(schema)
	if err != nil || dbConn == nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}
	idStr := mux.Vars(r)["id"]
	var purpose models.Purpose
	if err := dbConn.First(&purpose, "id = ?", idStr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "purpose not found")
		} else {
			writeError(w, http.StatusInternalServerError, "database error")
		}
		return
	}

	// Toggle the Active field
	purpose.Active = !purpose.Active
	if err := dbConn.Save(&purpose).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update purpose")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(purpose)
}

func ListPurposesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetFiduciaryAuthClaims(r.Context())
		if claims == nil {
			writeError(w, http.StatusForbidden, "fiduciary access required")
			return
		}
		tenantID := claims.TenantID
		// Now, get your tenant db etc using tenantID
		schema := "tenant_" + tenantID[:8]
		dbConn, err := db.GetTenantDB(schema)
		if err != nil || dbConn == nil {
			writeError(w, http.StatusInternalServerError, "tenant db not found")
			return
		}

		var purposes []models.Purpose
		if err := dbConn.Find(&purposes).Error; err != nil {
			log.Printf("[ERROR] Failed to list purposes: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to list purposes")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(purposes)
	}
}

// Get /api/v1/user/purposes/{id}
func UserGetPurposeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(contextkeys.UserClaimsKey).(*claims.DataPrincipalClaims)
		if !ok {
			writeError(w, http.StatusForbidden, "user access required")
			return
		}

		// 3) tenant lookup
		tid, err := uuid.Parse(claims.TenantID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid tenant id in claims")
			return
		}

		schema := "tenant_" + tid.String()[:8]
		tenantDB, err := db.GetTenantDB(schema)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "tenant DB not found")
			return
		}

		// 4) parse path param
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid purpose ID")
			return
		}

		// 5) fetch single purpose
		var purpose models.Purpose
		if err := tenantDB.
			Where("id = ? AND tenant_id = ?", id, tid).
			First(&purpose).
			Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, http.StatusNotFound, "purpose not found")
			} else {
				log.Printf("[ERROR] failed to fetch purpose: %v", err)
				writeError(w, http.StatusInternalServerError, "failed to fetch purpose")
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(purpose)
	}
}

// GET /api/v1/user/purposes/{tenantID}
func UserGetPurposeByTenant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(contextkeys.UserClaimsKey).(*claims.DataPrincipalClaims)
		if !ok {
			writeError(w, http.StatusForbidden, "user access required")
			return
		}

		vars := mux.Vars(r)
		idStr := vars["tenantID"]
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid tenant ID")
			return
		}

		// Authorize: ensure the user is accessing their own tenant's data
		if claims.TenantID != idStr {
			writeError(w, http.StatusForbidden, "you are not authorized to access this tenant's data")
			return
		}

		// 3) tenant db
		schema := "tenant_" + id.String()[:8]
		tenantDB, err := db.GetTenantDB(schema)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "tenant DB not found")
			return
		}
		// 4) fetch purposes for this tenant and user
		var purposes []models.Purpose
		if err := tenantDB.
			Where("tenant_id = ? AND active = true", id).
			Find(&purposes).Error; err != nil {
			log.Printf("[ERROR] failed to fetch purposes: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to fetch purposes")
			return
		}
		if len(purposes) == 0 {
			writeError(w, http.StatusNotFound, "no purposes found for this tenant")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(purposes); err != nil {
			log.Printf("[ERROR] failed to encode purposes: %v", err)
			http.Error(w, "failed to encode purposes", http.StatusInternalServerError)
			return
		}
	}
}

func DeletePurposeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetFiduciaryAuthClaims(r.Context())
		if claims == nil {
			writeError(w, http.StatusForbidden, "fiduciary access required")
			return
		}
		tenantID := claims.TenantID
		// Now, get your tenant db etc using tenantID
		schema := "tenant_" + tenantID[:8]
		dbConn, err := db.GetTenantDB(schema)
		if err != nil || dbConn == nil {
			writeError(w, http.StatusInternalServerError, "tenant db not found")
			return
		}

		purposeID := r.URL.Query().Get("id")
		if purposeID == "" {
			writeError(w, http.StatusBadRequest, "missing purpose ID")
			return
		}

		if err := dbConn.Where("id = ? AND tenant_id = ?", purposeID, tenantID).Delete(&models.Purpose{}).Error; err != nil {
			log.Printf("[ERROR] Failed to delete purpose: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to delete purpose")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func UpdatePurposeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetFiduciaryAuthClaims(r.Context())
		if claims == nil {
			writeError(w, http.StatusForbidden, "fiduciary access required")
			return
		}
		tenantID := claims.TenantID
		// Now, get your tenant db etc using tenantID
		schema := "tenant_" + tenantID[:8]
		dbConn, err := db.GetTenantDB(schema)
		if err != nil || dbConn == nil {
			writeError(w, http.StatusInternalServerError, "tenant db not found")
			return
		}

		var req CreatePurposeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid payload")
			return
		}

		purposeID := r.URL.Query().Get("id")
		if purposeID == "" {
			writeError(w, http.StatusBadRequest, "missing purpose ID")
			return
		}

		purpose := &models.Purpose{
			ID:          uuid.MustParse(purposeID),
			Name:        req.Name,
			Description: req.Description,
			Required:    req.Required,
			TenantID:    uuid.MustParse(tenantID),
			UpdatedAt:   time.Now(),
		}

		if err := dbConn.Model(&models.Purpose{}).Where("id = ? AND tenant_id = ?", purpose.ID, tenantID).Updates(purpose).Error; err != nil {
			log.Printf("[ERROR] Failed to update purpose: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to update purpose")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(purpose)
	}
}

// ==================== PURPOSE TEMPLATE ENDPOINTS ====================

// GET /api/v1/fiduciary/purpose-templates
func (h *PurposeHandler) ListPurposeTemplates(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	framework := r.URL.Query().Get("framework")
	category := r.URL.Query().Get("category")
	legalBasis := r.URL.Query().Get("legal_basis")
	search := r.URL.Query().Get("search")

	var templates []*models.PurposeTemplate
	var err error

	if search != "" {
		templates, err = h.PurposeTemplateService.SearchTemplates(search)
	} else if category != "" {
		templates, err = h.PurposeTemplateService.ListTemplatesByCategory(category)
	} else if legalBasis != "" {
		templates, err = h.PurposeTemplateService.GetTemplatesByLegalBasis(legalBasis)
	} else {
		templates, err = h.PurposeTemplateService.ListTemplates(framework)
	}

	if err != nil {
		log.Printf("[ERROR] Failed to list purpose templates: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list purpose templates")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(templates)
}

// GET /api/v1/fiduciary/purpose-templates/{id}
func (h *PurposeHandler) GetPurposeTemplate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	template, err := h.PurposeTemplateService.GetTemplate(id)
	if err != nil {
		log.Printf("[ERROR] Failed to get purpose template: %v", err)
		if err.Error() == "purpose template not found" {
			writeError(w, http.StatusNotFound, "purpose template not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to get purpose template")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(template)
}

// POST /api/v1/fiduciary/purposes/from-template/{templateId}
func (h *PurposeHandler) CreatePurposeFromTemplate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID := claims.TenantID
	schema := "tenant_" + tenantID[:8]
	dbConn, err := db.GetTenantDB(schema)
	if err != nil || dbConn == nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}

	vars := mux.Vars(r)
	templateIDStr := vars["templateId"]
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	var req CreateFromTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Create service with tenant DB
	templateService := services.NewPurposeTemplateService(dbConn)
	purpose, err := templateService.CreatePurposeFromTemplate(
		uuid.MustParse(tenantID), templateID, req.Customizations)
	if err != nil {
		log.Printf("[ERROR] Failed to create purpose from template: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create purpose from template")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(purpose)
}

// ==================== COMPLIANCE ENDPOINTS ====================

// GET /api/v1/fiduciary/purposes/{id}/compliance
func (h *PurposeHandler) GetPurposeCompliance(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID := claims.TenantID
	schema := "tenant_" + tenantID[:8]
	dbConn, err := db.GetTenantDB(schema)
	if err != nil || dbConn == nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid purpose ID")
		return
	}

	purposeService := services.NewPurposeService(dbConn)
	report, err := purposeService.ValidateCompliance(id)
	if err != nil {
		log.Printf("[ERROR] Failed to validate compliance: %v", err)
		if err.Error() == "purpose not found" {
			writeError(w, http.StatusNotFound, "purpose not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to validate compliance")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

// POST /api/v1/fiduciary/purposes/{id}/validate-compliance
func (h *PurposeHandler) ValidatePurposeCompliance(w http.ResponseWriter, r *http.Request) {
	// Same as GetPurposeCompliance but as POST for explicit validation trigger
	h.GetPurposeCompliance(w, r)
}

// ==================== HIERARCHY ENDPOINTS ====================

// GET /api/v1/fiduciary/purposes/{id}/children
func (h *PurposeHandler) GetPurposeChildren(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID := claims.TenantID
	schema := "tenant_" + tenantID[:8]
	dbConn, err := db.GetTenantDB(schema)
	if err != nil || dbConn == nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid purpose ID")
		return
	}

	purposeRepo := repository.NewPurposeRepository(dbConn)
	children, err := purposeRepo.GetChildPurposes(id, uuid.MustParse(tenantID))
	if err != nil {
		log.Printf("[ERROR] Failed to get child purposes: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get child purposes")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(children)
}

// GET /api/v1/fiduciary/purposes/{id}/tree
func (h *PurposeHandler) GetPurposeTree(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID := claims.TenantID
	schema := "tenant_" + tenantID[:8]
	dbConn, err := db.GetTenantDB(schema)
	if err != nil || dbConn == nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid purpose ID")
		return
	}

	purposeRepo := repository.NewPurposeRepository(dbConn)
	tree, err := purposeRepo.GetPurposeTree(id, uuid.MustParse(tenantID))
	if err != nil {
		log.Printf("[ERROR] Failed to get purpose tree: %v", err)
		if err.Error() == "root purpose not found" {
			writeError(w, http.StatusNotFound, "purpose not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to get purpose tree")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tree)
}

// GET /api/v1/fiduciary/purposes/{id}/usage-stats
func (h *PurposeHandler) GetPurposeUsageStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID := claims.TenantID
	schema := "tenant_" + tenantID[:8]
	dbConn, err := db.GetTenantDB(schema)
	if err != nil || dbConn == nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid purpose ID")
		return
	}

	purposeService := services.NewPurposeService(dbConn)
	stats, err := purposeService.GetPurposeUsageStats(id)
	if err != nil {
		log.Printf("[ERROR] Failed to get purpose usage stats: %v", err)
		if err.Error() == "purpose not found" {
			writeError(w, http.StatusNotFound, "purpose not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to get purpose usage stats")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// POST /api/v1/fiduciary/purposes/{id}/inherit/{parentId}
func (h *PurposeHandler) InheritFromParent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID := claims.TenantID
	schema := "tenant_" + tenantID[:8]
	dbConn, err := db.GetTenantDB(schema)
	if err != nil || dbConn == nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}

	vars := mux.Vars(r)
	childIDStr := vars["id"]

	childID, err := uuid.Parse(childIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid child purpose ID")
		return
	}

	purposeRepo := repository.NewPurposeRepository(dbConn)
	if _, err := purposeRepo.InheritDataObjects(childID); err != nil {
		log.Printf("[ERROR] Failed to inherit data objects: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to inherit data objects")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data objects inherited successfully"})
}

// GET /api/v1/fiduciary/purpose-templates/stats
func (h *PurposeHandler) GetTemplateStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	stats, err := h.PurposeTemplateService.GetTemplateStats()
	if err != nil {
		log.Printf("[ERROR] Failed to get template stats: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get template stats")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// GET /api/v1/fiduciary/purposes/forest
func (h *PurposeHandler) GetPurposeForest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetFiduciaryAuthClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusForbidden, "fiduciary access required")
		return
	}

	tenantID := claims.TenantID
	schema := "tenant_" + tenantID[:8]
	dbConn, err := db.GetTenantDB(schema)
	if err != nil || dbConn == nil {
		writeError(w, http.StatusInternalServerError, "tenant db not found")
		return
	}

	purposeRepo := repository.NewPurposeRepository(dbConn)
	forest, err := purposeRepo.GetPurposeHierarchy(uuid.MustParse(tenantID))
	if err != nil {
		log.Printf("[ERROR] Failed to get purpose forest: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get purpose forest")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(forest)
}
