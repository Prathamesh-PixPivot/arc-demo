package services

import (
	"context"
	"encoding/json"

	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuditService struct {
	repo *repository.AuditRepo
}

func NewAuditService(repo *repository.AuditRepo) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Create(ctx context.Context, userID, tenantID, purposeID uuid.UUID, actionType, consentStatus, initiator, sourceIP, geoRegion, jurisdiction string, details map[string]interface{}) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	logEntry := &models.AuditLog{
		LogID:         uuid.New(),
		UserID:        userID,
		TenantID:      tenantID,
		PurposeID:     purposeID,
		ActionType:    actionType,
		ConsentStatus: consentStatus,
		Initiator:     initiator,
		SourceIP:      sourceIP,
		GeoRegion:     geoRegion,
		Jurisdiction:  jurisdiction,
		Details:       datatypes.JSON(detailsJSON),
	}

	return s.repo.Create(logEntry)
}

func (s *AuditService) GetConsentAuditLogs(tenantID string) ([]models.AuditLog, error) {
	return s.repo.GetByTenant(tenantID)
}

func (s *AuditService) LogAction(ctx context.Context, dto *dto.AuditLogRequest) error {
	return s.Create(ctx, uuid.MustParse(dto.UserID), uuid.MustParse(dto.TenantID), uuid.MustParse(dto.PurposeID), dto.ActionType, dto.ConsentStatus, dto.Initiator, dto.SourceIP, dto.GeoRegion, dto.Jurisdiction, dto.Details)
}

