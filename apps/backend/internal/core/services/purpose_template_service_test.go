package services

import (
	"pixpivot/arc/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTemplateTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// Auto-migrate the schema
	db.AutoMigrate(&models.PurposeTemplate{}, &models.Purpose{})

	return db
}

func TestListTemplates_Success(t *testing.T) {
	db := setupTemplateTestDB()
	service := NewPurposeTemplateService(db)

	// Create test templates
	template1 := &models.PurposeTemplate{
		ID:                   uuid.New(),
		Name:                 "Email Marketing",
		Description:          "Send promotional emails",
		Category:             "marketing",
		LegalBasis:           "consent",
		RegulatoryFramework:  "dpdp",
		SuggestedRetentionDays: 1095,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	template2 := &models.PurposeTemplate{
		ID:                   uuid.New(),
		Name:                 "Website Analytics",
		Description:          "Analyze website usage",
		Category:             "analytics",
		LegalBasis:           "legitimate_interest",
		RegulatoryFramework:  "dpdp",
		SuggestedRetentionDays: 730,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := db.Create(template1).Error
	assert.NoError(t, err)
	err = db.Create(template2).Error
	assert.NoError(t, err)

	// Test listing all templates
	templates, err := service.ListTemplates("")
	assert.NoError(t, err)
	assert.Len(t, templates, 2)

	// Test filtering by framework
	templates, err = service.ListTemplates("dpdp")
	assert.NoError(t, err)
	assert.Len(t, templates, 2)

	// Test filtering by non-existent framework
	templates, err = service.ListTemplates("gdpr")
	assert.NoError(t, err)
	assert.Len(t, templates, 0)
}

func TestGetTemplate_Success(t *testing.T) {
	db := setupTemplateTestDB()
	service := NewPurposeTemplateService(db)

	// Create test template
	template := &models.PurposeTemplate{
		ID:                   uuid.New(),
		Name:                 "Transaction Processing",
		Description:          "Process payments",
		Category:             "necessary",
		LegalBasis:           "contract",
		RegulatoryFramework:  "dpdp",
		SuggestedRetentionDays: 2555,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := db.Create(template).Error
	assert.NoError(t, err)

	// Test getting existing template
	result, err := service.GetTemplate(template.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, template.Name, result.Name)
	assert.Equal(t, template.Category, result.Category)
}

func TestGetTemplate_NotFound(t *testing.T) {
	db := setupTemplateTestDB()
	service := NewPurposeTemplateService(db)

	// Test getting non-existent template
	_, err := service.GetTemplate(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "purpose template not found")
}

func TestCreatePurposeFromTemplate_Success(t *testing.T) {
	db := setupTemplateTestDB()
	service := NewPurposeTemplateService(db)

	// Create test template
	template := &models.PurposeTemplate{
		ID:                   uuid.New(),
		Name:                 "Customer Support",
		Description:          "Provide customer service",
		Category:             "functional",
		LegalBasis:           "contract",
		RegulatoryFramework:  "dpdp",
		SuggestedRetentionDays: 1095,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := db.Create(template).Error
	assert.NoError(t, err)

	tenantID := uuid.New()

	// Test creating purpose from template without customizations
	purpose, err := service.CreatePurposeFromTemplate(tenantID, template.ID, nil)
	assert.NoError(t, err)
	assert.NotNil(t, purpose)
	assert.Equal(t, template.Name, purpose.Name)
	assert.Equal(t, template.Description, purpose.Description)
	assert.Equal(t, template.LegalBasis, purpose.LegalBasis)
	assert.Equal(t, template.SuggestedRetentionDays, purpose.RetentionPeriodDays)
	assert.Equal(t, tenantID, purpose.TenantID)
	assert.Equal(t, template.ID, *purpose.TemplateID)
	assert.Equal(t, "compliant", purpose.ComplianceStatus)
	assert.True(t, purpose.Active)
}

func TestCreatePurposeFromTemplate_WithCustomizations(t *testing.T) {
	db := setupTemplateTestDB()
	service := NewPurposeTemplateService(db)

	// Create test template
	template := &models.PurposeTemplate{
		ID:                   uuid.New(),
		Name:                 "Default Name",
		Description:          "Default Description",
		Category:             "functional",
		LegalBasis:           "contract",
		RegulatoryFramework:  "dpdp",
		SuggestedRetentionDays: 365,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := db.Create(template).Error
	assert.NoError(t, err)

	tenantID := uuid.New()
	customizations := map[string]interface{}{
		"name":                   "Custom Name",
		"description":            "Custom Description",
		"required":               true,
		"review_cycle_months":    12,
		"retention_period_days":  730,
	}

	// Test creating purpose from template with customizations
	purpose, err := service.CreatePurposeFromTemplate(tenantID, template.ID, customizations)
	assert.NoError(t, err)
	assert.NotNil(t, purpose)
	assert.Equal(t, "Custom Name", purpose.Name)
	assert.Equal(t, "Custom Description", purpose.Description)
	assert.True(t, purpose.Required)
	assert.Equal(t, 12, purpose.ReviewCycleMonths)
	assert.Equal(t, 730, purpose.RetentionPeriodDays)
}

func TestListTemplatesByCategory_Success(t *testing.T) {
	db := setupTemplateTestDB()
	service := NewPurposeTemplateService(db)

	// Create templates in different categories
	marketingTemplate := &models.PurposeTemplate{
		ID:                   uuid.New(),
		Name:                 "Email Marketing",
		Category:             "marketing",
		LegalBasis:           "consent",
		RegulatoryFramework:  "dpdp",
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	analyticsTemplate := &models.PurposeTemplate{
		ID:                   uuid.New(),
		Name:                 "Website Analytics",
		Category:             "analytics",
		LegalBasis:           "legitimate_interest",
		RegulatoryFramework:  "dpdp",
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := db.Create(marketingTemplate).Error
	assert.NoError(t, err)
	err = db.Create(analyticsTemplate).Error
	assert.NoError(t, err)

	// Test filtering by marketing category
	templates, err := service.ListTemplatesByCategory("marketing")
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "Email Marketing", templates[0].Name)

	// Test filtering by analytics category
	templates, err = service.ListTemplatesByCategory("analytics")
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "Website Analytics", templates[0].Name)
}

func TestSearchTemplates_Success(t *testing.T) {
	db := setupTemplateTestDB()
	service := NewPurposeTemplateService(db)

	// Create test templates
	template1 := &models.PurposeTemplate{
		ID:                  uuid.New(),
		Name:                "Email Marketing Campaign",
		Description:         "Send promotional emails to customers",
		Category:            "marketing",
		LegalBasis:          "consent",
		RegulatoryFramework: "dpdp",
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	template2 := &models.PurposeTemplate{
		ID:                  uuid.New(),
		Name:                "Customer Support",
		Description:         "Provide customer service via email",
		Category:            "functional",
		LegalBasis:          "contract",
		RegulatoryFramework: "dpdp",
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	err := db.Create(template1).Error
	assert.NoError(t, err)
	err = db.Create(template2).Error
	assert.NoError(t, err)

	// Test searching by name
	templates, err := service.SearchTemplates("Email")
	assert.NoError(t, err)
	assert.Len(t, templates, 2) // Both contain "email" in name or description

	// Test searching by specific term
	templates, err = service.SearchTemplates("Marketing")
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "Email Marketing Campaign", templates[0].Name)

	// Test searching with no results
	templates, err = service.SearchTemplates("NonExistent")
	assert.NoError(t, err)
	assert.Len(t, templates, 0)
}

func TestValidateTemplateCompliance_Success(t *testing.T) {
	db := setupTemplateTestDB()
	service := NewPurposeTemplateService(db)

	// Create compliant template
	template := &models.PurposeTemplate{
		ID:                   uuid.New(),
		Name:                 "Valid Template",
		Description:          "Valid description",
		Category:             "marketing",
		LegalBasis:           "consent", // Correct for marketing
		RegulatoryFramework:  "dpdp",
		SuggestedRetentionDays: 365,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := db.Create(template).Error
	assert.NoError(t, err)

	// Validate compliance
	report, err := service.ValidateTemplateCompliance(template.ID)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "compliant", report.Status)
	assert.Empty(t, report.Issues)
	assert.NotEmpty(t, report.Recommendations) // Should have marketing-specific recommendations
}

func TestValidateTemplateCompliance_NonCompliant(t *testing.T) {
	db := setupTemplateTestDB()
	service := NewPurposeTemplateService(db)

	// Create non-compliant template
	template := &models.PurposeTemplate{
		ID:                   uuid.New(),
		Name:                 "", // Missing name
		Description:          "", // Missing description
		Category:             "marketing",
		LegalBasis:           "contract", // Wrong legal basis for marketing
		RegulatoryFramework:  "dpdp",
		SuggestedRetentionDays: -1, // Invalid retention period
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := db.Create(template).Error
	assert.NoError(t, err)

	// Validate compliance
	report, err := service.ValidateTemplateCompliance(template.ID)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "non_compliant", report.Status)
	assert.Contains(t, report.Issues, "Template name is required")
	assert.Contains(t, report.Issues, "Template description is required")
	assert.Contains(t, report.Issues, "Marketing purposes must use consent as legal basis under DPDP Act")
	assert.Contains(t, report.Issues, "Suggested retention period must be positive")
}

