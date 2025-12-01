package repository

import (
	"pixpivot/arc/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IssuedLicenseRepository struct {
	db *gorm.DB
}

func NewIssuedLicenseRepository(db *gorm.DB) *IssuedLicenseRepository {
	return &IssuedLicenseRepository{db: db}
}

func (r *IssuedLicenseRepository) Create(license *models.IssuedLicense) error {
	return r.db.Create(license).Error
}

func (r *IssuedLicenseRepository) GetByID(id uuid.UUID) (*models.IssuedLicense, error) {
	var license models.IssuedLicense
	if err := r.db.First(&license, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &license, nil
}

type LicenseListParams struct {
	Page     int
	Limit    int
	Search   string
	Type     string
	IsActive *bool
}

type LicenseListResponse struct {
	Licenses []*models.IssuedLicense `json:"licenses"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	Limit    int                     `json:"limit"`
}

func (r *IssuedLicenseRepository) List(params LicenseListParams) (*LicenseListResponse, error) {
	var licenses []*models.IssuedLicense
	var total int64
	query := r.db.Model(&models.IssuedLicense{})

	if params.Search != "" {
		query = query.Where("customer_name ILIKE ?", "%"+params.Search+"%")
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	if err := query.Offset(offset).Limit(params.Limit).Order("issued_at DESC").Find(&licenses).Error; err != nil {
		return nil, err
	}

	return &LicenseListResponse{
		Licenses: licenses,
		Total:    total,
		Page:     params.Page,
		Limit:    params.Limit,
	}, nil
}

func (r *IssuedLicenseRepository) Update(license *models.IssuedLicense) error {
	return r.db.Save(license).Error
}

func (r *IssuedLicenseRepository) Revoke(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.IssuedLicense{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_active":  false,
		"revoked_at": &now,
	}).Error
}
