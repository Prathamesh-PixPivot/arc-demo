package repository

import (
	"fmt"
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CookieRepository struct {
	db *gorm.DB
}

func NewCookieRepository(db *gorm.DB) *CookieRepository {
	return &CookieRepository{db: db}
}

// Cookie CRUD operations
func (r *CookieRepository) Create(cookie *models.Cookie) error {
	return r.db.Create(cookie).Error
}

func (r *CookieRepository) GetByID(id, tenantID uuid.UUID) (*models.Cookie, error) {
	var cookie models.Cookie
	err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&cookie).Error
	if err != nil {
		return nil, err
	}
	return &cookie, nil
}

func (r *CookieRepository) Update(cookie *models.Cookie) error {
	return r.db.Where("id = ? AND tenant_id = ?", cookie.ID, cookie.TenantID).Save(cookie).Error
}

func (r *CookieRepository) Delete(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Cookie{}).Error
}

func (r *CookieRepository) ListByTenant(tenantID uuid.UUID, category string, isActive *bool) ([]*models.Cookie, error) {
	var cookies []*models.Cookie
	query := r.db.Where("tenant_id = ?", tenantID)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	err := query.Order("created_at DESC").Find(&cookies).Error
	return cookies, err
}

func (r *CookieRepository) FindByNameAndDomain(tenantID uuid.UUID, name, domain string) (*models.Cookie, error) {
	var cookie models.Cookie
	err := r.db.Where("tenant_id = ? AND name = ? AND domain = ?", tenantID, name, domain).First(&cookie).Error
	if err != nil {
		return nil, err
	}
	return &cookie, nil
}

func (r *CookieRepository) BulkUpdateCategory(tenantID uuid.UUID, cookieIDs []uuid.UUID, category string) error {
	return r.db.Model(&models.Cookie{}).
		Where("tenant_id = ? AND id IN ?", tenantID, cookieIDs).
		Update("category", category).Error
}

func (r *CookieRepository) GetCookieStats(tenantID uuid.UUID) (map[string]int, error) {
	var results []struct {
		Category string
		Count    int
	}

	err := r.db.Model(&models.Cookie{}).
		Select("category, COUNT(*) as count").
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Group("category").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	for _, result := range results {
		stats[result.Category] = result.Count
	}

	return stats, nil
}

// Cookie Scan operations
func (r *CookieRepository) CreateScan(scan *models.CookieScan) error {
	return r.db.Create(scan).Error
}

func (r *CookieRepository) GetScanByID(id, tenantID uuid.UUID) (*models.CookieScan, error) {
	var scan models.CookieScan
	err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&scan).Error
	if err != nil {
		return nil, err
	}
	return &scan, nil
}

func (r *CookieRepository) UpdateScan(scan *models.CookieScan) error {
	return r.db.Where("id = ? AND tenant_id = ?", scan.ID, scan.TenantID).Save(scan).Error
}

func (r *CookieRepository) ListScansByTenant(tenantID uuid.UUID, limit int) ([]*models.CookieScan, error) {
	var scans []*models.CookieScan
	query := r.db.Where("tenant_id = ?", tenantID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&scans).Error
	return scans, err
}

func (r *CookieRepository) DeleteScan(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.CookieScan{}).Error
}

// Cookie Consent operations
func (r *CookieRepository) CreateCookieConsent(consent *models.CookieConsent) error {
	return r.db.Create(consent).Error
}

func (r *CookieRepository) GetCookieConsentByUserAndCookie(userConsentID, cookieID uuid.UUID) (*models.CookieConsent, error) {
	var consent models.CookieConsent
	err := r.db.Where("user_consent_id = ? AND cookie_id = ?", userConsentID, cookieID).First(&consent).Error
	if err != nil {
		return nil, err
	}
	return &consent, nil
}

func (r *CookieRepository) UpdateCookieConsent(consent *models.CookieConsent) error {
	return r.db.Save(consent).Error
}

func (r *CookieRepository) ListCookieConsentsByUser(userConsentID uuid.UUID) ([]*models.CookieConsent, error) {
	var consents []*models.CookieConsent
	err := r.db.Where("user_consent_id = ?", userConsentID).Find(&consents).Error
	return consents, err
}

func (r *CookieRepository) GetAllowedCookiesForTenant(tenantID uuid.UUID) ([]*models.Cookie, error) {
	var cookies []*models.Cookie
	err := r.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&cookies).Error
	return cookies, err
}

// Search and filter operations
func (r *CookieRepository) SearchCookies(tenantID uuid.UUID, searchTerm string, category string, provider string, limit int, offset int) ([]*models.Cookie, int64, error) {
	var cookies []*models.Cookie
	var total int64

	query := r.db.Model(&models.Cookie{}).Where("tenant_id = ?", tenantID)

	if searchTerm != "" {
		searchPattern := fmt.Sprintf("%%%s%%", searchTerm)
		query = query.Where("name ILIKE ? OR domain ILIKE ? OR purpose ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if provider != "" {
		query = query.Where("provider ILIKE ?", fmt.Sprintf("%%%s%%", provider))
	}

	// Get total count
	countQuery := query
	err := countQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&cookies).Error
	return cookies, total, err
}

func (r *CookieRepository) GetCookiesByCategory(tenantID uuid.UUID, category string) ([]*models.Cookie, error) {
	var cookies []*models.Cookie
	err := r.db.Where("tenant_id = ? AND category = ? AND is_active = ?", tenantID, category, true).Find(&cookies).Error
	return cookies, err
}

func (r *CookieRepository) BulkCreateCookies(cookies []*models.Cookie) error {
	return r.db.CreateInBatches(cookies, 100).Error
}

func (r *CookieRepository) GetRecentScans(tenantID uuid.UUID, limit int) ([]*models.CookieScan, error) {
	var scans []*models.CookieScan
	err := r.db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&scans).Error
	return scans, err
}
