package repository

import (
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TPRMRepository handles data persistence for third-party risk management
type TPRMRepository struct {
	db *gorm.DB
}

func NewTPRMRepository(db *gorm.DB) *TPRMRepository {
	return &TPRMRepository{db: db}
}

// Assessments
func (r *TPRMRepository) CreateAssessment(a *models.TPRMAssessment) error {
	return r.db.Create(a).Error
}

func (r *TPRMRepository) UpdateAssessment(a *models.TPRMAssessment) error {
	return r.db.Save(a).Error
}

func (r *TPRMRepository) GetAssessmentByID(id uuid.UUID) (*models.TPRMAssessment, error) {
	var a models.TPRMAssessment
	if err := r.db.Where("id = ?", id).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *TPRMRepository) ListAssessmentsByTenant(tenantID uuid.UUID, limit, offset int) ([]models.TPRMAssessment, error) {
	var list []models.TPRMAssessment
	q := r.db.Where("tenant_id = ?", tenantID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *TPRMRepository) ListAssessmentsByVendor(vendorID uuid.UUID, limit, offset int) ([]models.TPRMAssessment, error) {
	var list []models.TPRMAssessment
	q := r.db.Where("vendor_id = ?", vendorID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Evidence
func (r *TPRMRepository) AddEvidence(e *models.TPRMEvidence) error {
	return r.db.Create(e).Error
}

func (r *TPRMRepository) ListEvidenceByAssessment(assessmentID uuid.UUID) ([]models.TPRMEvidence, error) {
	var list []models.TPRMEvidence
	if err := r.db.Where("assessment_id = ?", assessmentID).Order("uploaded_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *TPRMRepository) GetEvidenceByID(id uuid.UUID) (*models.TPRMEvidence, error) {
	var e models.TPRMEvidence
	if err := r.db.Where("id = ?", id).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

// Findings
func (r *TPRMRepository) AddFinding(f *models.TPRMFinding) error {
	return r.db.Create(f).Error
}

func (r *TPRMRepository) ListFindingsByAssessment(assessmentID uuid.UUID) ([]models.TPRMFinding, error) {
	var list []models.TPRMFinding
	if err := r.db.Where("assessment_id = ?", assessmentID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Vendor risk
func (r *TPRMRepository) UpdateVendorRisk(vendorID uuid.UUID, riskScore float64, riskLevel string) error {
	return r.db.Model(&models.Vendor{}).
		Where("vendor_id = ?", vendorID).
		Updates(map[string]interface{}{"risk_score": riskScore, "risk_level": riskLevel}).Error
}

// DPA Templates
func (r *TPRMRepository) CreateDPATemplate(t *models.DPATemplate) error {
	return r.db.Create(t).Error
}

func (r *TPRMRepository) GetDPATemplate(id uuid.UUID) (*models.DPATemplate, error) {
	var t models.DPATemplate
	err := r.db.First(&t, "id = ?", id).Error
	return &t, err
}

// DPA Agreements
func (r *TPRMRepository) CreateDPAAgreement(a *models.DPAAgreement) error {
	return r.db.Create(a).Error
}

func (r *TPRMRepository) UpdateDPAAgreement(a *models.DPAAgreement) error {
	return r.db.Save(a).Error
}

// Audit Responses
func (r *TPRMRepository) SaveAuditResponse(resp *models.AuditResponse) error {
	// Upsert based on AssessmentID + QuestionID
	var existing models.AuditResponse
	err := r.db.Where("assessment_id = ? AND question_id = ?", resp.AssessmentID, resp.QuestionID).First(&existing).Error
	if err == nil {
		resp.ID = existing.ID
		return r.db.Save(resp).Error
	}
	return r.db.Create(resp).Error
}

func (r *TPRMRepository) GetAuditResponses(assessmentID uuid.UUID) ([]models.AuditResponse, error) {
	var list []models.AuditResponse
	err := r.db.Where("assessment_id = ?", assessmentID).Find(&list).Error
	return list, err
}

