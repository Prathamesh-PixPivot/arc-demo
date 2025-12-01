package services

import (
	"pixpivot/arc/internal/models"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PurposeService handles purpose operations including compliance and hierarchy
type PurposeService struct {
	DB *gorm.DB
}

// NewPurposeService creates a new purpose service
func NewPurposeService(db *gorm.DB) *PurposeService {
	return &PurposeService{
		DB: db,
	}
}

// ValidateCompliance performs comprehensive compliance validation for a purpose
func (s *PurposeService) ValidateCompliance(purposeID uuid.UUID) (*models.ComplianceReport, error) {
	var purpose models.Purpose
	if err := s.DB.First(&purpose, "id = ?", purposeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("purpose not found")
		}
		return nil, fmt.Errorf("failed to get purpose: %w", err)
	}

	report := &models.ComplianceReport{
		PurposeID:       purposeID,
		Status:          "compliant",
		Issues:          []string{},
		Recommendations: []string{},
		LastChecked:     time.Now(),
	}

	// Validate required fields
	if purpose.Name == "" {
		report.Issues = append(report.Issues, "Purpose name is required")
		report.Status = "non_compliant"
	}

	if purpose.Description == "" {
		report.Issues = append(report.Issues, "Purpose description is required")
		report.Status = "non_compliant"
	}

	if purpose.LegalBasis == "" {
		report.Issues = append(report.Issues, "Legal basis is required under DPDP Act")
		report.Status = "non_compliant"
	}

	// Validate legal basis appropriateness
	if err := s.validateLegalBasisAppropriate(&purpose, report); err != nil {
		return nil, fmt.Errorf("failed to validate legal basis: %w", err)
	}

	// Check data retention period
	if purpose.RetentionPeriodDays <= 0 {
		report.Issues = append(report.Issues, "Data retention period must be specified and positive")
		if report.Status != "non_compliant" {
			report.Status = "needs_review"
		}
	} else if purpose.RetentionPeriodDays > 2555 { // ~7 years
		report.Recommendations = append(report.Recommendations, "Consider if such a long retention period is necessary and proportionate")
		if report.Status == "compliant" {
			report.Status = "needs_review"
		}
	}

	// Check for required data objects if purpose is linked to template
	if purpose.TemplateID != nil {
		if err := s.checkRequiredDataObjects(&purpose, report); err != nil {
			return nil, fmt.Errorf("failed to check required data objects: %w", err)
		}
	}

	// Validate hierarchy consistency
	if purpose.ParentPurposeID != nil {
		if err := s.validateHierarchyConsistency(&purpose, report); err != nil {
			return nil, fmt.Errorf("failed to validate hierarchy: %w", err)
		}
	}

	// Add general recommendations based on legal basis
	s.addRecommendationsByLegalBasis(&purpose, report)

	// Update purpose compliance status
	purpose.ComplianceStatus = report.Status
	now := time.Now()
	purpose.LastComplianceCheck = &now
	if err := s.DB.Save(&purpose).Error; err != nil {
		return nil, fmt.Errorf("failed to update purpose compliance status: %w", err)
	}

	return report, nil
}

// validateLegalBasisAppropriate checks if the legal basis is appropriate for the purpose type
func (s *PurposeService) validateLegalBasisAppropriate(purpose *models.Purpose, report *models.ComplianceReport) error {
	// Get template if linked to understand purpose category
	if purpose.TemplateID != nil {
		var template models.PurposeTemplate
		if err := s.DB.First(&template, "id = ?", *purpose.TemplateID).Error; err == nil {
			// Marketing purposes should use consent
			if template.Category == "marketing" && purpose.LegalBasis != "consent" {
				report.Issues = append(report.Issues, "Marketing purposes must use 'consent' as legal basis under DPDP Act")
				report.Status = "non_compliant"
			}
			
			// Necessary purposes should use contract or legal obligation
			if template.Category == "necessary" && 
				purpose.LegalBasis != "contract" && 
				purpose.LegalBasis != "legal_obligation" {
				report.Issues = append(report.Issues, "Necessary purposes should use 'contract' or 'legal_obligation' as legal basis")
				if report.Status != "non_compliant" {
					report.Status = "needs_review"
				}
			}
		}
	}

	// Validate legal basis values
	validLegalBases := []string{"consent", "contract", "legal_obligation", "legitimate_interest"}
	isValid := false
	for _, valid := range validLegalBases {
		if purpose.LegalBasis == valid {
			isValid = true
			break
		}
	}
	
	if !isValid {
		report.Issues = append(report.Issues, fmt.Sprintf("Invalid legal basis '%s'. Must be one of: %s", 
			purpose.LegalBasis, strings.Join(validLegalBases, ", ")))
		report.Status = "non_compliant"
	}

	return nil
}

// checkRequiredDataObjects validates that required data objects are properly configured
func (s *PurposeService) checkRequiredDataObjects(purpose *models.Purpose, report *models.ComplianceReport) error {
	var template models.PurposeTemplate
	if err := s.DB.First(&template, "id = ?", *purpose.TemplateID).Error; err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	if len(template.RequiredDataObjects) > 0 {
		// Check if purpose has consent forms that include required data objects
		var consentForms []models.ConsentFormPurpose
		if err := s.DB.Where("purpose_id = ?", purpose.ID).Find(&consentForms).Error; err != nil {
			return fmt.Errorf("failed to get consent forms: %w", err)
		}

		if len(consentForms) == 0 {
			report.Recommendations = append(report.Recommendations, 
				fmt.Sprintf("Purpose should be used in consent forms with required data objects: %s", 
					strings.Join(template.RequiredDataObjects, ", ")))
		}
	}

	return nil
}

// validateHierarchyConsistency checks for circular references and logical consistency
func (s *PurposeService) validateHierarchyConsistency(purpose *models.Purpose, report *models.ComplianceReport) error {
	// Check for circular references
	visited := make(map[uuid.UUID]bool)
	current := purpose.ParentPurposeID
	
	for current != nil {
		if visited[*current] {
			report.Issues = append(report.Issues, "Circular reference detected in purpose hierarchy")
			report.Status = "non_compliant"
			break
		}
		
		visited[*current] = true
		
		var parent models.Purpose
		if err := s.DB.First(&parent, "id = ?", *current).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				report.Issues = append(report.Issues, "Parent purpose not found")
				report.Status = "non_compliant"
				break
			}
			return fmt.Errorf("failed to get parent purpose: %w", err)
		}
		
		current = parent.ParentPurposeID
	}

	return nil
}

// addRecommendationsByLegalBasis adds recommendations based on legal basis
func (s *PurposeService) addRecommendationsByLegalBasis(purpose *models.Purpose, report *models.ComplianceReport) {
	switch purpose.LegalBasis {
	case "consent":
		report.Recommendations = append(report.Recommendations, 
			"Ensure explicit, informed consent is obtained before processing",
			"Provide clear withdrawal mechanisms",
			"Document consent records with timestamp and method")
	case "contract":
		report.Recommendations = append(report.Recommendations,
			"Ensure processing is necessary for contract performance",
			"Document the contractual necessity",
			"Limit processing to what's strictly necessary")
	case "legal_obligation":
		report.Recommendations = append(report.Recommendations,
			"Document the specific legal obligation",
			"Ensure processing is limited to what's required by law",
			"Maintain records of legal basis")
	case "legitimate_interest":
		report.Recommendations = append(report.Recommendations,
			"Conduct and document legitimate interest assessment",
			"Ensure interests are balanced against individual rights",
			"Provide opt-out mechanisms where appropriate")
	}
}

// SuggestRetentionPeriod suggests an appropriate retention period for a purpose
func (s *PurposeService) SuggestRetentionPeriod(purpose *models.Purpose) int {
	// If linked to template, use template suggestion
	if purpose.TemplateID != nil {
		var template models.PurposeTemplate
		if err := s.DB.First(&template, "id = ?", *purpose.TemplateID).Error; err == nil {
			return template.SuggestedRetentionDays
		}
	}

	// Default suggestions based on legal basis
	switch purpose.LegalBasis {
	case "consent":
		return 1095 // 3 years - typical for consent-based processing
	case "contract":
		return 2555 // 7 years - typical for contractual records
	case "legal_obligation":
		return 2555 // 7 years - often required by law
	case "legitimate_interest":
		return 730 // 2 years - balanced approach
	default:
		return 365 // 1 year - conservative default
	}
}

// ValidateLegalBasis validates if the legal basis is appropriate for the purpose
func (s *PurposeService) ValidateLegalBasis(purpose *models.Purpose) error {
	validBases := []string{"consent", "contract", "legal_obligation", "legitimate_interest"}
	
	for _, valid := range validBases {
		if purpose.LegalBasis == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid legal basis '%s'. Must be one of: %s", 
		purpose.LegalBasis, strings.Join(validBases, ", "))
}

// GetPurposeUsageStats returns usage statistics for a purpose
func (s *PurposeService) GetPurposeUsageStats(purposeID uuid.UUID) (*models.UsageStats, error) {
	var purpose models.Purpose
	if err := s.DB.First(&purpose, "id = ?", purposeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("purpose not found")
		}
		return nil, fmt.Errorf("failed to get purpose: %w", err)
	}

	stats := &models.UsageStats{
		PurposeID:    purposeID,
		TotalGranted: purpose.TotalGranted,
		TotalRevoked: purpose.TotalRevoked,
		LastUsed:     purpose.LastUsedAt,
	}

	// Count active consents
	var activeConsents int64
	if err := s.DB.Model(&models.UserConsent{}).
		Where("purpose_id = ? AND status = ? AND (expires_at IS NULL OR expires_at > ?)", 
			purposeID, true, time.Now()).
		Count(&activeConsents).Error; err != nil {
		return nil, fmt.Errorf("failed to count active consents: %w", err)
	}
	stats.ActiveConsents = int(activeConsents)

	// Count consent forms using this purpose
	var consentForms int64
	if err := s.DB.Model(&models.ConsentFormPurpose{}).
		Where("purpose_id = ?", purposeID).
		Count(&consentForms).Error; err != nil {
		return nil, fmt.Errorf("failed to count consent forms: %w", err)
	}
	stats.ConsentForms = int(consentForms)

	// Count expiring consents (next 30 days)
	var expiringConsents int64
	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)
	if err := s.DB.Model(&models.UserConsent{}).
		Where("purpose_id = ? AND status = ? AND expires_at BETWEEN ? AND ?", 
			purposeID, true, time.Now(), thirtyDaysFromNow).
		Count(&expiringConsents).Error; err != nil {
		return nil, fmt.Errorf("failed to count expiring consents: %w", err)
	}
	stats.ExpiringConsents = int(expiringConsents)

	return stats, nil
}

// GetExpiringConsents returns consents that will expire within the specified number of days
func (s *PurposeService) GetExpiringConsents(purposeID uuid.UUID, days int) ([]*models.UserConsent, error) {
	var consents []*models.UserConsent
	
	endDate := time.Now().AddDate(0, 0, days)
	
	if err := s.DB.Where("purpose_id = ? AND status = ? AND expires_at BETWEEN ? AND ?", 
		purposeID, true, time.Now(), endDate).
		Find(&consents).Error; err != nil {
		return nil, fmt.Errorf("failed to get expiring consents: %w", err)
	}
	
	return consents, nil
}

