package services

import (
	"pixpivot/arc/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// Auto-migrate the schema
	db.AutoMigrate(&models.Purpose{}, &models.PurposeTemplate{}, &models.UserConsent{}, &models.ConsentFormPurpose{})

	return db
}

func TestValidateCompliance_Compliant(t *testing.T) {
	db := setupTestDB()
	service := NewPurposeService(db)

	// Create a test purpose
	purpose := &models.Purpose{
		ID:                  uuid.New(),
		Name:                "Test Purpose",
		Description:         "Test Description",
		LegalBasis:          "consent",
		RetentionPeriodDays: 365,
		TenantID:            uuid.New(),
		Active:              true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	err := db.Create(purpose).Error
	assert.NoError(t, err)

	// Validate compliance
	report, err := service.ValidateCompliance(purpose.ID)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "compliant", report.Status)
	assert.Empty(t, report.Issues)
}

func TestValidateCompliance_MissingFields(t *testing.T) {
	db := setupTestDB()
	service := NewPurposeService(db)

	// Create a test purpose with missing fields
	purpose := &models.Purpose{
		ID:       uuid.New(),
		Name:     "", // Missing name
		TenantID: uuid.New(),
		Active:   true,
	}

	err := db.Create(purpose).Error
	assert.NoError(t, err)

	// Validate compliance
	report, err := service.ValidateCompliance(purpose.ID)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "non_compliant", report.Status)
	assert.Contains(t, report.Issues, "Purpose name is required")
	assert.Contains(t, report.Issues, "Purpose description is required")
	assert.Contains(t, report.Issues, "Legal basis is required under DPDP Act")
}

func TestValidateCompliance_PurposeNotFound(t *testing.T) {
	db := setupTestDB()
	service := NewPurposeService(db)

	// Try to validate non-existent purpose
	_, err := service.ValidateCompliance(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "purpose not found")
}

func TestSuggestRetentionPeriod(t *testing.T) {
	db := setupTestDB()
	service := NewPurposeService(db)

	testCases := []struct {
		legalBasis   string
		expectedDays int
		description  string
	}{
		{"consent", 1095, "Consent-based purposes should suggest 3 years"},
		{"contract", 2555, "Contract-based purposes should suggest 7 years"},
		{"legal_obligation", 2555, "Legal obligation purposes should suggest 7 years"},
		{"legitimate_interest", 730, "Legitimate interest purposes should suggest 2 years"},
		{"unknown", 365, "Unknown legal basis should default to 1 year"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			purpose := &models.Purpose{
				LegalBasis: tc.legalBasis,
			}

			days := service.SuggestRetentionPeriod(purpose)
			assert.Equal(t, tc.expectedDays, days, tc.description)
		})
	}
}

func TestValidateLegalBasis(t *testing.T) {
	db := setupTestDB()
	service := NewPurposeService(db)

	validBases := []string{"consent", "contract", "legal_obligation", "legitimate_interest"}

	for _, basis := range validBases {
		t.Run("Valid basis: "+basis, func(t *testing.T) {
			purpose := &models.Purpose{LegalBasis: basis}
			err := service.ValidateLegalBasis(purpose)
			assert.NoError(t, err)
		})
	}

	t.Run("Invalid legal basis", func(t *testing.T) {
		purpose := &models.Purpose{LegalBasis: "invalid_basis"}
		err := service.ValidateLegalBasis(purpose)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid legal basis")
	})
}

func TestGetPurposeUsageStats(t *testing.T) {
	db := setupTestDB()
	service := NewPurposeService(db)

	// Create a test purpose
	purpose := &models.Purpose{
		ID:           uuid.New(),
		Name:         "Test Purpose",
		TenantID:     uuid.New(),
		TotalGranted: 10,
		TotalRevoked: 2,
		Active:       true,
	}

	err := db.Create(purpose).Error
	assert.NoError(t, err)

	// Create some test user consents
	activeConsent := &models.UserConsent{
		ID:        uuid.New(),
		PurposeID: purpose.ID,
		Status:    true,
		CreatedAt: time.Now(),
	}

	expiringConsent := &models.UserConsent{
		ID:        uuid.New(),
		PurposeID: purpose.ID,
		Status:    true,
		ExpiresAt: &[]time.Time{time.Now().Add(15 * 24 * time.Hour)}[0], // Expires in 15 days
		CreatedAt: time.Now(),
	}

	err = db.Create(activeConsent).Error
	assert.NoError(t, err)
	err = db.Create(expiringConsent).Error
	assert.NoError(t, err)

	// Get usage stats
	stats, err := service.GetPurposeUsageStats(purpose.ID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, purpose.ID, stats.PurposeID)
	assert.Equal(t, 10, stats.TotalGranted)
	assert.Equal(t, 2, stats.TotalRevoked)
	assert.Equal(t, 2, stats.ActiveConsents)   // Both consents are active
	assert.Equal(t, 1, stats.ExpiringConsents) // One expires within 30 days
}

func TestGetExpiringConsents(t *testing.T) {
	db := setupTestDB()
	service := NewPurposeService(db)

	// Create a test purpose
	purpose := &models.Purpose{
		ID:       uuid.New(),
		Name:     "Test Purpose",
		TenantID: uuid.New(),
		Active:   true,
	}

	err := db.Create(purpose).Error
	assert.NoError(t, err)

	// Create consents with different expiry dates
	now := time.Now()

	// Expires in 5 days (should be included)
	expiringConsent1 := &models.UserConsent{
		ID:        uuid.New(),
		PurposeID: purpose.ID,
		Status:    true,
		ExpiresAt: &[]time.Time{now.Add(5 * 24 * time.Hour)}[0],
		CreatedAt: now,
	}

	// Expires in 15 days (should be included for 30-day window)
	expiringConsent2 := &models.UserConsent{
		ID:        uuid.New(),
		PurposeID: purpose.ID,
		Status:    true,
		ExpiresAt: &[]time.Time{now.Add(15 * 24 * time.Hour)}[0],
		CreatedAt: now,
	}

	// Expires in 45 days (should not be included for 30-day window)
	futureConsent := &models.UserConsent{
		ID:        uuid.New(),
		PurposeID: purpose.ID,
		Status:    true,
		ExpiresAt: &[]time.Time{now.Add(45 * 24 * time.Hour)}[0],
		CreatedAt: now,
	}

	err = db.Create(expiringConsent1).Error
	assert.NoError(t, err)
	err = db.Create(expiringConsent2).Error
	assert.NoError(t, err)
	err = db.Create(futureConsent).Error
	assert.NoError(t, err)

	// Get expiring consents within 30 days
	expiring, err := service.GetExpiringConsents(purpose.ID, 30)
	assert.NoError(t, err)
	assert.Len(t, expiring, 2) // Should include first two consents

	// Get expiring consents within 10 days
	expiring, err = service.GetExpiringConsents(purpose.ID, 10)
	assert.NoError(t, err)
	assert.Len(t, expiring, 1) // Should include only first consent
}

