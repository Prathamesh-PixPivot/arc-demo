package services

import (
	"pixpivot/arc/internal/storage/repository"

	"pixpivot/arc/internal/models"
	"time"

	"github.com/google/uuid"
)

type DSRService struct {
	repo *repository.DSRRepository
}

func NewDSRService(repo *repository.DSRRepository) *DSRService {
	return &DSRService{repo: repo}
}

func (s *DSRService) CreateRequest(req *models.DSRRequest) error {
	// Calculate SLA based on type and regulation (defaulting to 30 days for GDPR/DPDP)
	req.DueDate = time.Now().AddDate(0, 0, 30)
	
	// Auto-assign priority based on type
	if req.Type == "erasure" || req.Type == "rectification" {
		req.Priority = "high"
	} else {
		req.Priority = "medium"
	}
	
	return s.repo.Create(req)
}

func (s *DSRService) UpdateStatus(id uuid.UUID, status string, note string, userID uuid.UUID) error {
	// TODO: Add validation for state transitions
	return s.repo.UpdateStatus(id, status, note)
}

func (s *DSRService) AddComment(comment *models.DSRComment) error {
	return s.repo.AddComment(comment)
}

func (s *DSRService) GetComments(requestID uuid.UUID) ([]models.DSRComment, error) {
	return s.repo.GetComments(requestID)
}

func (s *DSRService) ApproveDeleteRequest(requestID uuid.UUID) error {
	return s.repo.ApproveDeleteRequest(requestID)
}

