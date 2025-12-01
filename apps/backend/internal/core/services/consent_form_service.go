package services

import (
	"pixpivot/arc/config"
	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// ConsentFormRepositoryInterface defines the interface for consent form repository operations
type ConsentFormRepositoryInterface interface {
	GetConsentFormByID(formID uuid.UUID) (*models.ConsentForm, error)
	CreateConsentForm(form *models.ConsentForm) (*models.ConsentForm, error)
	UpdateConsentForm(form *models.ConsentForm) (*models.ConsentForm, error)
	DeleteConsentForm(formID uuid.UUID) error
	ListConsentForms(tenantID uuid.UUID) ([]models.ConsentForm, error)
	AddPurposeToConsentForm(formID, purposeID uuid.UUID, dataObjects, vendorIDs []string, expiryInDays int) (*models.ConsentFormPurpose, error)
	UpdatePurposeInConsentForm(formID, purposeID uuid.UUID, dataObjects, vendorIDs []string, expiryInDays int) (*models.ConsentFormPurpose, error)
	RemovePurposeFromConsentForm(formID, purposeID uuid.UUID) error
	GetConsentFormPurpose(formID, purposeID uuid.UUID) (*models.ConsentFormPurpose, error)
	PublishConsentForm(formID uuid.UUID) error
	PublishConsentFormWithVersioning(formID uuid.UUID, publishedBy uuid.UUID, changeLog string) error
	UpdateConsentFormStatus(formID uuid.UUID, status string) error
	GetVersionHistory(formID uuid.UUID) ([]*models.ConsentFormVersion, error)
	GetVersion(versionID uuid.UUID) (*models.ConsentFormVersion, error)
	RollbackToVersion(formID, versionID uuid.UUID, rolledBackBy uuid.UUID, reason string) error
	CreateVersion(form *models.ConsentForm, publishedBy uuid.UUID, changeLog string) (*models.ConsentFormVersion, error)
}

type ConsentFormService struct {
	repo               ConsentFormRepositoryInterface
	translationService *TranslationService
}

func NewConsentFormService(repo *repository.ConsentFormRepository) *ConsentFormService {
	return &ConsentFormService{
		repo:               repo,
		translationService: NewTranslationService(),
	}
}

func (s *ConsentFormService) CreateConsentForm(tenantID uuid.UUID, req *dto.CreateConsentFormRequest) (*models.ConsentForm, error) {
	var orgID uuid.UUID
	var err error
	if req.OrganizationEntityID != nil && *req.OrganizationEntityID != "" {
		orgID, err = uuid.Parse(*req.OrganizationEntityID)
		if err != nil {
			return nil, fmt.Errorf("invalid organizationEntityId: %w", err)
		}
	}

	form := &models.ConsentForm{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     req.Name,
		Title:    req.Title,
	}

	if req.Description != nil {
		form.Description = *req.Description
	}
	if req.Department != nil {
		form.Department = *req.Department
	}
	if req.Project != nil {
		form.Project = *req.Project
	}
	form.OrganizationEntityID = orgID
	if req.DataRetentionPeriod != nil {
		form.DataRetentionPeriod = *req.DataRetentionPeriod
	}
	if req.UserRightsSummary != nil {
		form.UserRightsSummary = *req.UserRightsSummary
	}
	if req.TermsAndConditions != nil {
		form.TermsAndConditions = *req.TermsAndConditions
	}
	if req.PrivacyPolicy != nil {
		form.PrivacyPolicy = *req.PrivacyPolicy
	}
	if req.Translations != nil {
		translationsJSON, err := json.Marshal(req.Translations)
		if err != nil {
			return nil, fmt.Errorf("invalid translations: %w", err)
		}
		form.Translations = datatypes.JSON(translationsJSON)
	}
	if req.Regions != nil {
		form.Regions = pq.StringArray(req.Regions)
	}
	return s.repo.CreateConsentForm(form)
}

func (s *ConsentFormService) UpdateConsentForm(formID uuid.UUID, req *dto.UpdateConsentFormRequest) (*models.ConsentForm, error) {
	form, err := s.repo.GetConsentFormByID(formID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		form.Name = *req.Name
	}
	if req.Title != nil {
		form.Title = *req.Title
	}
	if req.Description != nil {
		form.Description = *req.Description
	}
	if req.Department != nil {
		form.Department = *req.Department
	}
	if req.Project != nil {
		form.Project = *req.Project
	}
	if req.OrganizationEntityID != nil {
		if *req.OrganizationEntityID == "" {
			form.OrganizationEntityID = uuid.Nil
		} else {
			orgID, err := uuid.Parse(*req.OrganizationEntityID)
			if err != nil {
				return nil, fmt.Errorf("invalid organizationEntityId: %w", err)
			}
			form.OrganizationEntityID = orgID
		}
	}
	if req.DataRetentionPeriod != nil {
		form.DataRetentionPeriod = *req.DataRetentionPeriod
	}
	if req.UserRightsSummary != nil {
		form.UserRightsSummary = *req.UserRightsSummary
	}
	if req.TermsAndConditions != nil {
		form.TermsAndConditions = *req.TermsAndConditions
	}
	if req.PrivacyPolicy != nil {
		form.PrivacyPolicy = *req.PrivacyPolicy
	}
	if req.Translations != nil {
		translationsJSON, err := json.Marshal(req.Translations)
		if err != nil {
			return nil, fmt.Errorf("invalid translations: %w", err)
		}
		form.Translations = datatypes.JSON(translationsJSON)
	}
	if req.Regions != nil {
		form.Regions = pq.StringArray(req.Regions)
	}

	return s.repo.UpdateConsentForm(form)
}

func (s *ConsentFormService) DeleteConsentForm(formID uuid.UUID) error {
	return s.repo.DeleteConsentForm(formID)
}

func (s *ConsentFormService) GetConsentFormByID(formID uuid.UUID) (*models.ConsentForm, error) {
	return s.repo.GetConsentFormByID(formID)
}

func (s *ConsentFormService) ListConsentForms(tenantID uuid.UUID) ([]models.ConsentForm, error) {
	return s.repo.ListConsentForms(tenantID)
}

func (s *ConsentFormService) AddPurposeToConsentForm(formID uuid.UUID, req *dto.AddPurposeToConsentFormRequest) (*models.ConsentFormPurpose, error) {
	purposeID, err := uuid.Parse(req.PurposeID)
	if err != nil {
		return nil, err
	}
	return s.repo.AddPurposeToConsentForm(formID, purposeID, req.DataObjects, req.VendorIDs, req.ExpiryInDays)
}

func (s *ConsentFormService) UpdatePurposeInConsentForm(formID uuid.UUID, purposeID uuid.UUID, req *dto.UpdatePurposeInConsentFormRequest) (*models.ConsentFormPurpose, error) {
	return s.repo.UpdatePurposeInConsentForm(formID, purposeID, req.DataObjects, req.VendorIDs, req.ExpiryInDays)
}

func (s *ConsentFormService) RemovePurposeFromConsentForm(formID, purposeID uuid.UUID) error {
	return s.repo.RemovePurposeFromConsentForm(formID, purposeID)
}

func (s *ConsentFormService) GetIntegrationScript(formID uuid.UUID) *dto.IntegrationScriptResponse {
	cfg := config.LoadConfig()
	script := fmt.Sprintf(`<script>\n\tfunction openConsentForm() {\n\t\twindow.open('%s/#/consent/%s', 'ConsentForm', 'width=600,height=400');\n\t}\n</script>`, cfg.FrontendBaseURL, formID.String())

	return &dto.IntegrationScriptResponse{Script: script}
}

func (s *ConsentFormService) PublishConsentForm(formID uuid.UUID) error {
	return s.repo.PublishConsentForm(formID)
}

// ValidateForPublish validates a consent form before publishing
func (s *ConsentFormService) ValidateForPublish(formID uuid.UUID) (*dto.ValidateConsentFormResponse, error) {
	form, err := s.repo.GetConsentFormByID(formID)
	if err != nil {
		return nil, fmt.Errorf("failed to get consent form: %w", err)
	}

	var errors []dto.ValidationError
	summary := dto.ValidationSummary{}

	// Check required fields
	if strings.TrimSpace(form.Title) == "" {
		errors = append(errors, dto.ValidationError{
			Field:   "title",
			Message: "Title is required",
			Code:    "REQUIRED_FIELD_MISSING",
		})
	}
	if strings.TrimSpace(form.Description) == "" {
		errors = append(errors, dto.ValidationError{
			Field:   "description",
			Message: "Description is required",
			Code:    "REQUIRED_FIELD_MISSING",
		})
	}
	summary.RequiredFieldsComplete = len(errors) == 0

	// Check if at least one purpose is assigned
	if len(form.Purposes) == 0 {
		errors = append(errors, dto.ValidationError{
			Field:   "purposes",
			Message: "At least one purpose must be assigned",
			Code:    "NO_PURPOSES_ASSIGNED",
		})
		summary.PurposesAssigned = false
	} else {
		summary.PurposesAssigned = true
	}

	// Validate all purposes have data objects
	dataObjectsValid := true
	for _, purpose := range form.Purposes {
		if len(purpose.DataObjects) == 0 {
			errors = append(errors, dto.ValidationError{
				Field:   fmt.Sprintf("purposes[%s].dataObjects", purpose.PurposeID),
				Message: "Purpose must have at least one data object",
				Code:    "NO_DATA_OBJECTS",
			})
			dataObjectsValid = false
		}
	}
	summary.DataObjectsValid = dataObjectsValid

	// Validate expiry settings
	expiryValid := true
	for _, purpose := range form.Purposes {
		if purpose.ExpiryInDays <= 0 {
			errors = append(errors, dto.ValidationError{
				Field:   fmt.Sprintf("purposes[%s].expiryInDays", purpose.PurposeID),
				Message: "Expiry days must be greater than 0",
				Code:    "INVALID_EXPIRY",
			})
			expiryValid = false
		}
	}
	summary.ExpirySettingsValid = expiryValid

	// Check for duplicate purposes
	purposeMap := make(map[uuid.UUID]bool)
	noDuplicates := true
	for _, purpose := range form.Purposes {
		if purposeMap[purpose.PurposeID] {
			errors = append(errors, dto.ValidationError{
				Field:   "purposes",
				Message: fmt.Sprintf("Duplicate purpose found: %s", purpose.PurposeID),
				Code:    "DUPLICATE_PURPOSE",
			})
			noDuplicates = false
		}
		purposeMap[purpose.PurposeID] = true
	}
	summary.NoDuplicatePurposes = noDuplicates

	isValid := len(errors) == 0

	return &dto.ValidateConsentFormResponse{
		IsValid: isValid,
		Errors:  errors,
		Summary: summary,
	}, nil
}

// PublishConsentFormWithValidation publishes a consent form with validation and versioning
func (s *ConsentFormService) PublishConsentFormWithValidation(formID uuid.UUID, publishedBy uuid.UUID, changeLog string) error {
	// First validate the form
	validation, err := s.ValidateForPublish(formID)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !validation.IsValid {
		return fmt.Errorf("form validation failed: %d errors found", len(validation.Errors))
	}

	// Create version snapshot and publish
	return s.repo.PublishConsentFormWithVersioning(formID, publishedBy, changeLog)
}

// SubmitForReview changes form status to review
func (s *ConsentFormService) SubmitForReview(formID uuid.UUID, reviewNotes string) error {
	return s.repo.UpdateConsentFormStatus(formID, "review")
}

// GetVersionHistory returns all versions of a consent form
func (s *ConsentFormService) GetVersionHistory(formID uuid.UUID) ([]*dto.ConsentFormVersionResponse, error) {
	versions, err := s.repo.GetVersionHistory(formID)
	if err != nil {
		return nil, err
	}

	var response []*dto.ConsentFormVersionResponse
	for _, version := range versions {
		var snapshot map[string]interface{}
		if err := json.Unmarshal(version.Snapshot, &snapshot); err != nil {
			return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
		}

		response = append(response, &dto.ConsentFormVersionResponse{
			ID:            version.ID.String(),
			ConsentFormID: version.ConsentFormID.String(),
			VersionNumber: version.VersionNumber,
			Snapshot:      snapshot,
			PublishedAt:   version.PublishedAt,
			PublishedBy:   version.PublishedBy.String(),
			Status:        version.Status,
			ChangeLog:     version.ChangeLog,
			CreatedAt:     version.CreatedAt,
		})
	}

	return response, nil
}

// GetVersion returns a specific version of a consent form
func (s *ConsentFormService) GetVersion(versionID uuid.UUID) (*dto.ConsentFormVersionResponse, error) {
	version, err := s.repo.GetVersion(versionID)
	if err != nil {
		return nil, err
	}

	var snapshot map[string]interface{}
	if err := json.Unmarshal(version.Snapshot, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	return &dto.ConsentFormVersionResponse{
		ID:            version.ID.String(),
		ConsentFormID: version.ConsentFormID.String(),
		VersionNumber: version.VersionNumber,
		Snapshot:      snapshot,
		PublishedAt:   version.PublishedAt,
		PublishedBy:   version.PublishedBy.String(),
		Status:        version.Status,
		ChangeLog:     version.ChangeLog,
		CreatedAt:     version.CreatedAt,
	}, nil
}

// RollbackToVersion rolls back a consent form to a previous version
func (s *ConsentFormService) RollbackToVersion(formID, versionID uuid.UUID, rolledBackBy uuid.UUID, reason string) error {
	return s.repo.RollbackToVersion(formID, versionID, rolledBackBy, reason)
}

// AutoTranslateConsentForm translates the consent form to specified languages
func (s *ConsentFormService) AutoTranslateConsentForm(ctx context.Context, formID uuid.UUID, targetLangs []string) error {
	form, err := s.repo.GetConsentFormByID(formID)
	if err != nil {
		return err
	}

	if form.Translations == nil {
		form.Translations = datatypes.JSON([]byte("{}"))
	}

	var translations map[string]map[string]string
	if err := json.Unmarshal(form.Translations, &translations); err != nil {
		translations = make(map[string]map[string]string)
	}

	// Prepare content to translate
	contentToTranslate := map[string]string{
		"title":       form.Title,
		"description": form.Description,
	}
	// Add purposes to content
	for _, p := range form.Purposes {
		contentToTranslate[fmt.Sprintf("purpose_%s_name", p.PurposeID)] = p.Purpose.Name
		contentToTranslate[fmt.Sprintf("purpose_%s_description", p.PurposeID)] = p.Purpose.Description
	}

	for _, lang := range targetLangs {
		if _, exists := translations[lang]; exists {
			continue // Skip if already exists
		}

		translatedContent, err := s.translationService.TranslateMap(ctx, contentToTranslate, lang)
		if err != nil {
			log.Printf("Failed to translate to %s: %v", lang, err)
			continue
		}
		translations[lang] = translatedContent
	}

	translationsJSON, err := json.Marshal(translations)
	if err != nil {
		return err
	}

	form.Translations = datatypes.JSON(translationsJSON)
	_, err = s.repo.UpdateConsentForm(form)
	return err
}

