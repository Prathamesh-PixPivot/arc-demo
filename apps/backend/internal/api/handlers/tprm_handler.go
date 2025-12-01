package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TPRMHandler struct {
	service *services.TPRMService
}

func NewTPRMHandler(service *services.TPRMService) *TPRMHandler {
	return &TPRMHandler{service: service}
}

func (h *TPRMHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/assessments", h.CreateAssessment).Methods("POST")
	r.HandleFunc("/assessments", h.ListAssessments).Methods("GET")
	r.HandleFunc("/assessments/checklist/dpdpa", h.GetDPDPAChecklist).Methods("GET")
	r.HandleFunc("/assessments/{assessmentId}/submit", h.SubmitAuditResponse).Methods("POST")
	r.HandleFunc("/assessments/{assessmentId}/evidence", h.UploadEvidence).Methods("POST")
	r.HandleFunc("/assessments/{assessmentId}/evidence", h.ListEvidence).Methods("GET")
	r.HandleFunc("/assessments/{assessmentId}/findings", h.AddFinding).Methods("POST")
	r.HandleFunc("/assessments/{assessmentId}/compute-risk", h.ComputeRisk).Methods("POST")
	r.HandleFunc("/evidence/{evidenceId}", h.GetEvidenceMeta).Methods("GET")
	r.HandleFunc("/evidence/{evidenceId}/file", h.DownloadEvidenceFile).Methods("GET")

	// DPA Routes
	r.HandleFunc("/dpa/templates", h.CreateDPATemplate).Methods("POST")
	r.HandleFunc("/vendors/{vendorId}/dpa/generate", h.GenerateDPA).Methods("POST")
	r.HandleFunc("/vendors/{vendorId}/dpa/upload", h.UploadSignedDPA).Methods("POST")
}

type CreateAssessmentRequest struct {
	TenantID   string     `json:"tenantId"`
	VendorID   string     `json:"vendorId"`
	Title      string     `json:"title"`
	Framework  string     `json:"framework"`
	DueDate    *time.Time `json:"dueDate"`
	AssessorID *string    `json:"assessorId"`
	Notes      string     `json:"notes"`
}

func (h *TPRMHandler) CreateAssessment(w http.ResponseWriter, r *http.Request) {
	var req CreateAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		http.Error(w, "invalid tenantId", http.StatusBadRequest)
		return
	}
	vendorID, err := uuid.Parse(req.VendorID)
	if err != nil {
		http.Error(w, "invalid vendorId", http.StatusBadRequest)
		return
	}
	var assessorID *uuid.UUID
	if req.AssessorID != nil && *req.AssessorID != "" {
		id, err := uuid.Parse(*req.AssessorID)
		if err == nil {
			assessorID = &id
		}
	}
	a, err := h.service.CreateAssessment(tenantID, vendorID, req.Title, req.Framework, req.DueDate, assessorID, req.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

func (h *TPRMHandler) UploadEvidence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assessmentIDStr := vars["assessmentId"]
	assessmentID, err := uuid.Parse(assessmentIDStr)
	if err != nil {
		http.Error(w, "invalid assessmentId", http.StatusBadRequest)
		return
	}

	// Parse form data up to a reasonable size, e.g., 25MB
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}
	tenantIDStr := r.FormValue("tenantId")
	vendorIDStr := r.FormValue("vendorId")
	uploadedByStr := r.FormValue("uploadedBy")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "invalid tenantId", http.StatusBadRequest)
		return
	}
	vendorID, err := uuid.Parse(vendorIDStr)
	if err != nil {
		http.Error(w, "invalid vendorId", http.StatusBadRequest)
		return
	}
	uploadedBy, err := uuid.Parse(uploadedByStr)
	if err != nil {
		http.Error(w, "invalid uploadedBy", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file not provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buf := make([]byte, header.Size)
	n, err := file.Read(buf)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	data := buf[:n]

	e, err := h.service.StoreEvidence(tenantID, vendorID, assessmentID, header.Filename, header.Header.Get("Content-Type"), data, uploadedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

// ListAssessments returns assessments filtered by tenantId or vendorId
func (h *TPRMHandler) ListAssessments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	tenantIDStr := q.Get("tenantId")
	vendorIDStr := q.Get("vendorId")
	limitStr := q.Get("limit")
	offsetStr := q.Get("offset")

	var tenantID *uuid.UUID
	var vendorID *uuid.UUID
	if tenantIDStr != "" {
		if id, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &id
		} else {
			http.Error(w, "invalid tenantId", http.StatusBadRequest)
			return
		}
	}
	if vendorIDStr != "" {
		if id, err := uuid.Parse(vendorIDStr); err == nil {
			vendorID = &id
		} else {
			http.Error(w, "invalid vendorId", http.StatusBadRequest)
			return
		}
	}

	// parse limit/offset
	limit, offset := 0, 0
	if limitStr != "" {
		var l int
		_, _ = fmt.Sscanf(limitStr, "%d", &l)
		limit = l
	}
	if offsetStr != "" {
		var o int
		_, _ = fmt.Sscanf(offsetStr, "%d", &o)
		offset = o
	}

	list, err := h.service.ListAssessments(tenantID, vendorID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// ListEvidence returns evidence for an assessment
func (h *TPRMHandler) ListEvidence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assessmentIDStr := vars["assessmentId"]
	assessmentID, err := uuid.Parse(assessmentIDStr)
	if err != nil {
		http.Error(w, "invalid assessmentId", http.StatusBadRequest)
		return
	}
	list, err := h.service.ListEvidence(assessmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// GetEvidenceMeta returns evidence metadata and optional download URL
func (h *TPRMHandler) GetEvidenceMeta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	evidenceIDStr := vars["evidenceId"]
	evidenceID, err := uuid.Parse(evidenceIDStr)
	if err != nil {
		http.Error(w, "invalid evidenceId", http.StatusBadRequest)
		return
	}
	e, err := h.service.GetEvidenceByID(evidenceID)
	if err != nil {
		http.Error(w, "evidence not found", http.StatusNotFound)
		return
	}
	resp := map[string]interface{}{
		"evidence": e,
	}
	if url, err := h.service.PresignEvidenceURL(e, 15*time.Minute); err == nil {
		resp["downloadUrl"] = url
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DownloadEvidenceFile streams local file or redirects to S3
func (h *TPRMHandler) DownloadEvidenceFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	evidenceIDStr := vars["evidenceId"]
	evidenceID, err := uuid.Parse(evidenceIDStr)
	if err != nil {
		http.Error(w, "invalid evidenceId", http.StatusBadRequest)
		return
	}
	e, err := h.service.GetEvidenceByID(evidenceID)
	if err != nil {
		http.Error(w, "evidence not found", http.StatusNotFound)
		return
	}

	// Try presign first (S3), else read local file
	if url, err := h.service.PresignEvidenceURL(e, 15*time.Minute); err == nil {
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}
	// Local file path
	data, err := os.ReadFile(e.FilePath)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", e.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\"download\"")
	w.Write(data)
}

type AddFindingRequest struct {
	TenantID    string `json:"tenantId"`
	VendorID    string `json:"vendorId"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
}

func (h *TPRMHandler) AddFinding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assessmentIDStr := vars["assessmentId"]
	assessmentID, err := uuid.Parse(assessmentIDStr)
	if err != nil {
		http.Error(w, "invalid assessmentId", http.StatusBadRequest)
		return
	}

	var req AddFindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		http.Error(w, "invalid tenantId", http.StatusBadRequest)
		return
	}
	vendorID, err := uuid.Parse(req.VendorID)
	if err != nil {
		http.Error(w, "invalid vendorId", http.StatusBadRequest)
		return
	}

	f, err := h.service.AddFinding(tenantID, vendorID, assessmentID, req.Severity, req.Title, req.Description, req.Remediation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

func (h *TPRMHandler) ComputeRisk(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assessmentIDStr := vars["assessmentId"]
	assessmentID, err := uuid.Parse(assessmentIDStr)
	if err != nil {
		http.Error(w, "invalid assessmentId", http.StatusBadRequest)
		return
	}

	var req struct {
		VendorID string `json:"vendorId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	vendorID, err := uuid.Parse(req.VendorID)
	if err != nil {
		http.Error(w, "invalid vendorId", http.StatusBadRequest)
		return
	}

	score, level, err := h.service.ComputeRiskAndUpdate(assessmentID, vendorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"riskScore": score, "riskLevel": level})
}

func (h *TPRMHandler) GetDPDPAChecklist(w http.ResponseWriter, r *http.Request) {
	checklist := h.service.GetDPDPAChecklist()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checklist)
}

func (h *TPRMHandler) SubmitAuditResponse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assessmentID, err := uuid.Parse(vars["assessmentId"])
	if err != nil {
		http.Error(w, "invalid assessmentId", http.StatusBadRequest)
		return
	}

	var responses []models.AuditResponse
	if err := json.NewDecoder(r.Body).Decode(&responses); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.service.SubmitAuditResponse(assessmentID, responses); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "submitted"})
}

func (h *TPRMHandler) CreateDPATemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID string `json:"tenantId"`
		Name     string `json:"name"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	tenantID, _ := uuid.Parse(req.TenantID)

	t, err := h.service.CreateDPATemplate(tenantID, req.Name, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *TPRMHandler) GenerateDPA(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vendorID, _ := uuid.Parse(vars["vendorId"])

	var req struct {
		TenantID   string `json:"tenantId"`
		TemplateID string `json:"templateId"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	tenantID, _ := uuid.Parse(req.TenantID)
	templateID, _ := uuid.Parse(req.TemplateID)

	dpa, err := h.service.GenerateDPA(tenantID, vendorID, templateID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(dpa)
}

func (h *TPRMHandler) UploadSignedDPA(w http.ResponseWriter, r *http.Request) {
	// Placeholder for file upload logic similar to evidence upload
	w.WriteHeader(http.StatusNotImplemented)
}

