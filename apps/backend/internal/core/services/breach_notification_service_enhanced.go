package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"

	"github.com/google/uuid"
)

type EnhancedBreachNotificationService struct {
	repo                 *repository.BreachNotificationRepository
	impactAssessmentRepo *repository.BreachImpactAssessmentRepository
	stakeholderRepo      *repository.BreachStakeholderRepository
	workflowRepo         *repository.BreachWorkflowStageRepository
	communicationRepo    *repository.BreachCommunicationRepository
	evidenceRepo         *repository.BreachEvidenceRepository
	timelineRepo         *repository.BreachTimelineRepository
	templateRepo         *repository.BreachNotificationTemplateRepository
	emailService         *EmailService
}

func NewEnhancedBreachNotificationService(
	repo *repository.BreachNotificationRepository,
	impactRepo *repository.BreachImpactAssessmentRepository,
	stakeholderRepo *repository.BreachStakeholderRepository,
	workflowRepo *repository.BreachWorkflowStageRepository,
	commRepo *repository.BreachCommunicationRepository,
	evidenceRepo *repository.BreachEvidenceRepository,
	timelineRepo *repository.BreachTimelineRepository,
	templateRepo *repository.BreachNotificationTemplateRepository,
	emailService *EmailService,
) *EnhancedBreachNotificationService {
	return &EnhancedBreachNotificationService{
		repo:                 repo,
		impactAssessmentRepo: impactRepo,
		stakeholderRepo:      stakeholderRepo,
		workflowRepo:         workflowRepo,
		communicationRepo:    commRepo,
		evidenceRepo:         evidenceRepo,
		timelineRepo:         timelineRepo,
		templateRepo:         templateRepo,
		emailService:         emailService,
	}
}

// CalculateDPDPDeadlines calculates notification deadlines per DPDP Act
func (s *EnhancedBreachNotificationService) CalculateDPDPDeadlines(breach *models.BreachNotification) {
	// DPDP requires notification "without undue delay"
	// Industry best practice: 72 hours for DPB, reasonable timeframe for individuals

	detectionDate := breach.DetectionDate
	if detectionDate.IsZero() {
		detectionDate = time.Now()
	}

	// DPB notification deadline: 72 hours from detection (following GDPR best practice)
	dpbDeadline := detectionDate.Add(72 * time.Hour)
	breach.DPBNotificationDeadline = &dpbDeadline

	// Data Principal notification: 7 days from detection for non-critical, 24 hours for critical
	var dataPrincipalHours int
	if breach.Severity == "critical" || breach.Severity == "high" {
		dataPrincipalHours = 24
	} else {
		dataPrincipalHours = 7 * 24 // 7 days
	}
	dataPrincipalDeadline := detectionDate.Add(time.Duration(dataPrincipalHours) * time.Hour)
	breach.DataPrincipalNotificationDeadline = &dataPrincipalDeadline

	// Check if overdue
	now := time.Now()
	if breach.DPBNotificationDeadline != nil && now.After(*breach.DPBNotificationDeadline) && !breach.DPBReported {
		breach.IsOverdue = true
	}
}

// CreateBreachWithWorkflow creates a breach and initializes workflow
func (s *EnhancedBreachNotificationService) CreateBreachWithWorkflow(
	breach *models.BreachNotification,
	createdBy uuid.UUID,
) error {
	// Generate ID and calculate deadlines
	breach.ID = uuid.New()
	breach.CurrentWorkflowStage = "detection"
	breach.Status = "draft"
	s.CalculateDPDPDeadlines(breach)

	// Create breach
	if err := s.repo.CreateBreachNotification(breach); err != nil {
		return err
	}

	// Create initial workflow stages
	stages := []string{"detection", "assessment", "containment", "verification", "notification", "resolution"}
	for i, stage := range stages {
		workflowStage := &models.BreachWorkflowStage{
			ID:               uuid.New(),
			BreachID:         breach.ID,
			TenantID:         breach.TenantID,
			Stage:            stage,
			Status:           "pending",
			RequiresApproval: stage == "verification" || stage == "notification",
		}

		if i == 0 {
			workflowStage.Status = "in_progress"
		}

		if err := s.workflowRepo.Create(workflowStage); err != nil {
			return err
		}
	}

	// Create timeline entry
	timelineEntry := &models.BreachTimeline{
		ID:          uuid.New(),
		BreachID:    breach.ID,
		TenantID:    breach.TenantID,
		EventType:   "breach_created",
		Description: "Breach notification created",
		PerformedBy: &createdBy,
	}
	return s.timelineRepo.Create(timelineEntry)
}

// SubmitForVerification moves breach to verification stage
func (s *EnhancedBreachNotificationService) SubmitForVerification(
	breachID uuid.UUID,
	submittedBy uuid.UUID,
) error {
	breach, err := s.repo.GetBreachNotificationByID(breachID)
	if err != nil {
		return err
	}

	breach.CurrentWorkflowStage = "verification"
	breach.Status = "pending_verification"

	if err := s.repo.UpdateBreachNotification(breach); err != nil {
		return err
	}

	// Update workflow stage
	if err := s.workflowRepo.UpdateStageStatus(breachID, "verification", "in_progress"); err != nil {
		return err
	}

	// Create timeline entry
	timelineEntry := &models.BreachTimeline{
		ID:          uuid.New(),
		BreachID:    breachID,
		TenantID:    breach.TenantID,
		EventType:   "submitted_for_verification",
		Description: "Breach submitted for verification",
		PerformedBy: &submittedBy,
	}
	return s.timelineRepo.Create(timelineEntry)
}

// VerifyBreach verifies a breach (approves for notification)
func (s *EnhancedBreachNotificationService) VerifyBreach(
	breachID uuid.UUID,
	verifiedBy uuid.UUID,
	approved bool,
	rejectionReason string,
) error {
	breach, err := s.repo.GetBreachNotificationByID(breachID)
	if err != nil {
		return err
	}

	now := time.Now()

	if approved {
		breach.Status = "verified"
		breach.VerifiedBy = &verifiedBy
		breach.VerifiedAt = &now
		breach.CurrentWorkflowStage = "notification"

		if err := s.workflowRepo.ApproveStage(breachID, "verification", verifiedBy); err != nil {
			return err
		}

		// Mark notification stage as ready
		if err := s.workflowRepo.UpdateStageStatus(breachID, "notification", "pending"); err != nil {
			return err
		}
	} else {
		breach.Status = "rejected"

		if err := s.workflowRepo.RejectStage(breachID, "verification", verifiedBy, rejectionReason); err != nil {
			return err
		}
	}

	if err := s.repo.UpdateBreachNotification(breach); err != nil {
		return err
	}

	// Create timeline entry
	eventType := "breach_verified"
	if !approved {
		eventType = "breach_rejected"
	}
	timelineEntry := &models.BreachTimeline{
		ID:          uuid.New(),
		BreachID:    breachID,
		TenantID:    breach.TenantID,
		EventType:   eventType,
		Description: fmt.Sprintf("Breach %s by verifier", eventType),
		PerformedBy: &verifiedBy,
	}
	return s.timelineRepo.Create(timelineEntry)
}

// ApproveDataPrincipalNotification approves sending notifications to affected individuals
func (s *EnhancedBreachNotificationService) ApproveDataPrincipalNotification(
	breachID uuid.UUID,
	approvedBy uuid.UUID,
) error {
	breach, err := s.repo.GetBreachNotificationByID(breachID)
	if err != nil {
		return err
	}

	if breach.Status != "verified" {
		return errors.New("breach must be verified before approving data principal notification")
	}

	now := time.Now()
	breach.DataPrincipalNotificationApproved = true
	breach.DataPrincipalNotificationApprovedBy = &approvedBy
	breach.DataPrincipalNotificationApprovedAt = &now

	if err := s.repo.UpdateBreachNotification(breach); err != nil {
		return err
	}

	// Create timeline entry
	timelineEntry := &models.BreachTimeline{
		ID:          uuid.New(),
		BreachID:    breachID,
		TenantID:    breach.TenantID,
		EventType:   "data_principal_notification_approved",
		Description: "Data Principal notification approved",
		PerformedBy: &approvedBy,
	}
	return s.timelineRepo.Create(timelineEntry)
}

// SendDPBNotification sends notification to Data Protection Board
func (s *EnhancedBreachNotificationService) SendDPBNotification(
	breachID uuid.UUID,
	sentBy uuid.UUID,
) error {
	breach, err := s.repo.GetBreachNotificationByID(breachID)
	if err != nil {
		return err
	}

	if breach.Status != "verified" {
		return errors.New("breach must be verified before DPB notification")
	}

	// Get DPB template
	template, err := s.templateRepo.GetTemplate("dpb_notification_template", "dpb")
	if err != nil {
		return fmt.Errorf("DPB notification template not found: %w", err)
	}

	// Prepare notification content
	content := s.renderTemplate(template.Body, breach)

	// Create DPB stakeholder
	dpbStakeholder := &models.BreachStakeholder{
		ID:                 uuid.New(),
		BreachID:           breachID,
		TenantID:           breach.TenantID,
		StakeholderType:    "dpb",
		ContactName:        "Data Protection Board of India",
		ContactEmail:       "dpb@meity.gov.in", // Official DPB contact
		NotificationMethod: "email",
		NotificationStatus: "pending",
	}

	if err := s.stakeholderRepo.Create(dpbStakeholder); err != nil {
		return err
	}

	// Send email
	subject := fmt.Sprintf("Data Breach Notification - %s", breach.Title)
	if err := s.emailService.Send(dpbStakeholder.ContactEmail, subject, content); err != nil {
		dpbStakeholder.NotificationStatus = "failed"
		s.stakeholderRepo.Update(dpbStakeholder)
		return fmt.Errorf("failed to send DPB notification: %w", err)
	}

	// Update stakeholder
	now := time.Now()
	dpbStakeholder.NotificationSent = true
	dpbStakeholder.NotifiedAt = &now
	dpbStakeholder.NotificationStatus = "sent"

	if err := s.stakeholderRepo.Update(dpbStakeholder); err != nil {
		return err
	}

	// Update breach
	breach.DPBReported = true
	breach.DPBReportedDate = &now
	breach.Status = "notifying"

	if err := s.repo.UpdateBreachNotification(breach); err != nil {
		return err
	}

	// Create communication record
	communication := &models.BreachCommunication{
		ID:                uuid.New(),
		BreachID:          breachID,
		TenantID:          breach.TenantID,
		CommunicationType: "initial_notification",
		Recipient:         dpbStakeholder.ContactEmail,
		RecipientType:     "dpb",
		Subject:           subject,
		Content:           content,
		TemplateName:      template.TemplateName,
		SendMethod:        "email",
		SentAt:            &now,
		Status:            "sent",
		CreatedBy:         sentBy,
	}

	if err := s.communicationRepo.Create(communication); err != nil {
		return err
	}

	// Create timeline entry
	timelineEntry := &models.BreachTimeline{
		ID:          uuid.New(),
		BreachID:    breachID,
		TenantID:    breach.TenantID,
		EventType:   "dpb_notified",
		Description: "Data Protection Board notified",
		PerformedBy: &sentBy,
	}
	return s.timelineRepo.Create(timelineEntry)
}

// SendDataPrincipalNotifications sends notifications to affected individuals
func (s *EnhancedBreachNotificationService) SendDataPrincipalNotifications(
	breachID uuid.UUID,
	affectedEmails []string,
	sentBy uuid.UUID,
) error {
	breach, err := s.repo.GetBreachNotificationByID(breachID)
	if err != nil {
		return err
	}

	if !breach.DataPrincipalNotificationApproved {
		return errors.New("data principal notification must be approved first")
	}

	// Get template
	template, err := s.templateRepo.GetTemplate("data_principal_notification_template", "data_principal")
	if err != nil {
		return fmt.Errorf("data principal notification template not found: %w", err)
	}

	// Prepare content
	content := s.renderTemplate(template.Body, breach)
	subject := fmt.Sprintf("Important Security Notice - Data Breach Notification")

	successCount := 0
	failCount := 0

	// Send to each affected individual
	for _, email := range affectedEmails {
		stakeholder := &models.BreachStakeholder{
			ID:                 uuid.UUID{},
			BreachID:           breachID,
			TenantID:           breach.TenantID,
			StakeholderType:    "affected_individual",
			ContactEmail:       email,
			NotificationMethod: "email",
			NotificationStatus: "pending",
		}

		if err := s.stakeholderRepo.Create(stakeholder); err != nil {
			failCount++
			continue
		}

		if err := s.emailService.Send(email, subject, content); err != nil {
			stakeholder.NotificationStatus = "failed"
			s.stakeholderRepo.Update(stakeholder)
			failCount++
			continue
		}

		now := time.Now()
		stakeholder.NotificationSent = true
		stakeholder.NotifiedAt = &now
		stakeholder.NotificationStatus = "sent"
		s.stakeholderRepo.Update(stakeholder)
		successCount++
	}

	// Update breach
	now := time.Now()
	breach.DataPrincipalNotificationSentAt = &now
	breach.NotifiedUsersCount = successCount
	breach.Status = "notified"

	if err := s.repo.UpdateBreachNotification(breach); err != nil {
		return err
	}

	// Create timeline entry
	timelineEntry := &models.BreachTimeline{
		ID:          uuid.New(),
		BreachID:    breachID,
		TenantID:    breach.TenantID,
		EventType:   "data_principals_notified",
		Description: fmt.Sprintf("Notified %d affected individuals (%d failed)", successCount, failCount),
		PerformedBy: &sentBy,
	}
	return s.timelineRepo.Create(timelineEntry)
}

// CheckSLACompliance checks if breach notifications are within SLA
func (s *EnhancedBreachNotificationService) CheckSLACompliance(breachID uuid.UUID) (map[string]interface{}, error) {
	breach, err := s.repo.GetBreachNotificationByID(breachID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	slaStatus := make(map[string]interface{})

	// DPB notification SLA
	if breach.DPBNotificationDeadline != nil {
		slaStatus["dpb_deadline"] = breach.DPBNotificationDeadline.Format(time.RFC3339)
		slaStatus["dpb_notified"] = breach.DPBReported

		if breach.DPBReported && breach.DPBReportedDate != nil {
			slaStatus["dpb_notified_at"] = breach.DPBReportedDate.Format(time.RFC3339)
			slaStatus["dpb_within_sla"] = breach.DPBReportedDate.Before(*breach.DPBNotificationDeadline)
		} else {
			slaStatus["dpb_within_sla"] = now.Before(*breach.DPBNotificationDeadline)
			slaStatus["dpb_overdue"] = now.After(*breach.DPBNotificationDeadline)
		}
	}

	// Data Principal notification SLA
	if breach.DataPrincipalNotificationDeadline != nil {
		slaStatus["data_principal_deadline"] = breach.DataPrincipalNotificationDeadline.Format(time.RFC3339)
		slaStatus["data_principal_notified"] = breach.DataPrincipalNotificationSentAt != nil

		if breach.DataPrincipalNotificationSentAt != nil {
			slaStatus["data_principal_notified_at"] = breach.DataPrincipalNotificationSentAt.Format(time.RFC3339)
			slaStatus["data_principal_within_sla"] = breach.DataPrincipalNotificationSentAt.Before(*breach.DataPrincipalNotificationDeadline)
		} else {
			slaStatus["data_principal_within_sla"] = now.Before(*breach.DataPrincipalNotificationDeadline)
			slaStatus["data_principal_overdue"] = now.After(*breach.DataPrincipalNotificationDeadline)
		}
	}

	slaStatus["is_overdue"] = breach.IsOverdue

	return slaStatus, nil
}

// renderTemplate renders a notification template with breach data
func (s *EnhancedBreachNotificationService) renderTemplate(template string, breach *models.BreachNotification) string {
	result := template

	replacements := map[string]string{
		"{{breach_title}}":       breach.Title,
		"{{breach_description}}": breach.Description,
		"{{breach_date}}":        breach.BreachDate.Format("2006-01-02"),
		"{{detection_date}}":     breach.DetectionDate.Format("2006-01-02"),
		"{{affected_count}}":     fmt.Sprintf("%d", breach.AffectedUsersCount),
		"{{severity}}":           breach.Severity,
		"{{breach_type}}":        breach.BreachType,
	}

	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// GetBreachRegister returns all breaches for compliance register
func (s *EnhancedBreachNotificationService) GetBreachRegister(tenantID uuid.UUID) ([]models.BreachNotification, error) {
	return s.repo.ListBreachNotifications(tenantID)
}

// EscalateOverdueBreaches identifies and escalates overdue breach notifications
func (s *EnhancedBreachNotificationService) EscalateOverdueBreaches(tenantID uuid.UUID) error {
	breaches, err := s.repo.ListBreachNotifications(tenantID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, breach := range breaches {
		// Check if DPB notification is overdue
		if !breach.DPBReported && breach.DPBNotificationDeadline != nil && now.After(*breach.DPBNotificationDeadline) {
			breach.IsOverdue = true
			s.repo.UpdateBreachNotification(&breach)

			// Create escalation timeline
			timelineEntry := &models.BreachTimeline{
				ID:          uuid.New(),
				BreachID:    breach.ID,
				TenantID:    breach.TenantID,
				EventType:   "escalation",
				Description: "DPB notification deadline exceeded - escalated",
			}
			s.timelineRepo.Create(timelineEntry)

			// TODO: Send escalation notification to admins
		}
	}

	return nil
}
