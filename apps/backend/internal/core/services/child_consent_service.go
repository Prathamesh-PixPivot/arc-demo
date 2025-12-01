package services

import (
	"errors"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"
	"time"

	"github.com/google/uuid"
)

type ChildConsentService struct {
	repo *repository.ChildConsentRepository
}

func NewChildConsentService(repo *repository.ChildConsentRepository) *ChildConsentService {
	return &ChildConsentService{repo: repo}
}

func (s *ChildConsentService) AddChild(parentID, tenantID uuid.UUID, name string, dob time.Time, relation string) (*models.ChildProfile, error) {
	// Basic validation: Check age (e.g., must be under 18)
	age := time.Since(dob).Hours() / 24 / 365
	if age >= 18 {
		return nil, errors.New("child must be under 18 years old")
	}

	child := &models.ChildProfile{
		ID:           uuid.New(),
		ParentID:     parentID,
		TenantID:     tenantID,
		Name:         name,
		DateOfBirth:  dob,
		Relationship: relation,
		IsActive:     true,
	}

	if err := s.repo.CreateChildProfile(child); err != nil {
		return nil, err
	}

	return child, nil
}

func (s *ChildConsentService) ListChildren(parentID, tenantID uuid.UUID) ([]models.ChildProfile, error) {
	return s.repo.ListChildrenByParent(parentID, tenantID)
}

func (s *ChildConsentService) CreateConsentRequest(childID, tenantID uuid.UUID, requestType, resourceName string, purposeID *uuid.UUID) (*models.ParentalConsentRequest, error) {
	// Verify child exists and belongs to tenant
	child, err := s.repo.GetChildProfile(childID, tenantID)
	if err != nil {
		return nil, err
	}

	req := &models.ParentalConsentRequest{
		ID:           uuid.New(),
		ChildID:      childID,
		ParentID:     child.ParentID,
		TenantID:     tenantID,
		RequestType:  requestType,
		ResourceName: resourceName,
		PurposeID:    purposeID,
		Status:       "pending",
		ExpiresAt:    time.Now().Add(72 * time.Hour), // 3 days expiry
	}

	if err := s.repo.CreateConsentRequest(req); err != nil {
		return nil, err
	}

	// TODO: Trigger notification to parent (email/push)

	return req, nil
}

func (s *ChildConsentService) ApproveRequest(requestID, parentID, tenantID uuid.UUID) error {
	req, err := s.repo.GetConsentRequest(requestID, tenantID)
	if err != nil {
		return err
	}

	if req.ParentID != parentID {
		return errors.New("unauthorized: parent ID mismatch")
	}

	if req.Status != "pending" {
		return errors.New("request is not pending")
	}

	if time.Now().After(req.ExpiresAt) {
		req.Status = "expired"
		s.repo.UpdateConsentRequest(req)
		return errors.New("request has expired")
	}

	req.Status = "approved"
	now := time.Now()
	req.ApprovedAt = &now

	return s.repo.UpdateConsentRequest(req)
}

func (s *ChildConsentService) RejectRequest(requestID, parentID, tenantID uuid.UUID) error {
	req, err := s.repo.GetConsentRequest(requestID, tenantID)
	if err != nil {
		return err
	}

	if req.ParentID != parentID {
		return errors.New("unauthorized: parent ID mismatch")
	}

	if req.Status != "pending" {
		return errors.New("request is not pending")
	}

	req.Status = "rejected"
	now := time.Now()
	req.RejectedAt = &now

	return s.repo.UpdateConsentRequest(req)
}

func (s *ChildConsentService) GetPendingRequests(parentID, tenantID uuid.UUID) ([]models.ParentalConsentRequest, error) {
	return s.repo.ListPendingRequests(parentID, tenantID)
}
