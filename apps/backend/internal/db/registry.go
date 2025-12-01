// db/tenant_db.go
package db

import (
	"pixpivot/arc/internal/models"
	"crypto/sha3"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// gormschema "gorm.io/gorm/schema" // Not needed for DB-per-tenant
)

var (
	tenantDBCache    sync.Map              // schema -> *gorm.DB
)

func RegisterTenantCluster(tenantID uuid.UUID, cluster string) error {
	// Deprecated: Cluster logic removed. Kept for interface compatibility if needed, or can be removed.
	// For now, we just log it.
	log.Info().Str("tenant_id", tenantID.String()).Msg("RegisterTenantCluster called (No-op in Single DB mode)")
	return nil
}

func GetTenantDB(tenantID string) (*gorm.DB, error) {
	// Cache key is now just tenantID
	if db, ok := tenantDBCache.Load(tenantID); ok {
		return db.(*gorm.DB), nil
	}

	// Verify tenant exists in Global DB (MasterDB)
	var tenant models.Tenant
	if err := MasterDB.Where("tenant_id = ?", tenantID).First(&tenant).Error; err != nil {
		log.Error().Err(err).Msg("Tenant not found in Global DB")
		return nil, errors.New("tenant not found")
	}

	return loadTenantDB(tenantID)
}

func loadTenantDB(tenantID string) (*gorm.DB, error) {
	log.Info().Str("tenant_id", tenantID).Msg("Connecting to tenant DB")
	
	// Construct DSN for the specific tenant database
	// We assume the tenant DB resides on the same host/port as the Global DB for now
	// but has a different dbname: "tenant_{uuid}"
	
	// Parse Global DB DSN to get host/user/pass
	// This is a bit hacky, better to have a config helper. 
	// For now, we'll assume standard config is available or parse from MasterDB.Config? 
	// No, MasterDB.Config is gorm config.
	
	// Let's use the standard config package or environment variables again?
	// But we don't have access to `cfg` here easily unless we pass it or store it globally.
	// `db.InitDB` received `cfg`. Let's store the base DSN info in a package variable.
	
	dbName := "tenant_" + strings.ReplaceAll(tenantID, "-", "")
	
	// Reconstruct DSN. 
	// We need to store the base config from InitDB.
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		globalDBHost, globalDBUser, globalDBPassword, dbName, globalDBPort)

	tenantDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error().Err(err).Str("dbname", dbName).Msg("Failed to connect to tenant DB")
		return nil, err
	}

	// AutoMigrate for this tenant DB
	if err := tenantDB.AutoMigrate(
		&models.Consent{},
		&models.ConsentHistory{},
		&models.APIKey{},
		&models.Purpose{},
		&models.DataPrincipal{},
		&models.Grievance{},
		&models.Notification{},
		&models.AuditLog{},
		&models.DSRRequest{},
		// TPRM tables
		&models.TPRMAssessment{},
		&models.TPRMEvidence{},
		&models.TPRMFinding{},
	); err != nil {
		log.Error().Err(err).Msg("AutoMigrate failed for tenant DB")
		return nil, err
	}

	tenantDBCache.Store(tenantID, tenantDB)
	log.Info().Str("tenant_id", tenantID).Msg("Tenant DB connected and cached")
	return tenantDB, nil
}

// Use your master DB to store API keys and tenants
func GetMasterDB() *gorm.DB {
	return MasterDB.Session(&gorm.Session{NewDB: true})
}

func HashAPIKey(rawKey string) string {
	sum := sha3.New256()
	sum.Write([]byte(rawKey))
	return hex.EncodeToString(sum.Sum(nil))
}

func LookupTenantByAPIKey(rawKey string) (*models.Tenant, error) {
	db := GetMasterDB()
	var apiKey models.APIKey
	hashedKey := HashAPIKey(rawKey)
	err := db.Where("hashed_key = ? AND revoked = false", hashedKey).First(&apiKey).Error
	if err != nil {
		return nil, errors.New("API key not found or revoked")
	}
	var tenant models.Tenant
	if err := db.Where("tenant_id = ?", apiKey.TenantID).First(&tenant).Error; err != nil {
		return nil, errors.New("Tenant not found for API key")
	}
	return &tenant, nil
}

