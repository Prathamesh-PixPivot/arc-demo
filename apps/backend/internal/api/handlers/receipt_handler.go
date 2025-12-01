package handlers

import (
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/core/services"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ReceiptHandler struct {
	receiptService *services.ReceiptService
}

type GenerateReceiptRequest struct {
	UserConsentID uuid.UUID `json:"userConsentId"`
}

type GenerateReceiptResponse struct {
	ReceiptID     uuid.UUID `json:"receiptId"`
	ReceiptNumber string    `json:"receiptNumber"`
	PDFPath       string    `json:"pdfPath"`
	Message       string    `json:"message"`
}

type BulkGenerateReceiptsRequest struct {
	ConsentIDs []uuid.UUID `json:"consentIds"`
}

type BulkGenerateReceiptsResponse struct {
	ReceiptIDs []uuid.UUID `json:"receiptIds"`
	Generated  int         `json:"generated"`
	Failed     int         `json:"failed"`
	Message    string      `json:"message"`
}

type BulkDownloadRequest struct {
	ReceiptIDs []uuid.UUID `json:"receiptIds"`
}

type EmailReceiptRequest struct {
	UserEmail string `json:"userEmail"`
}

type VerifyReceiptResponse struct {
	Valid      bool                   `json:"valid"`
	Receipt    *models.ConsentReceipt `json:"receipt,omitempty"`
	Message    string                 `json:"message"`
	VerifiedAt string                 `json:"verifiedAt"`
}

func NewReceiptHandler(receiptService *services.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: receiptService,
	}
}

// GenerateReceipt generates a new consent receipt
// @Summary Generate consent receipt
// @Description Generate a new DPDP-compliant consent receipt for a user consent
// @Tags receipts
// @Accept json
// @Produce json
// @Param consentId path string true "User Consent ID"
// @Success 200 {object} GenerateReceiptResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/user/consents/{consentId}/receipt [post]
func (h *ReceiptHandler) GenerateReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	consentIDStr := vars["consentId"]

	consentID, err := uuid.Parse(consentIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid consent ID format")
		return
	}

	receipt, err := h.receiptService.GenerateReceipt(consentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate receipt: "+err.Error())
		return
	}

	response := GenerateReceiptResponse{
		ReceiptID:     receipt.ID,
		ReceiptNumber: receipt.ReceiptNumber,
		PDFPath:       receipt.PDFPath,
		Message:       "Receipt generated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetReceipt retrieves receipt metadata
// @Summary Get receipt metadata
// @Description Get metadata for a specific receipt
// @Tags receipts
// @Produce json
// @Param receiptId path string true "Receipt ID"
// @Success 200 {object} models.ConsentReceipt
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/user/receipts/{receiptId} [get]
func (h *ReceiptHandler) GetReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	receiptIDStr := vars["receiptId"]

	receiptID, err := uuid.Parse(receiptIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid receipt ID format")
		return
	}

	receipt, err := h.receiptService.GetReceipt(receiptID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Receipt not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipt)
}

// DownloadReceipt downloads the PDF receipt
// @Summary Download receipt PDF
// @Description Download the PDF file for a specific receipt
// @Tags receipts
// @Produce application/pdf
// @Param receiptId path string true "Receipt ID"
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/user/receipts/{receiptId}/download [get]
func (h *ReceiptHandler) DownloadReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	receiptIDStr := vars["receiptId"]

	receiptID, err := uuid.Parse(receiptIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid receipt ID format")
		return
	}

	receipt, err := h.receiptService.GetReceipt(receiptID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Receipt not found")
		return
	}

	// Read PDF file - using internal method via service
	// Note: This should be exposed as a public method in the service
	pdfData, err := os.ReadFile(receipt.PDFPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to read PDF file")
		return
	}

	// Increment download count through repository
	// Note: This is handled internally by the service

	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+receipt.ReceiptNumber+".pdf\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfData)))

	w.Write(pdfData)
}

// EmailReceipt sends receipt via email
// @Summary Email receipt
// @Description Send receipt PDF via email
// @Tags receipts
// @Accept json
// @Produce json
// @Param receiptId path string true "Receipt ID"
// @Param request body EmailReceiptRequest true "Email request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/user/receipts/{receiptId}/email [post]
func (h *ReceiptHandler) EmailReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	receiptIDStr := vars["receiptId"]

	receiptID, err := uuid.Parse(receiptIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid receipt ID format")
		return
	}

	var req EmailReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserEmail == "" {
		writeError(w, http.StatusBadRequest, "User email is required")
		return
	}

	err = h.receiptService.EmailReceipt(receiptID, req.UserEmail)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to send email: "+err.Error())
		return
	}

	response := SuccessResponse{
		Message: "Receipt emailed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyReceipt verifies a receipt by receipt number (public endpoint)
// @Summary Verify receipt
// @Description Verify the authenticity of a receipt using receipt number
// @Tags receipts
// @Produce json
// @Param receiptNumber path string true "Receipt Number"
// @Success 200 {object} VerifyReceiptResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/public/receipts/verify/{receiptNumber} [get]
func (h *ReceiptHandler) VerifyReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	receiptNumber := vars["receiptNumber"]

	if receiptNumber == "" {
		writeError(w, http.StatusBadRequest, "Receipt number is required")
		return
	}

	receipt, err := h.receiptService.VerifyReceipt(receiptNumber)
	if err != nil {
		response := VerifyReceiptResponse{
			Valid:      false,
			Message:    err.Error(),
			VerifiedAt: "",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Return 200 even for invalid receipts
		json.NewEncoder(w).Encode(response)
		return
	}

	response := VerifyReceiptResponse{
		Valid:      true,
		Receipt:    receipt,
		Message:    "Receipt is valid and authentic",
		VerifiedAt: receipt.GeneratedAt.Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BulkGenerateReceipts generates receipts for multiple consents
// @Summary Bulk generate receipts
// @Description Generate receipts for multiple user consents
// @Tags receipts
// @Accept json
// @Produce json
// @Param request body BulkGenerateReceiptsRequest true "Bulk generation request"
// @Success 200 {object} BulkGenerateReceiptsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/fiduciary/receipts/bulk-generate [post]
func (h *ReceiptHandler) BulkGenerateReceipts(w http.ResponseWriter, r *http.Request) {
	var req BulkGenerateReceiptsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.ConsentIDs) == 0 {
		writeError(w, http.StatusBadRequest, "At least one consent ID is required")
		return
	}

	receiptIDs, err := h.receiptService.BulkGenerateReceipts(req.ConsentIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate receipts: "+err.Error())
		return
	}

	generated := len(receiptIDs)
	failed := len(req.ConsentIDs) - generated

	response := BulkGenerateReceiptsResponse{
		ReceiptIDs: receiptIDs,
		Generated:  generated,
		Failed:     failed,
		Message:    "Bulk receipt generation completed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BulkDownloadReceipts downloads multiple receipts as ZIP
// @Summary Bulk download receipts
// @Description Download multiple receipts as a ZIP file
// @Tags receipts
// @Accept json
// @Produce application/zip
// @Param request body BulkDownloadRequest true "Bulk download request"
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/fiduciary/receipts/bulk-download [post]
func (h *ReceiptHandler) BulkDownloadReceipts(w http.ResponseWriter, r *http.Request) {
	var req BulkDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.ReceiptIDs) == 0 {
		writeError(w, http.StatusBadRequest, "At least one receipt ID is required")
		return
	}

	zipData, err := h.receiptService.DownloadReceiptsAsZip(req.ReceiptIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create ZIP file: "+err.Error())
		return
	}

	// Set headers for ZIP download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"receipts.zip\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(zipData)))

	w.Write(zipData)
}

// Helper method to read PDF file (expose from service)
func (h *ReceiptHandler) ReadPDFFile(filePath string) ([]byte, error) {
	// This is a wrapper to expose the service method if needed
	return nil, nil // Implementation would depend on service structure
}

// Helper method to increment download count
func (h *ReceiptHandler) IncrementDownloadCount(receiptID uuid.UUID) error {
	// This is a wrapper to expose the service method if needed
	return nil // Implementation would depend on service structure
}

// Common response types
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

