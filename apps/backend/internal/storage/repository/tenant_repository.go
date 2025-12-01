package repository

import (
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(tenant *models.Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *TenantRepository) GetByID(id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.First(&tenant, "tenant_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

type TenantListParams struct {
	Page   int
	Limit  int
	Search string
}

type TenantListResponse struct {
	Tenants []*models.Tenant `json:"tenants"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}

func (r *TenantRepository) List(params TenantListParams) (*TenantListResponse, error) {
	var tenants []*models.Tenant
	var total int64
	query := r.db.Model(&models.Tenant{})

	if params.Search != "" {
		query = query.Where("name ILIKE ?", "%"+params.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	if err := query.Offset(offset).Limit(params.Limit).Order("created_at DESC").Find(&tenants).Error; err != nil {
		return nil, err
	}

	return &TenantListResponse{
		Tenants: tenants,
		Total:   total,
		Page:    params.Page,
		Limit:   params.Limit,
	}, nil
}

func (r *TenantRepository) Update(tenant *models.Tenant) error {
	return r.db.Save(tenant).Error
}

func (r *TenantRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Tenant{}, "tenant_id = ?", id).Error
}
