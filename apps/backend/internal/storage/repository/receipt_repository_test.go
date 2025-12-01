package repository

import (
	"pixpivot/arc/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestReceiptDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate the schema
	db.AutoMigrate(&models.ConsentReceipt{}, &models.UserConsent{}, &models.Purpose{})

	return db
}

func TestReceiptRepository_CreateReceipt(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	receipt := &models.ConsentReceipt{
		ID:             uuid.New(),
		UserConsentID:  uuid.New(),
		TenantID:       uuid.New(),
		ReceiptNumber:  "RCP-20241014-ABCDEF",
		PDFPath:        "/path/to/receipt.pdf",
		QRCodeData:     "https://verify.com/receipt",
		GeneratedAt:    time.Now(),
		DownloadCount:  0,
		IsValid:        true,
		Metadata:       datatypes.JSON("{}"),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := repo.CreateReceipt(receipt)
	assert.NoError(t, err)

	// Verify the receipt was created
	var count int64
	db.Model(&models.ConsentReceipt{}).Where("id = ?", receipt.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestReceiptRepository_GetReceiptByID(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	receiptID := uuid.New()
	receipt := &models.ConsentReceipt{
		ID:             receiptID,
		UserConsentID:  uuid.New(),
		TenantID:       uuid.New(),
		ReceiptNumber:  "RCP-20241014-ABCDEF",
		PDFPath:        "/path/to/receipt.pdf",
		QRCodeData:     "https://verify.com/receipt",
		GeneratedAt:    time.Now(),
		DownloadCount:  0,
		IsValid:        true,
		Metadata:       datatypes.JSON("{}"),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	db.Create(receipt)

	retrievedReceipt, err := repo.GetReceiptByID(receiptID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedReceipt)
	assert.Equal(t, receiptID, retrievedReceipt.ID)
	assert.Equal(t, receipt.ReceiptNumber, retrievedReceipt.ReceiptNumber)
}

func TestReceiptRepository_GetReceiptByID_NotFound(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	nonExistentID := uuid.New()
	retrievedReceipt, err := repo.GetReceiptByID(nonExistentID)

	assert.Error(t, err)
	assert.Nil(t, retrievedReceipt)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestReceiptRepository_GetReceiptByNumber(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	receiptNumber := "RCP-20241014-ABCDEF"
	receipt := &models.ConsentReceipt{
		ID:             uuid.New(),
		UserConsentID:  uuid.New(),
		TenantID:       uuid.New(),
		ReceiptNumber:  receiptNumber,
		PDFPath:        "/path/to/receipt.pdf",
		QRCodeData:     "https://verify.com/receipt",
		GeneratedAt:    time.Now(),
		DownloadCount:  0,
		IsValid:        true,
		Metadata:       datatypes.JSON("{}"),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	db.Create(receipt)

	retrievedReceipt, err := repo.GetReceiptByNumber(receiptNumber)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedReceipt)
	assert.Equal(t, receiptNumber, retrievedReceipt.ReceiptNumber)
	assert.Equal(t, receipt.ID, retrievedReceipt.ID)
}

func TestReceiptRepository_GetReceiptsByUserConsent(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	userConsentID := uuid.New()
	tenantID := uuid.New()

	// Create multiple receipts for the same user consent
	receipts := []*models.ConsentReceipt{
		{
			ID:             uuid.New(),
			UserConsentID:  userConsentID,
			TenantID:       tenantID,
			ReceiptNumber:  "RCP-20241014-ABCDEF",
			PDFPath:        "/path/to/receipt1.pdf",
			GeneratedAt:    time.Now().Add(-2 * time.Hour),
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now().Add(-2 * time.Hour),
			UpdatedAt:      time.Now().Add(-2 * time.Hour),
		},
		{
			ID:             uuid.New(),
			UserConsentID:  userConsentID,
			TenantID:       tenantID,
			ReceiptNumber:  "RCP-20241014-GHIJKL",
			PDFPath:        "/path/to/receipt2.pdf",
			GeneratedAt:    time.Now().Add(-1 * time.Hour),
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now().Add(-1 * time.Hour),
			UpdatedAt:      time.Now().Add(-1 * time.Hour),
		},
	}

	for _, receipt := range receipts {
		db.Create(receipt)
	}

	retrievedReceipts, err := repo.GetReceiptsByUserConsent(userConsentID)
	assert.NoError(t, err)
	assert.Len(t, retrievedReceipts, 2)
	
	// Should be ordered by created_at DESC (newest first)
	assert.Equal(t, "RCP-20241014-GHIJKL", retrievedReceipts[0].ReceiptNumber)
	assert.Equal(t, "RCP-20241014-ABCDEF", retrievedReceipts[1].ReceiptNumber)
}

func TestReceiptRepository_GetReceiptsByTenant(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	tenantID := uuid.New()
	otherTenantID := uuid.New()

	// Create receipts for different tenants
	receipts := []*models.ConsentReceipt{
		{
			ID:             uuid.New(),
			UserConsentID:  uuid.New(),
			TenantID:       tenantID,
			ReceiptNumber:  "RCP-20241014-TENANT1",
			GeneratedAt:    time.Now(),
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			UserConsentID:  uuid.New(),
			TenantID:       tenantID,
			ReceiptNumber:  "RCP-20241014-TENANT2",
			GeneratedAt:    time.Now(),
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			UserConsentID:  uuid.New(),
			TenantID:       otherTenantID,
			ReceiptNumber:  "RCP-20241014-OTHER",
			GeneratedAt:    time.Now(),
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for _, receipt := range receipts {
		db.Create(receipt)
	}

	retrievedReceipts, err := repo.GetReceiptsByTenant(tenantID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, retrievedReceipts, 2)

	// Verify all receipts belong to the correct tenant
	for _, receipt := range retrievedReceipts {
		assert.Equal(t, tenantID, receipt.TenantID)
	}
}

func TestReceiptRepository_UpdateReceipt(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	receipt := &models.ConsentReceipt{
		ID:             uuid.New(),
		UserConsentID:  uuid.New(),
		TenantID:       uuid.New(),
		ReceiptNumber:  "RCP-20241014-ABCDEF",
		PDFPath:        "/path/to/receipt.pdf",
		GeneratedAt:    time.Now(),
		DownloadCount:  0,
		IsValid:        true,
		Metadata:       datatypes.JSON("{}"),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	db.Create(receipt)

	// Update the receipt
	emailedAt := time.Now()
	receipt.EmailedAt = &emailedAt
	receipt.DownloadCount = 5

	err := repo.UpdateReceipt(receipt)
	assert.NoError(t, err)

	// Verify the update
	var updatedReceipt models.ConsentReceipt
	db.First(&updatedReceipt, receipt.ID)
	assert.NotNil(t, updatedReceipt.EmailedAt)
	assert.Equal(t, 5, updatedReceipt.DownloadCount)
}

func TestReceiptRepository_IncrementDownloadCount(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	receipt := &models.ConsentReceipt{
		ID:             uuid.New(),
		UserConsentID:  uuid.New(),
		TenantID:       uuid.New(),
		ReceiptNumber:  "RCP-20241014-ABCDEF",
		PDFPath:        "/path/to/receipt.pdf",
		GeneratedAt:    time.Now(),
		DownloadCount:  3,
		IsValid:        true,
		Metadata:       datatypes.JSON("{}"),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	db.Create(receipt)

	err := repo.IncrementDownloadCount(receipt.ID)
	assert.NoError(t, err)

	// Verify the download count was incremented
	var updatedReceipt models.ConsentReceipt
	db.First(&updatedReceipt, receipt.ID)
	assert.Equal(t, 4, updatedReceipt.DownloadCount)
}

func TestReceiptRepository_GetReceiptsByIDs(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	receiptIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	
	// Create receipts
	for i, receiptID := range receiptIDs {
		receipt := &models.ConsentReceipt{
			ID:             receiptID,
			UserConsentID:  uuid.New(),
			TenantID:       uuid.New(),
			ReceiptNumber:  "RCP-20241014-" + string(rune('A'+i)),
			GeneratedAt:    time.Now(),
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		db.Create(receipt)
	}

	// Get receipts by IDs
	retrievedReceipts, err := repo.GetReceiptsByIDs(receiptIDs[:2]) // Get first 2
	assert.NoError(t, err)
	assert.Len(t, retrievedReceipts, 2)

	// Verify the correct receipts were retrieved
	retrievedIDs := make([]uuid.UUID, len(retrievedReceipts))
	for i, receipt := range retrievedReceipts {
		retrievedIDs[i] = receipt.ID
	}
	assert.Contains(t, retrievedIDs, receiptIDs[0])
	assert.Contains(t, retrievedIDs, receiptIDs[1])
}

func TestReceiptRepository_DeleteExpiredReceipts(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	now := time.Now()
	expiredTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	// Create receipts with different expiration times
	receipts := []*models.ConsentReceipt{
		{
			ID:             uuid.New(),
			UserConsentID:  uuid.New(),
			TenantID:       uuid.New(),
			ReceiptNumber:  "RCP-20241014-EXPIRED",
			GeneratedAt:    time.Now(),
			ExpiresAt:      &expiredTime, // Expired
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			UserConsentID:  uuid.New(),
			TenantID:       uuid.New(),
			ReceiptNumber:  "RCP-20241014-VALID",
			GeneratedAt:    time.Now(),
			ExpiresAt:      &futureTime, // Not expired
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			UserConsentID:  uuid.New(),
			TenantID:       uuid.New(),
			ReceiptNumber:  "RCP-20241014-NOEXP",
			GeneratedAt:    time.Now(),
			ExpiresAt:      nil, // No expiration
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for _, receipt := range receipts {
		db.Create(receipt)
	}

	// Delete expired receipts
	err := repo.DeleteExpiredReceipts()
	assert.NoError(t, err)

	// Verify only the expired receipt was deleted
	var count int64
	db.Model(&models.ConsentReceipt{}).Count(&count)
	assert.Equal(t, int64(2), count) // Only 2 should remain

	// Verify the expired receipt is gone
	var expiredReceipt models.ConsentReceipt
	err = db.Where("receipt_number = ?", "RCP-20241014-EXPIRED").First(&expiredReceipt).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestReceiptRepository_GetReceiptCount(t *testing.T) {
	db := setupTestReceiptDB()
	repo := NewReceiptRepository(db)

	tenantID := uuid.New()
	otherTenantID := uuid.New()

	// Create receipts for different tenants
	receipts := []*models.ConsentReceipt{
		{
			ID:             uuid.New(),
			UserConsentID:  uuid.New(),
			TenantID:       tenantID,
			ReceiptNumber:  "RCP-20241014-001",
			GeneratedAt:    time.Now(),
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			UserConsentID:  uuid.New(),
			TenantID:       tenantID,
			ReceiptNumber:  "RCP-20241014-002",
			GeneratedAt:    time.Now(),
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			UserConsentID:  uuid.New(),
			TenantID:       otherTenantID,
			ReceiptNumber:  "RCP-20241014-003",
			GeneratedAt:    time.Now(),
			DownloadCount:  0,
			IsValid:        true,
			Metadata:       datatypes.JSON("{}"),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for _, receipt := range receipts {
		db.Create(receipt)
	}

	count, err := repo.GetReceiptCount(tenantID)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	otherCount, err := repo.GetReceiptCount(otherTenantID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), otherCount)
}

