package services

import (
	"archive/zip"
	"bytes"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
	"gorm.io/datatypes"
)

type ReceiptService struct {
	receiptRepo     *repository.ReceiptRepository
	userConsentRepo *repository.UserConsentRepository
	purposeRepo     *repository.PurposeRepository
	emailService    *EmailService
	s3Client        *s3.S3
	baseURL         string
	storageType     string // "local" or "s3"
	storagePath     string
	s3Bucket        string
	s3Endpoint      string
	s3Region        string
	s3AccessKey     string
	s3SecretKey     string
	s3UseSSL        bool
	s3ForcePath     bool
}

type ReceiptData struct {
	ReceiptNumber    string
	UserID           uuid.UUID
	UserEmail        string
	UserName         string
	ConsentFormTitle string
	ConsentFormID    uuid.UUID
	Purposes         []PurposeData
	DataObjects      []DataObjectData
	GeneratedAt      time.Time
	OrganizationName string
	DPOContact       string
	TenantID         uuid.UUID
}

type PurposeData struct {
	Name        string
	Description string
	LegalBasis  string
	Consented   bool
	ExpiresAt   *time.Time
}

type DataObjectData struct {
	Name            string
	RetentionPeriod string
	Purpose         string
}

func NewReceiptService(
	receiptRepo *repository.ReceiptRepository,
	userConsentRepo *repository.UserConsentRepository,
	purposeRepo *repository.PurposeRepository,
	emailService *EmailService,
	baseURL string,
	storageType string,
	storagePath string,
	s3Bucket string,
	s3Endpoint string,
	s3AccessKey string,
	s3SecretKey string,
	s3Region string,
	useSSL bool,
	forcePathStyle bool,
) *ReceiptService {
	service := &ReceiptService{
		receiptRepo:     receiptRepo,
		userConsentRepo: userConsentRepo,
		purposeRepo:     purposeRepo,
		emailService:    emailService,
		baseURL:         baseURL,
		storageType:     storageType,
		storagePath:     storagePath,
		s3Bucket:        s3Bucket,
		s3Endpoint:      s3Endpoint,
		s3AccessKey:     s3AccessKey,
		s3SecretKey:     s3SecretKey,
		s3Region:        s3Region,
		s3UseSSL:        useSSL,
		s3ForcePath:     forcePathStyle,
	}

	// Initialize S3 client if using S3 storage
	if storageType == "s3" {
		sess, err := session.NewSession(&aws.Config{
			Region:           aws.String(service.s3Region),
			Endpoint:         aws.String(service.s3Endpoint),
			S3ForcePathStyle: aws.Bool(service.s3ForcePath),
			DisableSSL:       aws.Bool(!service.s3UseSSL),
			Credentials:      credentials.NewStaticCredentials(service.s3AccessKey, service.s3SecretKey, ""),
		})
		if err == nil {
			service.s3Client = s3.New(sess)
		}
	}

	return service
}

// GenerateReceipt creates a new consent receipt with PDF and QR code
func (s *ReceiptService) GenerateReceipt(userConsentID uuid.UUID) (*models.ConsentReceipt, error) {
	// Get user consent details
	userConsent, err := s.userConsentRepo.GetUserConsentByID(userConsentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user consent: %w", err)
	}

	// Generate receipt number
	receiptNumber := s.GenerateReceiptNumber()

	// Create receipt record
	receipt := &models.ConsentReceipt{
		ID:            uuid.New(),
		UserConsentID: userConsentID,
		TenantID:      userConsent.TenantID,
		ReceiptNumber: receiptNumber,
		GeneratedAt:   time.Now(),
		IsValid:       true,
		DownloadCount: 0,
		Metadata:      datatypes.JSON("{}"),
	}

	// Generate QR code data
	qrData, err := s.GenerateQRCode(receipt.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	receipt.QRCodeData = string(qrData)

	// Collect receipt data
	receiptData, err := s.collectReceiptData(userConsent, receiptNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to collect receipt data: %w", err)
	}

	// Generate PDF
	pdfPath, err := s.generatePDF(receiptData, receipt.ID, qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	receipt.PDFPath = pdfPath

	// Save receipt to database
	err = s.receiptRepo.CreateReceipt(receipt)
	if err != nil {
		return nil, fmt.Errorf("failed to save receipt: %w", err)
	}

	return receipt, nil
}

// GenerateReceiptNumber creates a unique receipt number
func (s *ReceiptService) GenerateReceiptNumber() string {
	now := time.Now()
	dateStr := now.Format("20060102")

	// Generate random 6-character suffix
	randomBytes := make([]byte, 3)
	rand.Read(randomBytes)
	randomStr := strings.ToUpper(hex.EncodeToString(randomBytes))

	return fmt.Sprintf("RCP-%s-%s", dateStr, randomStr)
}

// GenerateQRCode creates QR code data for receipt verification
func (s *ReceiptService) GenerateQRCode(receiptID uuid.UUID) ([]byte, error) {
	verificationURL := fmt.Sprintf("%s/verify/%s", s.baseURL, receiptID.String())

	qrCode, err := qrcode.Encode(verificationURL, qrcode.High, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return qrCode, nil
}

// VerifyReceipt validates a receipt by receipt number
func (s *ReceiptService) VerifyReceipt(receiptNumber string) (*models.ConsentReceipt, error) {
	receipt, err := s.receiptRepo.GetReceiptByNumber(receiptNumber)
	if err != nil {
		return nil, fmt.Errorf("receipt not found: %w", err)
	}

	if !receipt.IsValid {
		return nil, fmt.Errorf("receipt is invalid")
	}

	if receipt.ExpiresAt != nil && receipt.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("receipt has expired")
	}

	return receipt, nil
}

// EmailReceipt sends the receipt via email
func (s *ReceiptService) EmailReceipt(receiptID uuid.UUID, userEmail string) error {
	receipt, err := s.receiptRepo.GetReceiptByID(receiptID)
	if err != nil {
		return fmt.Errorf("failed to get receipt: %w", err)
	}

	// Read PDF file
	pdfData, err := s.readPDFFile(receipt.PDFPath)
	if err != nil {
		return fmt.Errorf("failed to read PDF file: %w", err)
	}

	// Send email with PDF attachment
	subject := fmt.Sprintf("Your Consent Receipt - %s", receipt.ReceiptNumber)
	body := s.generateEmailBody(receipt)

	err = s.emailService.SendEmailWithAttachment(userEmail, subject, body, pdfData, fmt.Sprintf("%s.pdf", receipt.ReceiptNumber))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Update emailed timestamp
	now := time.Now()
	receipt.EmailedAt = &now
	err = s.receiptRepo.UpdateReceipt(receipt)
	if err != nil {
		return fmt.Errorf("failed to update receipt: %w", err)
	}

	return nil
}

// GetReceipt retrieves a receipt by ID
func (s *ReceiptService) GetReceipt(receiptID uuid.UUID) (*models.ConsentReceipt, error) {
	return s.receiptRepo.GetReceiptByID(receiptID)
}

// BulkGenerateReceipts generates receipts for multiple consent IDs
func (s *ReceiptService) BulkGenerateReceipts(consentIDs []uuid.UUID) ([]uuid.UUID, error) {
	var receiptIDs []uuid.UUID

	for _, consentID := range consentIDs {
		receipt, err := s.GenerateReceipt(consentID)
		if err != nil {
			// Log error but continue with other receipts
			continue
		}
		receiptIDs = append(receiptIDs, receipt.ID)
	}

	return receiptIDs, nil
}

// DownloadReceiptsAsZip creates a ZIP file containing multiple receipts
func (s *ReceiptService) DownloadReceiptsAsZip(receiptIDs []uuid.UUID) ([]byte, error) {
	receipts, err := s.receiptRepo.GetReceiptsByIDs(receiptIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipts: %w", err)
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for _, receipt := range receipts {
		// Read PDF file
		pdfData, err := s.readPDFFile(receipt.PDFPath)
		if err != nil {
			continue // Skip this file if error
		}

		// Add file to ZIP
		fileName := fmt.Sprintf("%s.pdf", receipt.ReceiptNumber)
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			continue
		}

		_, err = fileWriter.Write(pdfData)
		if err != nil {
			continue
		}

		// Increment download count
		s.receiptRepo.IncrementDownloadCount(receipt.ID)
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	return buf.Bytes(), nil
}

// collectReceiptData gathers all necessary data for receipt generation
func (s *ReceiptService) collectReceiptData(userConsent *models.UserConsent, receiptNumber string) (*ReceiptData, error) {
	// Get purpose details
	purpose, err := s.purposeRepo.GetPurposeByID(userConsent.PurposeID, userConsent.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purpose: %w", err)
	}

	// Build receipt data
	receiptData := &ReceiptData{
		ReceiptNumber:    receiptNumber,
		UserID:           userConsent.UserID,
		GeneratedAt:      time.Now(),
		OrganizationName: "Your Organization", // TODO: Get from tenant/organization
		DPOContact:       "dpo@yourorg.com",   // TODO: Get from tenant settings
		TenantID:         userConsent.TenantID,
		Purposes: []PurposeData{
			{
				Name:        purpose.Name,
				Description: purpose.Description,
				LegalBasis:  purpose.LegalBasis,
				Consented:   userConsent.Status,
				ExpiresAt:   userConsent.ExpiresAt,
			},
		},
		// TODO: Add data objects collection
		DataObjects: []DataObjectData{},
	}

	return receiptData, nil
}

// generatePDF creates a DPDP-compliant PDF receipt
func (s *ReceiptService) generatePDF(data *ReceiptData, receiptID uuid.UUID, qrCode []byte) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "", 12)

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "CONSENT RECEIPT")
	pdf.Ln(15)

	// Add watermark (simplified - gofpdf doesn't have RotatedText)
	// TODO: Implement rotated watermark using TransformBegin/TransformEnd if needed

	// Receipt details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, fmt.Sprintf("Receipt Number: %s", data.ReceiptNumber))
	pdf.Ln(8)
	pdf.Cell(0, 8, fmt.Sprintf("Generated: %s", data.GeneratedAt.Format("2006-01-02 15:04:05")))
	pdf.Ln(15)

	// Organization details
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Data Fiduciary Information")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Organization: %s", data.OrganizationName))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("DPO Contact: %s", data.DPOContact))
	pdf.Ln(15)

	// Data Subject details
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Data Principal Information")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("User ID: %s", data.UserID.String()))
	pdf.Ln(15)

	// Consent details
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Consent Details")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 11)

	for _, purpose := range data.Purposes {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 6, fmt.Sprintf("Purpose: %s", purpose.Name))
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 5, fmt.Sprintf("Description: %s", purpose.Description), "", "", false)
		pdf.Ln(2)
		pdf.Cell(0, 5, fmt.Sprintf("Legal Basis: %s", purpose.LegalBasis))
		pdf.Ln(5)
		consentStatus := "Granted"
		if !purpose.Consented {
			consentStatus = "Withdrawn"
		}
		pdf.Cell(0, 5, fmt.Sprintf("Status: %s", consentStatus))
		pdf.Ln(5)
		if purpose.ExpiresAt != nil {
			pdf.Cell(0, 5, fmt.Sprintf("Expires: %s", purpose.ExpiresAt.Format("2006-01-02")))
			pdf.Ln(5)
		}
		pdf.Ln(5)
	}

	// Rights information
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Your Rights Under DPDP Act 2023")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	rights := []string{
		"• Right to withdraw consent at any time",
		"• Right to access your personal data",
		"• Right to correction of inaccurate data",
		"• Right to erasure of personal data",
		"• Right to data portability",
		"• Right to grievance redressal",
	}
	for _, right := range rights {
		pdf.Cell(0, 5, right)
		pdf.Ln(5)
	}
	pdf.Ln(10)

	// QR Code
	if len(qrCode) > 0 {
		// Save QR code temporarily
		qrPath := filepath.Join(os.TempDir(), fmt.Sprintf("qr_%s.png", receiptID.String()))
		err := os.WriteFile(qrPath, qrCode, 0644)
		if err == nil {
			pdf.Image(qrPath, 150, pdf.GetY(), 30, 30, false, "", 0, "")
			os.Remove(qrPath) // Clean up
		}
	}

	pdf.SetY(pdf.GetY() + 35)
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 5, "Scan QR code to verify this receipt")
	pdf.Ln(10)

	// Footer
	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, "This is a digitally generated consent receipt as per DPDP Act 2023")
	pdf.Ln(5)
	pdf.Cell(0, 5, fmt.Sprintf("Generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Save PDF
	fileName := fmt.Sprintf("%s.pdf", data.ReceiptNumber)
	filePath := s.getStoragePath(data.TenantID, fileName)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	os.MkdirAll(dir, 0755)

	err := pdf.OutputFileAndClose(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	// Upload to S3 if configured
	if s.storageType == "s3" && s.s3Client != nil {
		s3Key := fmt.Sprintf("receipts/%s/%s", data.TenantID.String(), fileName)
		err = s.uploadToS3(filePath, s3Key)
		if err != nil {
			return "", fmt.Errorf("failed to upload to S3: %w", err)
		}
		return s3Key, nil
	}

	return filePath, nil
}

// getStoragePath returns the storage path for a file
func (s *ReceiptService) getStoragePath(tenantID uuid.UUID, fileName string) string {
	year := strconv.Itoa(time.Now().Year())
	month := fmt.Sprintf("%02d", int(time.Now().Month()))

	return filepath.Join(s.storagePath, "receipts", tenantID.String(), year, month, fileName)
}

// uploadToS3 uploads a file to S3
func (s *ReceiptService) uploadToS3(filePath, s3Key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.s3Bucket),
		Key:    aws.String(s3Key),
		Body:   file,
	})

	return err
}

// readPDFFile reads a PDF file from storage
func (s *ReceiptService) readPDFFile(filePath string) ([]byte, error) {
	if s.storageType == "s3" && s.s3Client != nil {
		// Download from S3
		result, err := s.s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(s.s3Bucket),
			Key:    aws.String(filePath),
		})
		if err != nil {
			return nil, err
		}
		defer result.Body.Close()

		return io.ReadAll(result.Body)
	}

	// Read from local storage
	return os.ReadFile(filePath)
}

// generateEmailBody creates the email body for receipt delivery
func (s *ReceiptService) generateEmailBody(receipt *models.ConsentReceipt) string {
	return fmt.Sprintf(`
Dear User,

Thank you for providing your consent. Please find attached your official consent receipt.

Receipt Number: %s
Generated: %s

This receipt serves as proof of your consent and can be used for verification purposes.

You can verify this receipt online at: %s/verify/%s

If you have any questions or wish to withdraw your consent, please contact our Data Protection Officer.

Best regards,
Data Protection Team
`, receipt.ReceiptNumber, receipt.GeneratedAt.Format("2006-01-02 15:04:05"), s.baseURL, receipt.ReceiptNumber)
}

// CleanupExpiredReceipts removes expired receipts
func (s *ReceiptService) CleanupExpiredReceipts() error {
	return s.receiptRepo.DeleteExpiredReceipts()
}

