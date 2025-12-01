package main

import (
	"fmt"
	"os"
	"strings"

	"pixpivot/arc/config"
	"pixpivot/arc/internal/db"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/pkg/log"

	"github.com/google/uuid"
)

func main() {
	log.InitLogger()
	cfg := config.LoadConfig()
	db.InitDB(cfg)

	tenantName := os.Getenv("TENANT_NAME")
	cluster := os.Getenv("TENANT_CLUSTER")

	if tenantName == "" || cluster == "" {
		log.Logger.Fatal().Msg("TENANT_NAME and TENANT_CLUSTER must be set")
	}

	tenantID := uuid.New()
	// Align with registry.go: tenant_{uuid_no_dashes}
	dbName := "tenant_" + strings.ReplaceAll(tenantID.String(), "-", "")

	tenant := &models.Tenant{
		TenantID:              tenantID,
		Name:                  tenantName,
		Cluster:               cluster,
		ReviewFrequencyMonths: 6,
	}

	if err := db.MasterDB.Create(&tenant).Error; err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to insert tenant")
	}

	// Create Database
	// Note: We cannot use parameter substitution for identifiers like database names
	if err := db.MasterDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error; err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to create tenant database")
	}

	// Connect to the new Tenant DB
	// We need to construct the DSN manually as we don't have the helper exposed easily or we can use the one from registry if we import it?
	// registry.go is in "pixpivot/arc/internal/db".
	// But GetTenantDB uses cache and checks MasterDB. We just inserted into MasterDB, so it should work.

	_, err := db.GetTenantDB(tenantID.String())
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to connect to new tenant DB")
	}

	// AutoMigrate is already handled in GetTenantDB, but we can call it again or ensure it's done.
	// GetTenantDB calls loadTenantDB which calls AutoMigrate. So we are good.

	log.Logger.Info().
		Str("tenant_name", tenantName).
		Str("tenant_id", tenantID.String()).
		Str("cluster", cluster).
		Str("dbname", dbName).
		Msg("Tenant setup complete")
}

