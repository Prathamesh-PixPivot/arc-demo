package repository

import (
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChildConsentRepository struct {
	db *gorm.DB
}

func NewChildConsentRepository(db *gorm.DB) *ChildConsentRepository {
	return &ChildConsentRepository{db: db}
}

// Child Profile Methods

func (r *ChildConsentRepository) CreateChildProfile(child *models.ChildProfile) error {
	return r.db.Create(child).Error
}

func (r *ChildConsentRepository) GetChildProfile(id, tenantID uuid.UUID) (*models.ChildProfile, error) {
	var child models.ChildProfile
	err := r.db.First(&child, "id = ? AND tenant_id = ?", id, tenantID).Error
	return &child, err
}

func (r *ChildConsentRepository) ListChildrenByParent(parentID, tenantID uuid.UUID) ([]models.ChildProfile, error) {
	var children []models.ChildProfile
	err := r.db.Where("parent_id = ? AND tenant_id = ? AND is_active = ?", parentID, tenantID, true).Find(&children).Error
	return children, err
}

func (r *ChildConsentRepository) UpdateChildProfile(child *models.ChildProfile) error {
	// Ensure we don't accidentally update across tenants if the struct was modified
	return r.db.Where("id = ? AND tenant_id = ?", child.ID, child.TenantID).Save(child).Error
}

// Parental Consent Request Methods

func (r *ChildConsentRepository) CreateConsentRequest(req *models.ParentalConsentRequest) error {
	return r.db.Create(req).Error
}

func (r *ChildConsentRepository) GetConsentRequest(id, tenantID uuid.UUID) (*models.ParentalConsentRequest, error) {
	var req models.ParentalConsentRequest
	err := r.db.First(&req, "id = ? AND tenant_id = ?", id, tenantID).Error
	return &req, err
}

func (r *ChildConsentRepository) ListPendingRequests(parentID, tenantID uuid.UUID) ([]models.ParentalConsentRequest, error) {
	var reqs []models.ParentalConsentRequest
	err := r.db.Where("parent_id = ? AND tenant_id = ? AND status = ?", parentID, tenantID, "pending").Find(&reqs).Error
	return reqs, err
}

func (r *ChildConsentRepository) UpdateConsentRequest(req *models.ParentalConsentRequest) error {
	return r.db.Where("id = ? AND tenant_id = ?", req.ID, req.TenantID).Save(req).Error
}
