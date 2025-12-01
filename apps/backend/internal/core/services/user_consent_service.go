package services

import (
	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UserConsentService struct {
	repo            *repository.UserConsentRepository
	consentFormRepo *repository.ConsentFormRepository
	receiptService  *ReceiptService
}

func NewUserConsentService(repo *repository.UserConsentRepository, consentFormRepo *repository.ConsentFormRepository, receiptService *ReceiptService) *UserConsentService {
	return &UserConsentService{repo: repo, consentFormRepo: consentFormRepo, receiptService: receiptService}
}

func (s *UserConsentService) SubmitConsent(userID, tenantID, formID uuid.UUID, req *dto.SubmitConsentRequest) error {
	form, err := s.consentFormRepo.GetConsentFormByID(formID)
	if err != nil {
		return err
	}

	for _, purposeConsent := range req.Purposes {
		purposeID, err := uuid.Parse(purposeConsent.PurposeID)
		if err != nil {
			continue // or handle error
		}

		var expiry *time.Time
		for _, formPurpose := range form.Purposes {
			if formPurpose.PurposeID == purposeID {
				if formPurpose.ExpiryInDays > 0 {
					now := time.Now()
					expiryDate := now.AddDate(0, 0, formPurpose.ExpiryInDays)
					expiry = &expiryDate
				}
				break
			}
		}

		userConsent := &models.UserConsent{
			ID:            uuid.New(),
			UserID:        userID,
			PurposeID:     purposeID,
			TenantID:      tenantID,
			ConsentFormID: formID,
			Status:        purposeConsent.Consented,
			ExpiresAt:     expiry,
		}

		createdConsent, err := s.repo.CreateUserConsent(userConsent)
		if err != nil {
			// Handle error, maybe rollback transaction
			continue
		}

		// Generate receipt for granted consents
		if purposeConsent.Consented && s.receiptService != nil {
			go func(consentID uuid.UUID) {
				// Generate receipt asynchronously to avoid blocking the response
				_, err := s.receiptService.GenerateReceipt(consentID)
				if err != nil {
					// Log error but don't fail the consent submission
					// TODO: Add proper logging
				}
			}(createdConsent.ID)
		}
	}

	return nil
}

func (s *UserConsentService) WithdrawConsent(userID, purposeID, tenantID uuid.UUID) error {
	userConsent, err := s.repo.GetUserConsent(userID, purposeID, tenantID)
	if err != nil {
		return err
	}

	userConsent.Status = false
	_, err = s.repo.UpdateUserConsent(userConsent)
	return err
}

func (s *UserConsentService) GetUserConsents(userID, tenantID uuid.UUID) ([]models.UserConsent, error) {
	return s.repo.ListUserConsents(userID, tenantID)
}

func (s *UserConsentService) GetUserConsentForPurpose(userID, purposeID, tenantID uuid.UUID) (*models.UserConsent, error) {
	return s.repo.GetUserConsent(userID, purposeID, tenantID)
}

// CreatePublicConsent creates a consent from public submission
// This is a simplified version for public consent submissions
func (s *UserConsentService) CreatePublicConsent(ctx context.Context, tenantID uuid.UUID, principal *models.DataPrincipal, req *dto.CreateUserConsentRequest) (*models.UserConsent, error) {
	// For now, create a simple consent record
	// TODO: Implement proper public consent creation with data principal handling

	// Create a basic user consent for the first purpose (simplified)
	if len(req.Purposes) == 0 {
		return nil, fmt.Errorf("no purposes provided")
	}

	purposeID, err := uuid.Parse(req.Purposes[0])
	if err != nil {
		return nil, fmt.Errorf("invalid purpose ID: %v", err)
	}

	userConsent := &models.UserConsent{
		ID:            uuid.New(),
		UserID:        principal.ID, // Using principal ID as user ID
		PurposeID:     purposeID,
		TenantID:      tenantID,
		ConsentFormID: req.ConsentFormID,
		Status:        true, // Assuming consent is granted
		ExpiresAt:     nil,  // TODO: Calculate expiry based on form settings
	}

	return s.repo.CreateUserConsent(userConsent)
}

// GetPublicConsentForm retrieves a consent form for public use
func (s *UserConsentService) GetPublicConsentForm(ctx context.Context, tenantID uuid.UUID, formID uuid.UUID) (interface{}, error) {
	// TODO: Implement proper public consent form retrieval
	// For now, return a simple response to fix compilation
	return map[string]interface{}{
		"form_id": formID,
		"tenant_id": tenantID,
		"title": "Consent Form",
		"description": "Please review and provide your consent",
		"purposes": []interface{}{},
	}, nil
}

