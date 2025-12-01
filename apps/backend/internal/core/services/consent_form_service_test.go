package services

import (
	"pixpivot/arc/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConsentFormRepository is a mock implementation of ConsentFormRepositoryInterface
type MockConsentFormRepository struct {
	mock.Mock
}

// Implement all required interface methods
func (m *MockConsentFormRepository) GetConsentFormByID(formID uuid.UUID) (*models.ConsentForm, error) {
	args := m.Called(formID)
	return args.Get(0).(*models.ConsentForm), args.Error(1)
}

func (m *MockConsentFormRepository) CreateConsentForm(form *models.ConsentForm) (*models.ConsentForm, error) {
	args := m.Called(form)
	return args.Get(0).(*models.ConsentForm), args.Error(1)
}

func (m *MockConsentFormRepository) UpdateConsentForm(form *models.ConsentForm) (*models.ConsentForm, error) {
	args := m.Called(form)
	return args.Get(0).(*models.ConsentForm), args.Error(1)
}

func (m *MockConsentFormRepository) DeleteConsentForm(formID uuid.UUID) error {
	args := m.Called(formID)
	return args.Error(0)
}

func (m *MockConsentFormRepository) ListConsentForms(tenantID uuid.UUID) ([]models.ConsentForm, error) {
	args := m.Called(tenantID)
	return args.Get(0).([]models.ConsentForm), args.Error(1)
}

func (m *MockConsentFormRepository) AddPurposeToConsentForm(formID, purposeID uuid.UUID, dataObjects, vendorIDs []string, expiryInDays int) (*models.ConsentFormPurpose, error) {
	args := m.Called(formID, purposeID, dataObjects, vendorIDs, expiryInDays)
	return args.Get(0).(*models.ConsentFormPurpose), args.Error(1)
}

func (m *MockConsentFormRepository) UpdatePurposeInConsentForm(formID, purposeID uuid.UUID, dataObjects, vendorIDs []string, expiryInDays int) (*models.ConsentFormPurpose, error) {
	args := m.Called(formID, purposeID, dataObjects, vendorIDs, expiryInDays)
	return args.Get(0).(*models.ConsentFormPurpose), args.Error(1)
}

func (m *MockConsentFormRepository) RemovePurposeFromConsentForm(formID, purposeID uuid.UUID) error {
	args := m.Called(formID, purposeID)
	return args.Error(0)
}

func (m *MockConsentFormRepository) GetConsentFormPurpose(formID, purposeID uuid.UUID) (*models.ConsentFormPurpose, error) {
	args := m.Called(formID, purposeID)
	return args.Get(0).(*models.ConsentFormPurpose), args.Error(1)
}

func (m *MockConsentFormRepository) PublishConsentForm(formID uuid.UUID) error {
	args := m.Called(formID)
	return args.Error(0)
}

func (m *MockConsentFormRepository) PublishConsentFormWithVersioning(formID uuid.UUID, publishedBy uuid.UUID, changeLog string) error {
	args := m.Called(formID, publishedBy, changeLog)
	return args.Error(0)
}

func (m *MockConsentFormRepository) UpdateConsentFormStatus(formID uuid.UUID, status string) error {
	args := m.Called(formID, status)
	return args.Error(0)
}

func (m *MockConsentFormRepository) GetVersionHistory(formID uuid.UUID) ([]*models.ConsentFormVersion, error) {
	args := m.Called(formID)
	return args.Get(0).([]*models.ConsentFormVersion), args.Error(1)
}

func (m *MockConsentFormRepository) GetVersion(versionID uuid.UUID) (*models.ConsentFormVersion, error) {
	args := m.Called(versionID)
	return args.Get(0).(*models.ConsentFormVersion), args.Error(1)
}

func (m *MockConsentFormRepository) RollbackToVersion(formID, versionID uuid.UUID, rolledBackBy uuid.UUID, reason string) error {
	args := m.Called(formID, versionID, rolledBackBy, reason)
	return args.Error(0)
}

func (m *MockConsentFormRepository) CreateVersion(form *models.ConsentForm, publishedBy uuid.UUID, changeLog string) (*models.ConsentFormVersion, error) {
	args := m.Called(form, publishedBy, changeLog)
	return args.Get(0).(*models.ConsentFormVersion), args.Error(1)
}

func TestValidateForPublish_ValidForm(t *testing.T) {
	// Arrange
	mockRepo := new(MockConsentFormRepository)
	service := &ConsentFormService{repo: mockRepo}
	
	formID := uuid.New()
	form := &models.ConsentForm{
		ID:          formID,
		Title:       "Test Form",
		Description: "Test Description",
		Purposes: []models.ConsentFormPurpose{
			{
				ID:           uuid.New(),
				PurposeID:    uuid.New(),
				DataObjects:  []string{"email", "name"},
				ExpiryInDays: 365,
			},
		},
	}

	mockRepo.On("GetConsentFormByID", formID).Return(form, nil)

	// Act
	result, err := service.ValidateForPublish(formID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Empty(t, result.Errors)
	assert.True(t, result.Summary.RequiredFieldsComplete)
	assert.True(t, result.Summary.PurposesAssigned)
	assert.True(t, result.Summary.DataObjectsValid)
	assert.True(t, result.Summary.ExpirySettingsValid)
	assert.True(t, result.Summary.NoDuplicatePurposes)

	mockRepo.AssertExpectations(t)
}

func TestValidateForPublish_MissingRequiredFields(t *testing.T) {
	// Arrange
	mockRepo := new(MockConsentFormRepository)
	service := &ConsentFormService{repo: mockRepo}
	
	formID := uuid.New()
	form := &models.ConsentForm{
		ID:          formID,
		Title:       "", // Missing title
		Description: "", // Missing description
		Purposes:    []models.ConsentFormPurpose{},
	}

	mockRepo.On("GetConsentFormByID", formID).Return(form, nil)

	// Act
	result, err := service.ValidateForPublish(formID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
	assert.Len(t, result.Errors, 3) // title, description, no purposes
	assert.False(t, result.Summary.RequiredFieldsComplete)
	assert.False(t, result.Summary.PurposesAssigned)

	// Check specific error messages
	errorMessages := make([]string, len(result.Errors))
	for i, err := range result.Errors {
		errorMessages[i] = err.Message
	}
	assert.Contains(t, errorMessages, "Title is required")
	assert.Contains(t, errorMessages, "Description is required")
	assert.Contains(t, errorMessages, "At least one purpose must be assigned")

	mockRepo.AssertExpectations(t)
}

func TestValidateForPublish_InvalidPurposes(t *testing.T) {
	// Arrange
	mockRepo := new(MockConsentFormRepository)
	service := &ConsentFormService{repo: mockRepo}
	
	formID := uuid.New()
	purposeID := uuid.New()
	form := &models.ConsentForm{
		ID:          formID,
		Title:       "Test Form",
		Description: "Test Description",
		Purposes: []models.ConsentFormPurpose{
			{
				ID:           uuid.New(),
				PurposeID:    purposeID,
				DataObjects:  []string{}, // No data objects
				ExpiryInDays: 0,          // Invalid expiry
			},
			{
				ID:           uuid.New(),
				PurposeID:    purposeID, // Duplicate purpose
				DataObjects:  []string{"email"},
				ExpiryInDays: 365,
			},
		},
	}

	mockRepo.On("GetConsentFormByID", formID).Return(form, nil)

	// Act
	result, err := service.ValidateForPublish(formID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, result.IsValid)
	assert.False(t, result.Summary.DataObjectsValid)
	assert.False(t, result.Summary.ExpirySettingsValid)
	assert.False(t, result.Summary.NoDuplicatePurposes)

	mockRepo.AssertExpectations(t)
}

func TestPublishConsentFormWithValidation_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockConsentFormRepository)
	service := &ConsentFormService{repo: mockRepo}
	
	formID := uuid.New()
	publishedBy := uuid.New()
	changeLog := "Initial publication"
	
	form := &models.ConsentForm{
		ID:          formID,
		Title:       "Test Form",
		Description: "Test Description",
		Purposes: []models.ConsentFormPurpose{
			{
				ID:           uuid.New(),
				PurposeID:    uuid.New(),
				DataObjects:  []string{"email", "name"},
				ExpiryInDays: 365,
			},
		},
	}

	mockRepo.On("GetConsentFormByID", formID).Return(form, nil)
	mockRepo.On("PublishConsentFormWithVersioning", formID, publishedBy, changeLog).Return(nil)

	// Act
	err := service.PublishConsentFormWithValidation(formID, publishedBy, changeLog)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPublishConsentFormWithValidation_ValidationFails(t *testing.T) {
	// Arrange
	mockRepo := new(MockConsentFormRepository)
	service := &ConsentFormService{repo: mockRepo}
	
	formID := uuid.New()
	publishedBy := uuid.New()
	changeLog := "Initial publication"
	
	form := &models.ConsentForm{
		ID:          formID,
		Title:       "", // Missing title - validation will fail
		Description: "Test Description",
		Purposes:    []models.ConsentFormPurpose{},
	}

	mockRepo.On("GetConsentFormByID", formID).Return(form, nil)

	// Act
	err := service.PublishConsentFormWithValidation(formID, publishedBy, changeLog)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "form validation failed")
	
	// Ensure PublishConsentFormWithVersioning was not called
	mockRepo.AssertNotCalled(t, "PublishConsentFormWithVersioning")
	mockRepo.AssertExpectations(t)
}

func TestSubmitForReview_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockConsentFormRepository)
	service := &ConsentFormService{repo: mockRepo}
	
	formID := uuid.New()
	reviewNotes := "Please review this form"

	mockRepo.On("UpdateConsentFormStatus", formID, "review").Return(nil)

	// Act
	err := service.SubmitForReview(formID, reviewNotes)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetVersionHistory_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockConsentFormRepository)
	service := &ConsentFormService{repo: mockRepo}
	
	formID := uuid.New()
	versions := []*models.ConsentFormVersion{
		{
			ID:            uuid.New(),
			ConsentFormID: formID,
			VersionNumber: 2,
			Status:        "published",
			ChangeLog:     "Updated terms",
		},
		{
			ID:            uuid.New(),
			ConsentFormID: formID,
			VersionNumber: 1,
			Status:        "published",
			ChangeLog:     "Initial version",
		},
	}

	mockRepo.On("GetVersionHistory", formID).Return(versions, nil)

	// Act
	result, err := service.GetVersionHistory(formID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, result[0].VersionNumber)
	assert.Equal(t, 1, result[1].VersionNumber)

	mockRepo.AssertExpectations(t)
}

func TestRollbackToVersion_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockConsentFormRepository)
	service := &ConsentFormService{repo: mockRepo}
	
	formID := uuid.New()
	versionID := uuid.New()
	rolledBackBy := uuid.New()
	reason := "Reverting due to issues"

	mockRepo.On("RollbackToVersion", formID, versionID, rolledBackBy, reason).Return(nil)

	// Act
	err := service.RollbackToVersion(formID, versionID, rolledBackBy, reason)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

