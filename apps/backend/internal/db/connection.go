// db/init_db.go
package db

import (
	"log"
	"pixpivot/arc/config"
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	MasterDB *gorm.DB
	// Global DB Config for constructing tenant DSNs
	// Note: These are package-private, but registry.go is in the same package
	globalDBHost     string
	globalDBPort     string
	globalDBUser     string
	globalDBPassword string
)

func InitDB(cfg config.Config) {
	if cfg.DatabaseURL == "" {
		log.Fatal("InitDB: DATABASE_URL is not set")
	}
	master, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("InitDB: failed to connect to master DB: %v", err)
	}
	if err := master.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Fatalf("InitDB: failed to enable uuid-ossp on master DB: %v", err)
	}
	MasterDB = master

	// Store config for tenant connections
	globalDBHost = cfg.DBHost
	globalDBPort = cfg.DBPort
	globalDBUser = cfg.DBUser
	globalDBPassword = cfg.DBPassword

	log.Println("Connected & configured Master DB (Single Instance)")

	if err := MasterDB.AutoMigrate(
		&models.Tenant{},
		&models.FiduciaryUser{},
		&models.OrganizationEntity{},
		&models.DataPrincipal{},
		&models.UserTenantLink{},
		&models.Notification{},
		&models.Permission{},
		&models.Role{},
		&models.IssuedLicense{},
		&models.BreachImpactAssessment{},
		&models.BreachStakeholder{},
		&models.BreachCommunication{},
		&models.BreachWorkflowStage{},
		&models.BreachEvidence{},
		&models.BreachTimeline{},
		&models.BreachNotificationTemplate{},
	); err != nil {
		log.Fatalf("InitDB: public-schema migration failed on master: %v", err)
	}

	// Seed breach notification templates
	SeedBreachNotificationTemplates(MasterDB)

	log.Println("Master DB migrations complete")
}

func GetTenantIDFromAPIKey(apiKey string) (uuid.UUID, error) {
	var link models.APIKey
	if err := MasterDB.
		Where("hashed_key = ? AND revoked = false", HashAPIKey(apiKey)).
		First(&link).Error; err != nil {
		return uuid.Nil, err
	}
	return link.TenantID, nil
}
