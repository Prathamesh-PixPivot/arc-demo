package repository

import (
	"pixpivot/arc/internal/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConsentFormRepository struct {
	db *gorm.DB
}

func NewConsentFormRepository(db *gorm.DB) *ConsentFormRepository {
	return &ConsentFormRepository{db: db}
}

func (r *ConsentFormRepository) CreateConsentForm(form *models.ConsentForm) (*models.ConsentForm, error) {
	if err := r.db.Create(form).Error; err != nil {
		return nil, err
	}
	return form, nil
}

func (r *ConsentFormRepository) UpdateConsentForm(form *models.ConsentForm) (*models.ConsentForm, error) {
	if err := r.db.Save(form).Error; err != nil {
		return nil, err
	}
	return form, nil
}

func (r *ConsentFormRepository) DeleteConsentForm(formID uuid.UUID) error {
	return r.db.Delete(&models.ConsentForm{}, formID).Error
}

func (r *ConsentFormRepository) GetConsentFormByID(formID uuid.UUID) (*models.ConsentForm, error) {
	var form models.ConsentForm
	if err := r.db.Preload("Purposes").Preload("Purposes.Purpose").First(&form, formID).Error; err != nil {
		return nil, err
	}
	return &form, nil
}

func (r *ConsentFormRepository) ListConsentForms(tenantID uuid.UUID) ([]models.ConsentForm, error) {
	var forms []models.ConsentForm
	if err := r.db.Where("tenant_id = ?", tenantID).Find(&forms).Error; err != nil {
		return nil, err
	}
	return forms, nil
}

func (r *ConsentFormRepository) AddPurposeToConsentForm(formID, purposeID uuid.UUID, dataObjects, vendorIDs []string, expiryInDays int) (*models.ConsentFormPurpose, error) {
	formPurpose := &models.ConsentFormPurpose{
		ID:            uuid.New(),
		ConsentFormID: formID,
		PurposeID:     purposeID,
		DataObjects:   dataObjects,
		VendorIDs:     vendorIDs,
		ExpiryInDays:  expiryInDays,
	}
	if err := r.db.Create(formPurpose).Error; err != nil {
		return nil, err
	}
	return formPurpose, nil
}

func (r *ConsentFormRepository) UpdatePurposeInConsentForm(formID, purposeID uuid.UUID, dataObjects, vendorIDs []string, expiryInDays int) (*models.ConsentFormPurpose, error) {
	var formPurpose models.ConsentFormPurpose
	if err := r.db.Where("consent_form_id = ? AND purpose_id = ?", formID, purposeID).First(&formPurpose).Error; err != nil {
		return nil, err
	}

	formPurpose.DataObjects = dataObjects
	formPurpose.VendorIDs = vendorIDs
	formPurpose.ExpiryInDays = expiryInDays

	if err := r.db.Save(&formPurpose).Error; err != nil {
		return nil, err
	}
	return &formPurpose, nil
}

func (r *ConsentFormRepository) RemovePurposeFromConsentForm(formID, purposeID uuid.UUID) error {
	return r.db.Where("consent_form_id = ? AND purpose_id = ?", formID, purposeID).Delete(&models.ConsentFormPurpose{}).Error
}

func (r *ConsentFormRepository) GetConsentFormPurpose(formID, purposeID uuid.UUID) (*models.ConsentFormPurpose, error) {
	var formPurpose models.ConsentFormPurpose
	if err := r.db.Where("consent_form_id = ? AND purpose_id = ?").First(&formPurpose).Error; err != nil {
		return nil, err
	}
	return &formPurpose, nil
}

func (r *ConsentFormRepository) PublishConsentForm(formID uuid.UUID) error {
	return r.db.Model(&models.ConsentForm{}).Where("id = ?", formID).Update("published", true).Error
}

// CreateVersion creates a new version snapshot of a consent form
func (r *ConsentFormRepository) CreateVersion(form *models.ConsentForm, publishedBy uuid.UUID, changeLog string) (*models.ConsentFormVersion, error) {
	// Create snapshot of the current form state
	snapshot := map[string]interface{}{
		"id":                   form.ID,
		"name":                 form.Name,
		"title":                form.Title,
		"description":          form.Description,
		"department":           form.Department,
		"project":              form.Project,
		"organizationEntityId": form.OrganizationEntityID,
		"dataRetentionPeriod":  form.DataRetentionPeriod,
		"userRightsSummary":    form.UserRightsSummary,
		"termsAndConditions":   form.TermsAndConditions,
		"privacyPolicy":        form.PrivacyPolicy,
		"purposes":             form.Purposes,
		"currentVersion":       form.CurrentVersion,
		"status":               form.Status,
	}

	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	version := &models.ConsentFormVersion{
		ID:            uuid.New(),
		ConsentFormID: form.ID,
		VersionNumber: form.CurrentVersion,
		Snapshot:      snapshotBytes,
		PublishedBy:   publishedBy,
		Status:        "published",
		ChangeLog:     changeLog,
	}

	if err := r.db.Create(version).Error; err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	return version, nil
}

// GetVersionHistory returns all versions of a consent form
func (r *ConsentFormRepository) GetVersionHistory(formID uuid.UUID) ([]*models.ConsentFormVersion, error) {
	var versions []*models.ConsentFormVersion
	if err := r.db.Where("consent_form_id = ?", formID).Order("version_number DESC").Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to get version history: %w", err)
	}
	return versions, nil
}

// GetVersion returns a specific version by ID
func (r *ConsentFormRepository) GetVersion(versionID uuid.UUID) (*models.ConsentFormVersion, error) {
	var version models.ConsentFormVersion
	if err := r.db.First(&version, versionID).Error; err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}
	return &version, nil
}

// RollbackToVersion rolls back a consent form to a previous version
func (r *ConsentFormRepository) RollbackToVersion(formID, versionID uuid.UUID, rolledBackBy uuid.UUID, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get the target version
		var targetVersion models.ConsentFormVersion
		if err := tx.First(&targetVersion, versionID).Error; err != nil {
			return fmt.Errorf("failed to get target version: %w", err)
		}

		// Verify the version belongs to the form
		if targetVersion.ConsentFormID != formID {
			return fmt.Errorf("version does not belong to the specified form")
		}

		// Get current form
		var form models.ConsentForm
		if err := tx.Preload("Purposes").First(&form, formID).Error; err != nil {
			return fmt.Errorf("failed to get current form: %w", err)
		}

		// Create a version snapshot of current state before rollback
		currentSnapshot := map[string]interface{}{
			"id":                   form.ID,
			"name":                 form.Name,
			"title":                form.Title,
			"description":          form.Description,
			"department":           form.Department,
			"project":              form.Project,
			"organizationEntityId": form.OrganizationEntityID,
			"dataRetentionPeriod":  form.DataRetentionPeriod,
			"userRightsSummary":    form.UserRightsSummary,
			"termsAndConditions":   form.TermsAndConditions,
			"privacyPolicy":        form.PrivacyPolicy,
			"purposes":             form.Purposes,
			"currentVersion":       form.CurrentVersion,
			"status":               form.Status,
		}

		currentSnapshotBytes, err := json.Marshal(currentSnapshot)
		if err != nil {
			return fmt.Errorf("failed to marshal current snapshot: %w", err)
		}

		// Create rollback version entry
		rollbackVersion := &models.ConsentFormVersion{
			ID:            uuid.New(),
			ConsentFormID: formID,
			VersionNumber: form.CurrentVersion + 1,
			Snapshot:      currentSnapshotBytes,
			PublishedBy:   rolledBackBy,
			Status:        "rollback",
			ChangeLog:     fmt.Sprintf("Rollback to version %d. Reason: %s", targetVersion.VersionNumber, reason),
		}

		if err := tx.Create(rollbackVersion).Error; err != nil {
			return fmt.Errorf("failed to create rollback version: %w", err)
		}

		// Parse target version snapshot
		var targetSnapshot map[string]interface{}
		if err := json.Unmarshal(targetVersion.Snapshot, &targetSnapshot); err != nil {
			return fmt.Errorf("failed to unmarshal target snapshot: %w", err)
		}

		// Update form with target version data
		updates := map[string]interface{}{
			"current_version":    form.CurrentVersion + 1,
			"status":            "published",
			"last_published_at": time.Now(),
			"last_published_by": rolledBackBy,
		}

		// Update basic fields if they exist in snapshot
		if name, ok := targetSnapshot["name"].(string); ok {
			updates["name"] = name
		}
		if title, ok := targetSnapshot["title"].(string); ok {
			updates["title"] = title
		}
		if description, ok := targetSnapshot["description"].(string); ok {
			updates["description"] = description
		}
		if department, ok := targetSnapshot["department"].(string); ok {
			updates["department"] = department
		}
		if project, ok := targetSnapshot["project"].(string); ok {
			updates["project"] = project
		}
		if dataRetentionPeriod, ok := targetSnapshot["dataRetentionPeriod"].(string); ok {
			updates["data_retention_period"] = dataRetentionPeriod
		}
		if userRightsSummary, ok := targetSnapshot["userRightsSummary"].(string); ok {
			updates["user_rights_summary"] = userRightsSummary
		}
		if termsAndConditions, ok := targetSnapshot["termsAndConditions"].(string); ok {
			updates["terms_and_conditions"] = termsAndConditions
		}
		if privacyPolicy, ok := targetSnapshot["privacyPolicy"].(string); ok {
			updates["privacy_policy"] = privacyPolicy
		}

		if err := tx.Model(&form).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update form: %w", err)
		}

		// Handle purposes rollback (simplified - in production you might want more sophisticated handling)
		// Delete current purposes and recreate from snapshot
		if err := tx.Where("consent_form_id = ?", formID).Delete(&models.ConsentFormPurpose{}).Error; err != nil {
			return fmt.Errorf("failed to delete current purposes: %w", err)
		}

		// Recreate purposes from snapshot if they exist
		if purposesData, ok := targetSnapshot["purposes"]; ok {
			purposesBytes, err := json.Marshal(purposesData)
			if err != nil {
				return fmt.Errorf("failed to marshal purposes: %w", err)
			}

			var purposes []models.ConsentFormPurpose
			if err := json.Unmarshal(purposesBytes, &purposes); err != nil {
				return fmt.Errorf("failed to unmarshal purposes: %w", err)
			}

			for _, purpose := range purposes {
				purpose.ID = uuid.New() // Generate new IDs
				purpose.ConsentFormID = formID
				if err := tx.Create(&purpose).Error; err != nil {
					return fmt.Errorf("failed to recreate purpose: %w", err)
				}
			}
		}

		return nil
	})
}

// PublishConsentFormWithVersioning publishes a consent form with validation and versioning
func (r *ConsentFormRepository) PublishConsentFormWithVersioning(formID uuid.UUID, publishedBy uuid.UUID, changeLog string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get the form with purposes
		var form models.ConsentForm
		if err := tx.Preload("Purposes").First(&form, formID).Error; err != nil {
			return fmt.Errorf("failed to get consent form: %w", err)
		}

		// Create version snapshot
		version := &models.ConsentFormVersion{
			ID:            uuid.New(),
			ConsentFormID: form.ID,
			VersionNumber: form.CurrentVersion,
			PublishedBy:   publishedBy,
			Status:        "published",
			ChangeLog:     changeLog,
		}

		// Create snapshot of the current form state
		snapshot := map[string]interface{}{
			"id":                   form.ID,
			"name":                 form.Name,
			"title":                form.Title,
			"description":          form.Description,
			"department":           form.Department,
			"project":              form.Project,
			"organizationEntityId": form.OrganizationEntityID,
			"dataRetentionPeriod":  form.DataRetentionPeriod,
			"userRightsSummary":    form.UserRightsSummary,
			"termsAndConditions":   form.TermsAndConditions,
			"privacyPolicy":        form.PrivacyPolicy,
			"purposes":             form.Purposes,
			"currentVersion":       form.CurrentVersion,
			"status":               "published",
		}

		snapshotBytes, err := json.Marshal(snapshot)
		if err != nil {
			return fmt.Errorf("failed to marshal snapshot: %w", err)
		}

		version.Snapshot = snapshotBytes

		// Create the version record
		if err := tx.Create(version).Error; err != nil {
			return fmt.Errorf("failed to create version: %w", err)
		}

		// Update the form
		now := time.Now()
		updates := map[string]interface{}{
			"published":         true,
			"status":           "published",
			"current_version":  form.CurrentVersion + 1,
			"last_published_at": &now,
			"last_published_by": &publishedBy,
		}

		if err := tx.Model(&form).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update form: %w", err)
		}

		return nil
	})
}

// UpdateConsentFormStatus updates the status of a consent form
func (r *ConsentFormRepository) UpdateConsentFormStatus(formID uuid.UUID, status string) error {
	return r.db.Model(&models.ConsentForm{}).Where("id = ?", formID).Update("status", status).Error
}

