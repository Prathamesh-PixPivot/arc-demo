package repository

import (
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BreachImpactAssessmentRepository handles breach impact assessment data
type BreachImpactAssessmentRepository struct {
	db *gorm.DB
}

func NewBreachImpactAssessmentRepository(db *gorm.DB) *BreachImpactAssessmentRepository {
	return &BreachImpactAssessmentRepository{db: db}
}

func (r *BreachImpactAssessmentRepository) Create(assessment *models.BreachImpactAssessment) error {
	return r.db.Create(assessment).Error
}

func (r *BreachImpactAssessmentRepository) GetByBreachID(breachID uuid.UUID) (*models.BreachImpactAssessment, error) {
	var assessment models.BreachImpactAssessment
	err := r.db.Where("breach_id = ?", breachID).First(&assessment).Error
	return &assessment, err
}

func (r *BreachImpactAssessmentRepository) Update(assessment *models.BreachImpactAssessment) error {
	return r.db.Save(assessment).Error
}

// BreachStakeholderRepository handles breach stakeholder notifications
type BreachStakeholderRepository struct {
	db *gorm.DB
}

func NewBreachStakeholderRepository(db *gorm.DB) *BreachStakeholderRepository {
	return &BreachStakeholderRepository{db: db}
}

func (r *BreachStakeholderRepository) Create(stakeholder *models.BreachStakeholder) error {
	stakeholder.ID = uuid.New()
	return r.db.Create(stakeholder).Error
}

func (r *BreachStakeholderRepository) GetByBreachID(breachID uuid.UUID) ([]models.BreachStakeholder, error) {
	var stakeholders []models.BreachStakeholder
	err := r.db.Where("breach_id = ?", breachID).Find(&stakeholders).Error
	return stakeholders, err
}

func (r *BreachStakeholderRepository) GetByType(breachID uuid.UUID, stakeholderType string) ([]models.BreachStakeholder, error) {
	var stakeholders []models.BreachStakeholder
	err := r.db.Where("breach_id = ? AND stakeholder_type = ?", breachID, stakeholderType).Find(&stakeholders).Error
	return stakeholders, err
}

func (r *BreachStakeholderRepository) Update(stakeholder *models.BreachStakeholder) error {
	return r.db.Save(stakeholder).Error
}

// BreachWorkflowStageRepository handles workflow progression
type BreachWorkflowStageRepository struct {
	db *gorm.DB
}

func NewBreachWorkflowStageRepository(db *gorm.DB) *BreachWorkflowStageRepository {
	return &BreachWorkflowStageRepository{db: db}
}

func (r *BreachWorkflowStageRepository) Create(stage *models.BreachWorkflowStage) error {
	return r.db.Create(stage).Error
}

func (r *BreachWorkflowStageRepository) GetByBreachID(breachID uuid.UUID) ([]models.BreachWorkflowStage, error) {
	var stages []models.BreachWorkflowStage
	err := r.db.Where("breach_id = ?", breachID).Order("created_at ASC").Find(&stages).Error
	return stages, err
}

func (r *BreachWorkflowStageRepository) GetStage(breachID uuid.UUID, stage string) (*models.BreachWorkflowStage, error) {
	var workflowStage models.BreachWorkflowStage
	err := r.db.Where("breach_id = ? AND stage = ?", breachID, stage).First(&workflowStage).Error
	return &workflowStage, err
}

func (r *BreachWorkflowStageRepository) UpdateStageStatus(breachID uuid.UUID, stage string, status string) error {
	return r.db.Model(&models.BreachWorkflowStage{}).
		Where("breach_id = ? AND stage = ?", breachID, stage).
		Update("status", status).Error
}

func (r *BreachWorkflowStageRepository) ApproveStage(breachID uuid.UUID, stage string, approvedBy uuid.UUID) error {
	now := gorm.Expr("NOW()")
	return r.db.Model(&models.BreachWorkflowStage{}).
		Where("breach_id = ? AND stage = ?", breachID, stage).
		Updates(map[string]interface{}{
			"status":       "approved",
			"approved_by":  approvedBy,
			"approved_at":  now,
			"completed_at": now,
		}).Error
}

func (r *BreachWorkflowStageRepository) RejectStage(breachID uuid.UUID, stage string, rejectedBy uuid.UUID, reason string) error {
	now := gorm.Expr("NOW()")
	return r.db.Model(&models.BreachWorkflowStage{}).
		Where("breach_id = ? AND stage = ?", breachID, stage).
		Updates(map[string]interface{}{
			"status":           "rejected",
			"rejected_by":      rejectedBy,
			"rejected_at":      now,
			"rejection_reason": reason,
		}).Error
}

// BreachCommunicationRepository handles communication records
type BreachCommunicationRepository struct {
	db *gorm.DB
}

func NewBreachCommunicationRepository(db *gorm.DB) *BreachCommunicationRepository {
	return &BreachCommunicationRepository{db: db}
}

func (r *BreachCommunicationRepository) Create(communication *models.BreachCommunication) error {
	return r.db.Create(communication).Error
}

func (r *BreachCommunicationRepository) GetByBreachID(breachID uuid.UUID) ([]models.BreachCommunication, error) {
	var communications []models.BreachCommunication
	err := r.db.Where("breach_id = ?", breachID).Order("created_at DESC").Find(&communications).Error
	return communications, err
}

func (r *BreachCommunicationRepository) Update(communication *models.BreachCommunication) error {
	return r.db.Save(communication).Error
}

// BreachEvidenceRepository handles evidence storage
type BreachEvidenceRepository struct {
	db *gorm.DB
}

func NewBreachEvidenceRepository(db *gorm.DB) *BreachEvidenceRepository {
	return &BreachEvidenceRepository{db: db}
}

func (r *BreachEvidenceRepository) Create(evidence *models.BreachEvidence) error {
	evidence.ID = uuid.New()
	return r.db.Create(evidence).Error
}

func (r *BreachEvidenceRepository) GetByBreachID(breachID uuid.UUID) ([]models.BreachEvidence, error) {
	var evidences []models.BreachEvidence
	err := r.db.Where("breach_id = ?", breachID).Order("collected_at DESC").Find(&evidences).Error
	return evidences, err
}

func (r *BreachEvidenceRepository) GetByID(id uuid.UUID) (*models.BreachEvidence, error) {
	var evidence models.BreachEvidence
	err := r.db.First(&evidence, "id = ?", id).Error
	return &evidence, err
}

func (r *BreachEvidenceRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.BreachEvidence{}, "id = ?", id).Error
}

// BreachTimelineRepository handles audit trail
type BreachTimelineRepository struct {
	db *gorm.DB
}

func NewBreachTimelineRepository(db *gorm.DB) *BreachTimelineRepository {
	return &BreachTimelineRepository{db: db}
}

func (r *BreachTimelineRepository) Create(entry *models.BreachTimeline) error {
	entry.ID = uuid.New()
	return r.db.Create(entry).Error
}

func (r *BreachTimelineRepository) GetByBreachID(breachID uuid.UUID) ([]models.BreachTimeline, error) {
	var timeline []models.BreachTimeline
	err := r.db.Where("breach_id = ?", breachID).Order("occurred_at ASC").Find(&timeline).Error
	return timeline, err
}

// BreachNotificationTemplateRepository handles templates
type BreachNotificationTemplateRepository struct {
	db *gorm.DB
}

func NewBreachNotificationTemplateRepository(db *gorm.DB) *BreachNotificationTemplateRepository {
	return &BreachNotificationTemplateRepository{db: db}
}

func (r *BreachNotificationTemplateRepository) Create(template *models.BreachNotificationTemplate) error {
	template.ID = uuid.New()
	return r.db.Create(template).Error
}

func (r *BreachNotificationTemplateRepository) GetTemplate(templateName string, recipientType string) (*models.BreachNotificationTemplate, error) {
	var template models.BreachNotificationTemplate
	err := r.db.Where("template_name = ? AND recipient_type = ? AND is_active = true", templateName, recipientType).
		First(&template).Error
	return &template, err
}

func (r *BreachNotificationTemplateRepository) GetByTenant(tenantID uuid.UUID) ([]models.BreachNotificationTemplate, error) {
	var templates []models.BreachNotificationTemplate
	err := r.db.Where("tenant_id = ? AND is_active = true", tenantID).Find(&templates).Error
	return templates, err
}

func (r *BreachNotificationTemplateRepository) GetSystemTemplates() ([]models.BreachNotificationTemplate, error) {
	var templates []models.BreachNotificationTemplate
	err := r.db.Where("is_system_template = true AND is_active = true").Find(&templates).Error
	return templates, err
}

func (r *BreachNotificationTemplateRepository) Update(template *models.BreachNotificationTemplate) error {
	return r.db.Save(template).Error
}

func (r *BreachNotificationTemplateRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.BreachNotificationTemplate{}, "id = ? AND is_system_template = false", id).Error
}
