package services

import (
	"crypto/rsa"
	"fmt"
	"time"

	"pixpivot/arc/internal/licensing"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type SuperAdminService struct {
	licenseRepo *repository.IssuedLicenseRepository
	tenantRepo  *repository.TenantRepository
	userRepo    *repository.FiduciaryRepository
	privateKey  *rsa.PrivateKey
}

func NewSuperAdminService(
	licenseRepo *repository.IssuedLicenseRepository,
	tenantRepo *repository.TenantRepository,
	userRepo *repository.FiduciaryRepository,
	privateKey *rsa.PrivateKey,
) *SuperAdminService {
	return &SuperAdminService{
		licenseRepo: licenseRepo,
		tenantRepo:  tenantRepo,
		userRepo:    userRepo,
		privateKey:  privateKey,
	}
}

// ... GenerateLicense ...

// Tenant Management

func (s *SuperAdminService) ListTenants(params repository.TenantListParams) (*repository.TenantListResponse, error) {
	return s.tenantRepo.List(params)
}

func (s *SuperAdminService) CreateTenant(tenant *models.Tenant) error {
	if tenant.TenantID == uuid.Nil {
		tenant.TenantID = uuid.New()
	}
	return s.tenantRepo.Create(tenant)
}

func (s *SuperAdminService) UpdateTenant(tenant *models.Tenant) error {
	return s.tenantRepo.Update(tenant)
}

func (s *SuperAdminService) DeleteTenant(id uuid.UUID) error {
	return s.tenantRepo.Delete(id)
}

type GenerateLicenseRequest struct {
	CustomerName string                  `json:"customerName"`
	PlanTier     string                  `json:"planTier"` // ENTERPRISE, PRO
	Type         string                  `json:"type"`     // SAAS, ON_PREM
	ExpiresAt    *time.Time              `json:"expiresAt"`
	Features     []string                `json:"features"`
	Limits       licensing.LicenseLimits `json:"limits"`
}

func (s *SuperAdminService) GenerateLicense(req GenerateLicenseRequest) (*models.IssuedLicense, error) {
	now := time.Now()
	customerID := uuid.New()
	licenseID := uuid.New()

	// 1. Create License Payload
	license := licensing.License{
		LicenseID:    licenseID,
		CustomerID:   customerID,
		CustomerName: req.CustomerName,
		Issuer:       "PixPivot",
		IssuedAt:     now,
		ExpiresAt:    req.ExpiresAt,
		Type:         licensing.LicenseType(req.Type),
		PlanTier:     licensing.PlanTier(req.PlanTier),
		Features:     req.Features,
		Limits:       req.Limits,
	}

	// 2. Sign License
	claims := licensing.LicenseClaims{
		License: license,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   "PixPivot",
			Subject:  customerID.String(),
			Audience: []string{"arc-backend"},
			IssuedAt: jwt.NewNumericDate(now),
			ID:       licenseID.String(),
		},
	}
	if req.ExpiresAt != nil {
		claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(*req.ExpiresAt)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedString, err := token.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign license: %w", err)
	}

	// 3. Store in DB
	issuedLicense := &models.IssuedLicense{
		ID:           licenseID,
		CustomerID:   customerID,
		CustomerName: req.CustomerName,
		PlanTier:     req.PlanTier,
		Type:         req.Type,
		IssuedAt:     now,
		ExpiresAt:    req.ExpiresAt,
		LicenseKey:   signedString,
		IsActive:     true,
	}

	if err := s.licenseRepo.Create(issuedLicense); err != nil {
		return nil, err
	}

	return issuedLicense, nil
}

func (s *SuperAdminService) ListLicenses(params repository.LicenseListParams) (*repository.LicenseListResponse, error) {
	return s.licenseRepo.List(params)
}

func (s *SuperAdminService) RevokeLicense(id uuid.UUID) error {
	return s.licenseRepo.Revoke(id)
}

// Tenant Management (Basic wrapper around existing repos or direct DB if needed)
// For now, we assume FiduciaryRepository handles users, but we might need a TenantRepository for the Tenant table.
// models.Tenant exists but I don't see a dedicated repository for it in the imports I've seen.
// I'll check if TenantRepository exists.
