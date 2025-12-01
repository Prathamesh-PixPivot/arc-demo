package services

import (
	"pixpivot/arc/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PurposeTemplateService handles purpose template operations
type PurposeTemplateService struct {
	DB *gorm.DB
}

// NewPurposeTemplateService creates a new purpose template service
func NewPurposeTemplateService(db *gorm.DB) *PurposeTemplateService {
	return &PurposeTemplateService{
		DB: db,
	}
}

// ListTemplates returns purpose templates filtered by regulatory framework
func (s *PurposeTemplateService) ListTemplates(framework string) ([]*models.PurposeTemplate, error) {
	var templates []*models.PurposeTemplate
	
	query := s.DB.Where("is_active = ?", true)
	
	if framework != "" {
		query = query.Where("regulatory_framework = ?", framework)
	}
	
	if err := query.Order("category, name").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to list purpose templates: %w", err)
	}
	
	return templates, nil
}

// GetTemplate returns a specific purpose template by ID
func (s *PurposeTemplateService) GetTemplate(id uuid.UUID) (*models.PurposeTemplate, error) {
	var template models.PurposeTemplate
	
	if err := s.DB.Where("id = ? AND is_active = ?", id, true).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("purpose template not found")
		}
		return nil, fmt.Errorf("failed to get purpose template: %w", err)
	}
	
	return &template, nil
}

// ListTemplatesByCategory returns templates filtered by category
func (s *PurposeTemplateService) ListTemplatesByCategory(category string) ([]*models.PurposeTemplate, error) {
	var templates []*models.PurposeTemplate
	
	query := s.DB.Where("is_active = ?", true)
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	if err := query.Order("name").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to list templates by category: %w", err)
	}
	
	return templates, nil
}

// CreatePurposeFromTemplate creates a new purpose based on a template
func (s *PurposeTemplateService) CreatePurposeFromTemplate(tenantID, templateID uuid.UUID, customizations map[string]interface{}) (*models.Purpose, error) {
	// Get the template
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	
	// Create purpose from template
	purpose := &models.Purpose{
		ID:                  uuid.New(),
		Name:                template.Name,
		Description:         template.Description,
		LegalBasis:          template.LegalBasis,
		TenantID:            tenantID,
		TemplateID:          &templateID,
		RetentionPeriodDays: template.SuggestedRetentionDays,
		ComplianceStatus:    "compliant", // Templates are pre-validated
		Active:              true,
		Required:            false, // Default, can be overridden
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	
	// Apply customizations if provided
	if customizations != nil {
		if name, ok := customizations["name"].(string); ok && name != "" {
			purpose.Name = name
		}
		if description, ok := customizations["description"].(string); ok && description != "" {
			purpose.Description = description
		}
		if required, ok := customizations["required"].(bool); ok {
			purpose.Required = required
		}
		if reviewCycle, ok := customizations["review_cycle_months"].(int); ok && reviewCycle > 0 {
			purpose.ReviewCycleMonths = reviewCycle
		}
		if retention, ok := customizations["retention_period_days"].(int); ok && retention > 0 {
			purpose.RetentionPeriodDays = retention
		}
	}
	
	// Save the purpose
	if err := s.DB.Create(purpose).Error; err != nil {
		return nil, fmt.Errorf("failed to create purpose from template: %w", err)
	}
	
	return purpose, nil
}

// GetTemplatesByLegalBasis returns templates filtered by legal basis
func (s *PurposeTemplateService) GetTemplatesByLegalBasis(legalBasis string) ([]*models.PurposeTemplate, error) {
	var templates []*models.PurposeTemplate
	
	query := s.DB.Where("is_active = ?", true)
	
	if legalBasis != "" {
		query = query.Where("legal_basis = ?", legalBasis)
	}
	
	if err := query.Order("category, name").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to list templates by legal basis: %w", err)
	}
	
	return templates, nil
}

// SearchTemplates searches templates by name or description
func (s *PurposeTemplateService) SearchTemplates(query string) ([]*models.PurposeTemplate, error) {
	var templates []*models.PurposeTemplate
	
	searchQuery := "%" + query + "%"
	
	if err := s.DB.Where("is_active = ? AND (name ILIKE ? OR description ILIKE ?)", 
		true, searchQuery, searchQuery).
		Order("category, name").
		Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to search templates: %w", err)
	}
	
	return templates, nil
}

// GetTemplateStats returns statistics about purpose templates
func (s *PurposeTemplateService) GetTemplateStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total templates
	var totalCount int64
	if err := s.DB.Model(&models.PurposeTemplate{}).Where("is_active = ?", true).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count templates: %w", err)
	}
	stats["total_templates"] = totalCount
	
	// Templates by category
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	if err := s.DB.Model(&models.PurposeTemplate{}).
		Select("category, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("category").
		Find(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	stats["by_category"] = categoryStats
	
	// Templates by legal basis
	var legalBasisStats []struct {
		LegalBasis string `json:"legal_basis"`
		Count      int64  `json:"count"`
	}
	if err := s.DB.Model(&models.PurposeTemplate{}).
		Select("legal_basis, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("legal_basis").
		Find(&legalBasisStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get legal basis stats: %w", err)
	}
	stats["by_legal_basis"] = legalBasisStats
	
	// Templates by regulatory framework
	var frameworkStats []struct {
		Framework string `json:"framework"`
		Count     int64  `json:"count"`
	}
	if err := s.DB.Model(&models.PurposeTemplate{}).
		Select("regulatory_framework as framework, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("regulatory_framework").
		Find(&frameworkStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get framework stats: %w", err)
	}
	stats["by_framework"] = frameworkStats
	
	return stats, nil
}

// ValidateTemplateCompliance validates if a template meets compliance requirements
func (s *PurposeTemplateService) ValidateTemplateCompliance(templateID uuid.UUID) (*models.ComplianceReport, error) {
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	
	report := &models.ComplianceReport{
		PurposeID:       templateID,
		Status:          "compliant",
		Issues:          []string{},
		Recommendations: []string{},
		LastChecked:     time.Now(),
	}
	
	// Validate required fields
	if template.Name == "" {
		report.Issues = append(report.Issues, "Template name is required")
		report.Status = "non_compliant"
	}
	
	if template.Description == "" {
		report.Issues = append(report.Issues, "Template description is required")
		report.Status = "non_compliant"
	}
	
	if template.LegalBasis == "" {
		report.Issues = append(report.Issues, "Legal basis is required")
		report.Status = "non_compliant"
	}
	
	// Validate legal basis for category
	if template.Category == "marketing" && template.LegalBasis != "consent" {
		report.Issues = append(report.Issues, "Marketing purposes must use consent as legal basis under DPDP Act")
		report.Status = "non_compliant"
	}
	
	// Validate retention period
	if template.SuggestedRetentionDays <= 0 {
		report.Issues = append(report.Issues, "Suggested retention period must be positive")
		if report.Status != "non_compliant" {
			report.Status = "needs_review"
		}
	}
	
	// Add recommendations based on category
	switch template.Category {
	case "marketing":
		report.Recommendations = append(report.Recommendations, "Ensure explicit consent is obtained before using this purpose")
		report.Recommendations = append(report.Recommendations, "Provide clear opt-out mechanisms")
	case "analytics":
		report.Recommendations = append(report.Recommendations, "Consider anonymizing data where possible")
		report.Recommendations = append(report.Recommendations, "Implement data minimization principles")
	case "necessary":
		report.Recommendations = append(report.Recommendations, "Document the necessity for contract performance or legal obligation")
	}
	
	return report, nil
}

