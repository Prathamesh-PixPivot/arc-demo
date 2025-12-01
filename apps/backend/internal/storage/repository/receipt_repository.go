package repository

import (
	"pixpivot/arc/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReceiptRepository struct {
	db *gorm.DB
}

func NewReceiptRepository(db *gorm.DB) *ReceiptRepository {
	return &ReceiptRepository{db: db}
}

// CreateReceipt creates a new consent receipt
func (r *ReceiptRepository) CreateReceipt(receipt *models.ConsentReceipt) error {
	return r.db.Create(receipt).Error
}

// GetReceiptByID retrieves a receipt by its ID
func (r *ReceiptRepository) GetReceiptByID(receiptID uuid.UUID) (*models.ConsentReceipt, error) {
	var receipt models.ConsentReceipt
	err := r.db.Where("id = ?", receiptID).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetReceiptByNumber retrieves a receipt by its receipt number
func (r *ReceiptRepository) GetReceiptByNumber(receiptNumber string) (*models.ConsentReceipt, error) {
	var receipt models.ConsentReceipt
	err := r.db.Where("receipt_number = ?", receiptNumber).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetReceiptsByUserConsent retrieves all receipts for a user consent
func (r *ReceiptRepository) GetReceiptsByUserConsent(userConsentID uuid.UUID) ([]models.ConsentReceipt, error) {
	var receipts []models.ConsentReceipt
	err := r.db.Where("user_consent_id = ?", userConsentID).Order("created_at DESC").Find(&receipts).Error
	return receipts, err
}

// GetReceiptsByTenant retrieves all receipts for a tenant with pagination
func (r *ReceiptRepository) GetReceiptsByTenant(tenantID uuid.UUID, limit, offset int) ([]models.ConsentReceipt, error) {
	var receipts []models.ConsentReceipt
	err := r.db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&receipts).Error
	return receipts, err
}

// UpdateReceipt updates an existing receipt
func (r *ReceiptRepository) UpdateReceipt(receipt *models.ConsentReceipt) error {
	return r.db.Save(receipt).Error
}

// IncrementDownloadCount increments the download count for a receipt
func (r *ReceiptRepository) IncrementDownloadCount(receiptID uuid.UUID) error {
	return r.db.Model(&models.ConsentReceipt{}).
		Where("id = ?", receiptID).
		Update("download_count", gorm.Expr("download_count + 1")).Error
}

// GetReceiptsByIDs retrieves multiple receipts by their IDs
func (r *ReceiptRepository) GetReceiptsByIDs(receiptIDs []uuid.UUID) ([]models.ConsentReceipt, error) {
	var receipts []models.ConsentReceipt
	err := r.db.Where("id IN ?", receiptIDs).Find(&receipts).Error
	return receipts, err
}

// DeleteExpiredReceipts deletes receipts that have expired
func (r *ReceiptRepository) DeleteExpiredReceipts() error {
	return r.db.Where("expires_at < NOW() AND expires_at IS NOT NULL").
		Delete(&models.ConsentReceipt{}).Error
}

// GetReceiptCount returns the total count of receipts for a tenant
func (r *ReceiptRepository) GetReceiptCount(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ConsentReceipt{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}

