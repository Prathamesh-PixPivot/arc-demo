package services

import (
	"bytes"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type TPRMService struct {
	repo        *repository.TPRMRepository
	baseURL     string
	storageType string
	storagePath string
	s3Bucket    string
	s3Client    *s3.S3
	s3Endpoint  string
	s3Region    string
	s3AccessKey string
	s3SecretKey string
	s3UseSSL    bool
	s3ForcePath bool
}

func NewTPRMService(
	repo *repository.TPRMRepository,
	baseURL, storageType, storagePath, s3Bucket string,
	s3Endpoint, s3AccessKey, s3SecretKey, s3Region string,
	useSSL, forcePathStyle bool,
) *TPRMService {
	svc := &TPRMService{
		repo:        repo,
		baseURL:     baseURL,
		storageType: storageType,
		storagePath: storagePath,
		s3Bucket:    s3Bucket,
		s3Endpoint:  s3Endpoint,
		s3AccessKey: s3AccessKey,
		s3SecretKey: s3SecretKey,
		s3Region:    s3Region,
		s3UseSSL:    useSSL,
		s3ForcePath: forcePathStyle,
	}

	if storageType == "s3" {
		sess, err := session.NewSession(&aws.Config{
			Region:           aws.String(svc.s3Region),
			Endpoint:         aws.String(svc.s3Endpoint),
			S3ForcePathStyle: aws.Bool(svc.s3ForcePath),
			DisableSSL:       aws.Bool(!svc.s3UseSSL),
			Credentials:      credentials.NewStaticCredentials(svc.s3AccessKey, svc.s3SecretKey, ""),
		})
		if err == nil {
			svc.s3Client = s3.New(sess)
		}
	}
	return svc
}

// CreateAssessment initializes a new assessment for a vendor under a tenant
func (s *TPRMService) CreateAssessment(tenantID, vendorID uuid.UUID, title, framework string, dueDate *time.Time, assessorID *uuid.UUID, notes string) (*models.TPRMAssessment, error) {
	a := &models.TPRMAssessment{
		ID:         uuid.New(),
		TenantID:   tenantID,
		VendorID:   vendorID,
		Title:      strings.TrimSpace(title),
		Framework:  strings.ToUpper(strings.TrimSpace(framework)),
		Status:     "pending",
		DueDate:    dueDate,
		AssessorID: assessorID,
		Notes:      notes,
		RiskScore:  0,
	}
	if err := s.repo.CreateAssessment(a); err != nil {
		return nil, err
	}
	return a, nil
}

// StoreEvidence saves evidence to local storage and tracks it in DB
func (s *TPRMService) StoreEvidence(tenantID, vendorID, assessmentID uuid.UUID, fileName, contentType string, data []byte, uploadedBy uuid.UUID) (*models.TPRMEvidence, error) {
	safeName := strings.ReplaceAll(fileName, "..", "")

	var storedPath string
	if s.storageType == "local" {
		dir := filepath.Join(s.storagePath, "evidence", tenantID.String(), assessmentID.String())
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		fp := filepath.Join(dir, safeName)
		if err := os.WriteFile(fp, data, 0o644); err != nil {
			return nil, err
		}
		storedPath = fp
	} else if s.storageType == "s3" && s.s3Client != nil {
		// S3 key convention
		s3Key := filepath.ToSlash(filepath.Join("evidence", tenantID.String(), assessmentID.String(), safeName))
		_, err := s.s3Client.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(s.s3Bucket),
			Key:         aws.String(s3Key),
			Body:        bytes.NewReader(data),
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload to s3: %w", err)
		}
		storedPath = s3Key
	} else {
		return nil, fmt.Errorf("unsupported storageType for evidence: %s", s.storageType)
	}

	e := &models.TPRMEvidence{
		ID:           uuid.New(),
		AssessmentID: assessmentID,
		TenantID:     tenantID,
		VendorID:     vendorID,
		FilePath:     storedPath,
		ContentType:  contentType,
		SizeBytes:    int64(len(data)),
		UploadedAt:   time.Now(),
		UploadedBy:   uploadedBy,
	}
	if err := s.repo.AddEvidence(e); err != nil {
		return nil, err
	}

	// bump evidence count on assessment
	a, err := s.repo.GetAssessmentByID(assessmentID)
	if err == nil && a != nil {
		a.EvidenceCount += 1
		_ = s.repo.UpdateAssessment(a)
	}
	return e, nil
}

// ListAssessments returns assessments filtered by tenant or vendor
func (s *TPRMService) ListAssessments(tenantID *uuid.UUID, vendorID *uuid.UUID, limit, offset int) ([]models.TPRMAssessment, error) {
	if tenantID != nil {
		return s.repo.ListAssessmentsByTenant(*tenantID, limit, offset)
	}
	if vendorID != nil {
		return s.repo.ListAssessmentsByVendor(*vendorID, limit, offset)
	}
	// default: by tenant is required; return empty
	return []models.TPRMAssessment{}, nil
}

// ListEvidence returns evidence for an assessment
func (s *TPRMService) ListEvidence(assessmentID uuid.UUID) ([]models.TPRMEvidence, error) {
	return s.repo.ListEvidenceByAssessment(assessmentID)
}

// GetEvidenceByID fetches a single evidence row
func (s *TPRMService) GetEvidenceByID(id uuid.UUID) (*models.TPRMEvidence, error) {
	return s.repo.GetEvidenceByID(id)
}

// PresignEvidenceURL returns a presigned URL if using S3
func (s *TPRMService) PresignEvidenceURL(e *models.TPRMEvidence, expiry time.Duration) (string, error) {
	if s.storageType == "s3" && s.s3Client != nil {
		req, _ := s.s3Client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(s.s3Bucket),
			Key:    aws.String(e.FilePath),
		})
		return req.Presign(expiry)
	}
	return "", fmt.Errorf("presign not available for storageType=%s", s.storageType)
}

// AddFinding records a finding for the assessment
func (s *TPRMService) AddFinding(tenantID, vendorID, assessmentID uuid.UUID, severity, title, description, remediation string) (*models.TPRMFinding, error) {
	sev := strings.ToLower(strings.TrimSpace(severity))
	if sev == "" {
		sev = "low"
	}
	f := &models.TPRMFinding{
		ID:           uuid.New(),
		AssessmentID: assessmentID,
		TenantID:     tenantID,
		VendorID:     vendorID,
		Severity:     sev,
		Title:        strings.TrimSpace(title),
		Description:  strings.TrimSpace(description),
		Remediation:  strings.TrimSpace(remediation),
		Status:       "open",
	}
	if err := s.repo.AddFinding(f); err != nil {
		return nil, err
	}

	// bump findings count on assessment
	a, err := s.repo.GetAssessmentByID(assessmentID)
	if err == nil && a != nil {
		a.FindingsCount += 1
		_ = s.repo.UpdateAssessment(a)
	}
	return f, nil
}

// ComputeRiskAndUpdate calculates risk from findings and updates assessment and vendor risk fields
func (s *TPRMService) ComputeRiskAndUpdate(assessmentID, vendorID uuid.UUID) (float64, string, error) {
	findings, err := s.repo.ListFindingsByAssessment(assessmentID)
	if err != nil {
		return 0, "", err
	}
	var sum float64
	for _, f := range findings {
		switch strings.ToLower(f.Severity) {
		case "critical":
			sum += 10
		case "high":
			sum += 6
		case "medium":
			sum += 3
		default:
			sum += 1
		}
	}
	// Normalize to 0..100
	risk := sum * 5
	if risk > 100 {
		risk = 100
	}
	level := riskLevel(risk)

	// update assessment
	a, err := s.repo.GetAssessmentByID(assessmentID)
	if err == nil && a != nil {
		a.RiskScore = risk
		a.Status = "completed"
		now := time.Now()
		a.CompletedAt = &now
		_ = s.repo.UpdateAssessment(a)
	}
	// update vendor risk snapshot
	if err := s.repo.UpdateVendorRisk(vendorID, risk, level); err != nil {
		return risk, level, err
	}
	return risk, level, nil
}

func riskLevel(score float64) string {
	switch {
	case score >= 81:
		return "critical"
	case score >= 51:
		return "high"
	case score >= 21:
		return "medium"
	default:
		return "low"
	}
}

// ==========================================
// DPDPA Specific Logic
// ==========================================

func (s *TPRMService) GetDPDPAChecklist() models.AuditChecklist {
	return models.AuditChecklist{
		ID:          "dpdpa-v1",
		Name:        "DPDPA 2023 Compliance Checklist",
		Description: "Standard audit checklist based on Digital Personal Data Protection Act, 2023",
		Categories: []models.AuditCategory{
			{
				Name: "Security Safeguards (Section 8(4))",
				Questions: []models.AuditQuestion{
					{ID: "sec-1", Text: "Are technical and organizational measures in place to prevent data breach?", ReferenceSection: "8(4)", RiskWeight: 10},
					{ID: "sec-2", Text: "Is personal data encrypted at rest and in transit?", ReferenceSection: "8(4)", RiskWeight: 8},
					{ID: "sec-3", Text: "Are access controls implemented on a need-to-know basis?", ReferenceSection: "8(4)", RiskWeight: 7},
				},
			},
			{
				Name: "Data Breach Notification (Section 8(6))",
				Questions: []models.AuditQuestion{
					{ID: "br-1", Text: "Is there a mechanism to notify the Data Fiduciary immediately upon breach?", ReferenceSection: "8(6)", RiskWeight: 10},
					{ID: "br-2", Text: "Is there an incident response plan?", ReferenceSection: "8(6)", RiskWeight: 6},
				},
			},
			{
				Name: "Data Erasure (Section 8(7))",
				Questions: []models.AuditQuestion{
					{ID: "del-1", Text: "Is there a process to erase data when consent is withdrawn?", ReferenceSection: "8(7)", RiskWeight: 9},
					{ID: "del-2", Text: "Is data erased when the purpose is no longer served?", ReferenceSection: "8(7)", RiskWeight: 8},
				},
			},
			{
				Name: "Sub-processing (Section 8(2))",
				Questions: []models.AuditQuestion{
					{ID: "sub-1", Text: "Is prior written consent obtained before engaging sub-processors?", ReferenceSection: "8(2)", RiskWeight: 8},
				},
			},
		},
	}
}

func (s *TPRMService) SubmitAuditResponse(assessmentID uuid.UUID, responses []models.AuditResponse) error {
	for _, r := range responses {
		r.AssessmentID = assessmentID
		if err := s.repo.SaveAuditResponse(&r); err != nil {
			return err
		}
	}

	// Auto-calculate risk based on responses
	// Simple logic: If any high-weight question is "no", risk increases
	checklist := s.GetDPDPAChecklist()
	var riskScore float64
	var maxScore float64

	responseMap := make(map[string]string)
	for _, r := range responses {
		responseMap[r.QuestionID] = r.Response
	}

	for _, cat := range checklist.Categories {
		for _, q := range cat.Questions {
			maxScore += float64(q.RiskWeight)
			if val, ok := responseMap[q.ID]; ok {
				if val == "no" {
					riskScore += float64(q.RiskWeight)
				}
			} else {
				// Unanswered treated as risk
				riskScore += float64(q.RiskWeight)
			}
		}
	}

	normalizedRisk := (riskScore / maxScore) * 100

	// Update assessment
	a, err := s.repo.GetAssessmentByID(assessmentID)
	if err == nil {
		a.RiskScore = normalizedRisk
		s.repo.UpdateAssessment(a)
	}

	return nil
}

func (s *TPRMService) CreateDPATemplate(tenantID uuid.UUID, name, content string) (*models.DPATemplate, error) {
	t := &models.DPATemplate{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     name,
		Content:  content,
		Version:  "1.0",
		IsActive: true,
	}
	return t, s.repo.CreateDPATemplate(t)
}

func (s *TPRMService) GenerateDPA(tenantID, vendorID, templateID uuid.UUID) (*models.DPAAgreement, error) {
	// In a real app, this would merge template content with vendor data and generate a PDF
	// For now, we create the agreement record
	a := &models.DPAAgreement{
		ID:         uuid.New(),
		TenantID:   tenantID,
		VendorID:   vendorID,
		TemplateID: &templateID,
		Status:     "generated",
	}
	return a, s.repo.CreateDPAAgreement(a)
}

func (s *TPRMService) UploadSignedDPA(agreementID uuid.UUID, fileData []byte) error {
	// Mock upload
	// In real implementation, upload to S3/Local and update SignedURL
	// For now, just update status
	a := &models.DPAAgreement{ID: agreementID, Status: "signed"} // Partial update
	// We need to fetch first to update properly or use a specific update method
	// Assuming simple update for now
	return s.repo.UpdateDPAAgreement(a)
}

