package services

import (
	"bytes"
	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"

	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type SDKGeneratorService struct {
	repo             *repository.SDKRepository
	consentFormRepo  *repository.ConsentFormRepository
	baseURL          string
	templatePath     string
}

func NewSDKGeneratorService(repo *repository.SDKRepository, consentFormRepo *repository.ConsentFormRepository, baseURL string) *SDKGeneratorService {
	return &SDKGeneratorService{
		repo:            repo,
		consentFormRepo: consentFormRepo,
		baseURL:         baseURL,
		templatePath:    "templates/sdk",
	}
}

type SDKTemplateData struct {
	APIEndpoint            string
	TenantID               string
	FormID                 string
	Theme                  string
	Position               string
	Language               string
	AutoShow               bool
	ShowPreferenceCenter   bool
	CookieExpiry           int
	CustomCSS              string
}

func (s *SDKGeneratorService) GenerateSDK(tenantID, formID uuid.UUID, config *models.SDKConfig) (string, error) {
	// Check if SDK is cached
	if cachedSDK, err := s.GetCachedSDK(tenantID, formID); err == nil && cachedSDK != "" {
		return cachedSDK, nil
	}

	// Load template
	templatePath := filepath.Join(s.templatePath, "consent-sdk.js.tmpl")
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read SDK template: %w", err)
	}

	// Parse template
	tmpl, err := template.New("sdk").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse SDK template: %w", err)
	}

	// Prepare template data
	themeJSON, err := json.Marshal(config.Theme)
	if err != nil {
		return "", fmt.Errorf("failed to marshal theme: %w", err)
	}

	data := SDKTemplateData{
		APIEndpoint:          s.baseURL,
		TenantID:             tenantID.String(),
		FormID:               formID.String(),
		Theme:                string(themeJSON),
		Position:             config.Position,
		Language:             config.Language,
		AutoShow:             config.AutoShow,
		ShowPreferenceCenter: config.ShowPreferenceCenter,
		CookieExpiry:         config.CookieExpiry,
		CustomCSS:            config.CustomCSS,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute SDK template: %w", err)
	}

	// Minify JavaScript
	sdk := buf.String()
	minifiedSDK, err := s.MinifyJavaScript(sdk)
	if err != nil {
		// If minification fails, use original
		minifiedSDK = sdk
	}

	// Cache the generated SDK
	if err := s.CacheSDK(tenantID, formID, minifiedSDK); err != nil {
		// Log error but don't fail the generation
		fmt.Printf("Warning: Failed to cache SDK: %v\n", err)
	}

	return minifiedSDK, nil
}

func (s *SDKGeneratorService) MinifyJavaScript(code string) (string, error) {
	// Basic JavaScript minification
	// Remove comments
	commentRegex := regexp.MustCompile(`//.*$|/\*[\s\S]*?\*/`)
	code = commentRegex.ReplaceAllString(code, "")

	// Remove extra whitespace
	whitespaceRegex := regexp.MustCompile(`\s+`)
	code = whitespaceRegex.ReplaceAllString(code, " ")

	// Remove whitespace around operators and punctuation
	operatorRegex := regexp.MustCompile(`\s*([{}();,=+\-*/<>!&|])\s*`)
	code = operatorRegex.ReplaceAllString(code, "$1")

	// Trim
	code = strings.TrimSpace(code)

	return code, nil
}

func (s *SDKGeneratorService) GenerateCDNPath(tenantID, formID uuid.UUID) string {
	return fmt.Sprintf("/api/v1/public/sdk/%s/%s.js", tenantID.String(), formID.String())
}

func (s *SDKGeneratorService) CacheSDK(tenantID, formID uuid.UUID, sdk string) error {
	// For now, we'll implement a simple in-memory cache
	// In production, you'd want to use Redis or similar
	cacheKey := fmt.Sprintf("sdk:%s:%s", tenantID.String(), formID.String())
	
	// Store in a simple map with expiration (this is a simplified implementation)
	// In a real implementation, you'd use a proper cache like Redis
	return s.repo.CacheSDK(cacheKey, sdk, time.Hour)
}

func (s *SDKGeneratorService) GetCachedSDK(tenantID, formID uuid.UUID) (string, error) {
	cacheKey := fmt.Sprintf("sdk:%s:%s", tenantID.String(), formID.String())
	return s.repo.GetCachedSDK(cacheKey)
}

func (s *SDKGeneratorService) InvalidateCache(tenantID, formID uuid.UUID) error {
	cacheKey := fmt.Sprintf("sdk:%s:%s", tenantID.String(), formID.String())
	return s.repo.InvalidateCache(cacheKey)
}

func (s *SDKGeneratorService) CreateSDKConfig(tenantID uuid.UUID, req *dto.CreateSDKConfigRequest) (*models.SDKConfig, error) {
	// Validate consent form exists and belongs to tenant
	form, err := s.consentFormRepo.GetConsentFormByID(req.ConsentFormID)
	if err != nil {
		return nil, fmt.Errorf("consent form not found: %w", err)
	}
	
	if form.TenantID != tenantID {
		return nil, fmt.Errorf("consent form does not belong to tenant")
	}

	// Marshal theme to JSON
	themeJSON, err := json.Marshal(req.Theme)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal theme: %w", err)
	}

	config := &models.SDKConfig{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		ConsentFormID:        req.ConsentFormID,
		Theme:                datatypes.JSON(themeJSON),
		Position:             req.Position,
		Language:             req.Language,
		ShowPreferenceCenter: req.ShowPreferenceCenter,
		AutoShow:             req.AutoShow,
		CookieExpiry:         req.CookieExpiry,
		CustomCSS:            req.CustomCSS,
	}

	// Set defaults
	if config.Position == "" {
		config.Position = "bottom"
	}
	if config.Language == "" {
		config.Language = "en"
	}
	if config.CookieExpiry == 0 {
		config.CookieExpiry = 365
	}

	if err := s.repo.Create(config); err != nil {
		return nil, fmt.Errorf("failed to create SDK config: %w", err)
	}

	// Invalidate cache for this form
	s.InvalidateCache(tenantID, req.ConsentFormID)

	return config, nil
}

func (s *SDKGeneratorService) UpdateSDKConfig(configID uuid.UUID, tenantID uuid.UUID, req *dto.UpdateSDKConfigRequest) (*models.SDKConfig, error) {
	config, err := s.repo.GetByID(configID)
	if err != nil {
		return nil, fmt.Errorf("SDK config not found: %w", err)
	}

	if config.TenantID != tenantID {
		return nil, fmt.Errorf("SDK config does not belong to tenant")
	}

	// Marshal theme to JSON
	themeJSON, err := json.Marshal(req.Theme)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal theme: %w", err)
	}

	// Update fields
	config.Theme = datatypes.JSON(themeJSON)
	config.Position = req.Position
	config.Language = req.Language
	config.ShowPreferenceCenter = req.ShowPreferenceCenter
	config.AutoShow = req.AutoShow
	config.CookieExpiry = req.CookieExpiry
	config.CustomCSS = req.CustomCSS

	if err := s.repo.Update(config); err != nil {
		return nil, fmt.Errorf("failed to update SDK config: %w", err)
	}

	// Invalidate cache for this form
	s.InvalidateCache(tenantID, config.ConsentFormID)

	return config, nil
}

func (s *SDKGeneratorService) GetSDKConfigByFormID(tenantID, formID uuid.UUID) (*models.SDKConfig, error) {
	config, err := s.repo.GetByFormID(tenantID, formID)
	if err != nil {
		return nil, fmt.Errorf("SDK config not found: %w", err)
	}

	return config, nil
}

func (s *SDKGeneratorService) GetSDKConfig(configID uuid.UUID, tenantID uuid.UUID) (*models.SDKConfig, error) {
	config, err := s.repo.GetByID(configID)
	if err != nil {
		return nil, fmt.Errorf("SDK config not found: %w", err)
	}

	if config.TenantID != tenantID {
		return nil, fmt.Errorf("SDK config does not belong to tenant")
	}

	return config, nil
}

func (s *SDKGeneratorService) DeleteSDKConfig(configID uuid.UUID, tenantID uuid.UUID) error {
	config, err := s.repo.GetByID(configID)
	if err != nil {
		return fmt.Errorf("SDK config not found: %w", err)
	}

	if config.TenantID != tenantID {
		return fmt.Errorf("SDK config does not belong to tenant")
	}

	// Invalidate cache for this form
	s.InvalidateCache(tenantID, config.ConsentFormID)

	return s.repo.Delete(configID)
}

func (s *SDKGeneratorService) GenerateIntegrationCode(tenantID, formID uuid.UUID) (*dto.SDKGenerationResponse, error) {
	sdkURL := fmt.Sprintf("%s%s", s.baseURL, s.GenerateCDNPath(tenantID, formID))
	previewURL := fmt.Sprintf("%s/sdk-preview/%s/%s", s.baseURL, tenantID.String(), formID.String())
	
	integrationCode := fmt.Sprintf(`<script src="%s" async></script>`, sdkURL)

	return &dto.SDKGenerationResponse{
		SDKURL:          sdkURL,
		IntegrationCode: integrationCode,
		PreviewURL:      previewURL,
	}, nil
}

