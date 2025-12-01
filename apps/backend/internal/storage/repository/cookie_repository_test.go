package repository

import (
	"pixpivot/arc/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCookieRepositoryTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// Auto-migrate the schema
	db.AutoMigrate(&models.Cookie{}, &models.CookieScan{})

	return db
}

func TestCookieRepository_GetByID_Success(t *testing.T) {
	db := setupCookieRepositoryTestDB()
	repo := NewCookieRepository(db)

	tenantID := uuid.New()
	cookieID := uuid.New()

	cookie := &models.Cookie{
		ID:       cookieID,
		Name:     "Test Cookie",
		TenantID: tenantID,
		Domain:   "example.com",
	}

	err := db.Create(cookie).Error
	assert.NoError(t, err)

	// Test getting cookie
	found, err := repo.GetByID(cookieID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, cookieID, found.ID)
	assert.Equal(t, tenantID, found.TenantID)
}

func TestCookieRepository_GetByID_WrongTenant(t *testing.T) {
	db := setupCookieRepositoryTestDB()
	repo := NewCookieRepository(db)

	tenantID := uuid.New()
	otherTenantID := uuid.New()
	cookieID := uuid.New()

	cookie := &models.Cookie{
		ID:       cookieID,
		Name:     "Test Cookie",
		TenantID: tenantID,
		Domain:   "example.com",
	}

	err := db.Create(cookie).Error
	assert.NoError(t, err)

	// Test getting cookie with wrong tenant
	_, err = repo.GetByID(cookieID, otherTenantID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestCookieRepository_ListByTenant(t *testing.T) {
	db := setupCookieRepositoryTestDB()
	repo := NewCookieRepository(db)

	tenantID := uuid.New()
	otherTenantID := uuid.New()

	c1 := &models.Cookie{ID: uuid.New(), Name: "C1", TenantID: tenantID, Domain: "example.com"}
	c2 := &models.Cookie{ID: uuid.New(), Name: "C2", TenantID: tenantID, Domain: "example.com"}
	c3 := &models.Cookie{ID: uuid.New(), Name: "C3", TenantID: otherTenantID, Domain: "other.com"}

	db.Create(c1)
	db.Create(c2)
	db.Create(c3)

	cookies, err := repo.ListByTenant(tenantID, "", nil)
	assert.NoError(t, err)
	assert.Len(t, cookies, 2)
}

func TestCookieRepository_Update_Success(t *testing.T) {
	db := setupCookieRepositoryTestDB()
	repo := NewCookieRepository(db)

	tenantID := uuid.New()
	cookieID := uuid.New()

	cookie := &models.Cookie{
		ID:       cookieID,
		Name:     "Old Name",
		TenantID: tenantID,
		Domain:   "example.com",
	}

	db.Create(cookie)

	cookie.Name = "New Name"
	err := repo.Update(cookie)
	assert.NoError(t, err)

	var updated models.Cookie
	db.First(&updated, "id = ?", cookieID)
	assert.Equal(t, "New Name", updated.Name)
}

func TestCookieRepository_Delete_Success(t *testing.T) {
	db := setupCookieRepositoryTestDB()
	repo := NewCookieRepository(db)

	tenantID := uuid.New()
	cookieID := uuid.New()

	cookie := &models.Cookie{
		ID:       cookieID,
		Name:     "To Delete",
		TenantID: tenantID,
		Domain:   "example.com",
	}

	db.Create(cookie)

	err := repo.Delete(cookieID, tenantID)
	assert.NoError(t, err)

	var count int64
	db.Model(&models.Cookie{}).Where("id = ?", cookieID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestCookieRepository_Delete_WrongTenant(t *testing.T) {
	db := setupCookieRepositoryTestDB()
	repo := NewCookieRepository(db)

	tenantID := uuid.New()
	otherTenantID := uuid.New()
	cookieID := uuid.New()

	cookie := &models.Cookie{
		ID:       cookieID,
		Name:     "To Delete",
		TenantID: tenantID,
		Domain:   "example.com",
	}

	db.Create(cookie)

	// Try to delete with wrong tenant ID (should fail silently or return error depending on GORM,
	// but usually Delete with Where returns nil error if no rows affected, but we want to ensure it DOES NOT delete)
	// My implementation: return r.db.Where(...).Delete(...).Error
	err := repo.Delete(cookieID, otherTenantID)
	assert.NoError(t, err) // GORM delete returns no error if record not found

	// Verify it was NOT deleted
	var count int64
	db.Model(&models.Cookie{}).Where("id = ?", cookieID).Count(&count)
	assert.Equal(t, int64(1), count)
}
