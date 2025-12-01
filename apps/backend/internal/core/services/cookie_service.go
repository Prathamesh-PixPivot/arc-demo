package services

import (
	"fmt"
	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"

	"github.com/google/uuid"
)

type CookieService struct {
	repo *repository.CookieRepository
}

func NewCookieService(repo *repository.CookieRepository) *CookieService {
	return &CookieService{repo: repo}
}

// Cookie CRUD operations
func (s *CookieService) CreateCookie(tenantID uuid.UUID, req *dto.CreateCookieRequest) (*models.Cookie, error) {
	// Validate category
	if !s.isValidCategory(req.Category) {
		return nil, fmt.Errorf("invalid category: %s", req.Category)
	}

	// Check if cookie already exists
	existing, err := s.repo.FindByNameAndDomain(tenantID, req.Name, req.Domain)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("cookie with name '%s' and domain '%s' already exists", req.Name, req.Domain)
	}

	cookie := &models.Cookie{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Name:          req.Name,
		Domain:        req.Domain,
		Path:          req.Path,
		Category:      req.Category,
		Purpose:       req.Purpose,
		Provider:      req.Provider,
		ExpiryDays:    req.ExpiryDays,
		IsFirstParty:  req.IsFirstParty,
		IsSecure:      req.IsSecure,
		IsHttpOnly:    req.IsHttpOnly,
		SameSite:      req.SameSite,
		Description:   req.Description,
		DataCollected: req.DataCollected,
		IsActive:      true,
	}

	// Set default path if not provided
	if cookie.Path == "" {
		cookie.Path = "/"
	}

	if err := s.repo.Create(cookie); err != nil {
		return nil, fmt.Errorf("failed to create cookie: %w", err)
	}

	return cookie, nil
}

func (s *CookieService) GetCookie(id, tenantID uuid.UUID) (*models.Cookie, error) {
	return s.repo.GetByID(id, tenantID)
}

func (s *CookieService) UpdateCookie(id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateCookieRequest) (*models.Cookie, error) {
	cookie, err := s.repo.GetByID(id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("cookie not found: %w", err)
	}

	// Validate category if provided
	if req.Category != "" && !s.isValidCategory(req.Category) {
		return nil, fmt.Errorf("invalid category: %s", req.Category)
	}

	// Update fields
	if req.Name != "" {
		cookie.Name = req.Name
	}
	if req.Domain != "" {
		cookie.Domain = req.Domain
	}
	if req.Path != "" {
		cookie.Path = req.Path
	}
	if req.Category != "" {
		cookie.Category = req.Category
	}
	if req.Purpose != "" {
		cookie.Purpose = req.Purpose
	}
	if req.Provider != "" {
		cookie.Provider = req.Provider
	}
	if req.ExpiryDays != 0 {
		cookie.ExpiryDays = req.ExpiryDays
	}
	cookie.IsFirstParty = req.IsFirstParty
	cookie.IsSecure = req.IsSecure
	cookie.IsHttpOnly = req.IsHttpOnly
	if req.SameSite != "" {
		cookie.SameSite = req.SameSite
	}
	if req.Description != "" {
		cookie.Description = req.Description
	}
	if req.DataCollected != "" {
		cookie.DataCollected = req.DataCollected
	}
	cookie.IsActive = req.IsActive

	if err := s.repo.Update(cookie); err != nil {
		return nil, fmt.Errorf("failed to update cookie: %w", err)
	}

	return cookie, nil
}

func (s *CookieService) DeleteCookie(id uuid.UUID, tenantID uuid.UUID) error {
	// Verify existence and ownership
	_, err := s.repo.GetByID(id, tenantID)
	if err != nil {
		return fmt.Errorf("cookie not found: %w", err)
	}

	return s.repo.Delete(id, tenantID)
}

func (s *CookieService) ListCookies(tenantID uuid.UUID, category string, isActive *bool) ([]*models.Cookie, error) {
	return s.repo.ListByTenant(tenantID, category, isActive)
}

func (s *CookieService) SearchCookies(tenantID uuid.UUID, searchTerm string, category string, provider string, limit int, offset int) ([]*models.Cookie, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.SearchCookies(tenantID, searchTerm, category, provider, limit, offset)
}

func (s *CookieService) BulkUpdateCategories(tenantID uuid.UUID, req *dto.BulkCategorizeCookiesRequest) error {
	// Validate category
	if !s.isValidCategory(req.Category) {
		return fmt.Errorf("invalid category: %s", req.Category)
	}

	if len(req.CookieIDs) == 0 {
		return fmt.Errorf("no cookie IDs provided")
	}

	return s.repo.BulkUpdateCategory(tenantID, req.CookieIDs, req.Category)
}

func (s *CookieService) GetCookieStats(tenantID uuid.UUID) (map[string]int, error) {
	stats, err := s.repo.GetCookieStats(tenantID)
	if err != nil {
		return nil, err
	}

	// Ensure all categories are present
	allCategories := []string{
		models.CookieCategoryNecessary,
		models.CookieCategoryFunctional,
		models.CookieCategoryAnalytics,
		models.CookieCategoryMarketing,
	}

	for _, category := range allCategories {
		if _, exists := stats[category]; !exists {
			stats[category] = 0
		}
	}

	return stats, nil
}

func (s *CookieService) GetCookiesByCategory(tenantID uuid.UUID, category string) ([]*models.Cookie, error) {
	if !s.isValidCategory(category) {
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	return s.repo.GetCookiesByCategory(tenantID, category)
}

func (s *CookieService) GetAllowedCookiesForTenant(tenantID uuid.UUID) ([]*models.Cookie, error) {
	return s.repo.GetAllowedCookiesForTenant(tenantID)
}

// Cookie consent management
func (s *CookieService) SetCookieConsent(userConsentID, cookieID uuid.UUID, allowed bool) error {
	// Check if consent already exists
	existing, err := s.repo.GetCookieConsentByUserAndCookie(userConsentID, cookieID)
	if err != nil {
		// Create new consent
		consent := &models.CookieConsent{
			ID:            uuid.New(),
			UserConsentID: userConsentID,
			CookieID:      cookieID,
			Allowed:       allowed,
		}
		return s.repo.CreateCookieConsent(consent)
	}

	// Update existing consent
	existing.Allowed = allowed
	return s.repo.UpdateCookieConsent(existing)
}

func (s *CookieService) GetUserCookieConsents(userConsentID uuid.UUID) ([]*models.CookieConsent, error) {
	return s.repo.ListCookieConsentsByUser(userConsentID)
}

func (s *CookieService) IsCookieAllowed(userConsentID, cookieID uuid.UUID) (bool, error) {
	consent, err := s.repo.GetCookieConsentByUserAndCookie(userConsentID, cookieID)
	if err != nil {
		// No consent found, default to not allowed
		return false, nil
	}
	return consent.Allowed, nil
}

// Utility functions
func (s *CookieService) isValidCategory(category string) bool {
	validCategories := []string{
		models.CookieCategoryNecessary,
		models.CookieCategoryFunctional,
		models.CookieCategoryAnalytics,
		models.CookieCategoryMarketing,
	}

	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

func (s *CookieService) GetValidCategories() []string {
	return []string{
		models.CookieCategoryNecessary,
		models.CookieCategoryFunctional,
		models.CookieCategoryAnalytics,
		models.CookieCategoryMarketing,
	}
}

func (s *CookieService) GetCookiesByDomain(tenantID uuid.UUID, domain string) ([]*models.Cookie, error) {
	cookies, err := s.repo.ListByTenant(tenantID, "", nil)
	if err != nil {
		return nil, err
	}

	var filteredCookies []*models.Cookie
	for _, cookie := range cookies {
		if cookie.Domain == domain {
			filteredCookies = append(filteredCookies, cookie)
		}
	}

	return filteredCookies, nil
}

func (s *CookieService) GetRecentlyAddedCookies(tenantID uuid.UUID, limit int) ([]*models.Cookie, error) {
	if limit <= 0 {
		limit = 10
	}

	cookies, err := s.repo.ListByTenant(tenantID, "", nil)
	if err != nil {
		return nil, err
	}

	// Return the first 'limit' cookies (they're ordered by created_at DESC)
	if len(cookies) > limit {
		cookies = cookies[:limit]
	}

	return cookies, nil
}

func (s *CookieService) ValidateCookieData(req *dto.CreateCookieRequest) []string {
	var errors []string

	if req.Name == "" {
		errors = append(errors, "cookie name is required")
	}

	if req.Domain == "" {
		errors = append(errors, "cookie domain is required")
	}

	if req.Category == "" {
		errors = append(errors, "cookie category is required")
	} else if !s.isValidCategory(req.Category) {
		errors = append(errors, fmt.Sprintf("invalid category: %s", req.Category))
	}

	if req.ExpiryDays < 0 {
		errors = append(errors, "expiry days cannot be negative")
	}

	if req.SameSite != "" {
		validSameSite := []string{"None", "Lax", "Strict"}
		valid := false
		for _, v := range validSameSite {
			if req.SameSite == v {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, "invalid SameSite value, must be None, Lax, or Strict")
		}
	}

	return errors
}

// GetPublicCookieSettings returns cookie settings for public use
func (s *CookieService) GetPublicCookieSettings(tenantID uuid.UUID) (interface{}, error) {
	// TODO: Implement proper public cookie settings
	return map[string]interface{}{
		"tenant_id":       tenantID,
		"cookies_enabled": true,
		"categories":      []string{"necessary", "functional", "analytics", "marketing"},
	}, nil
}

// SubmitPublicCookieConsent handles public cookie consent submission
func (s *CookieService) SubmitPublicCookieConsent(tenantID uuid.UUID, consent interface{}) error {
	// TODO: Implement proper public cookie consent handling
	return nil
}
