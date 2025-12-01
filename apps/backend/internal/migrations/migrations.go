package migrations

import (
	"pixpivot/arc/config"
	"pixpivot/arc/internal/db"
	"pixpivot/arc/internal/models"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

func RunMigrations(cfg *config.Config) error {
	db.InitDB(*cfg)

	// Run SQL migrations first
	if err := runSQLMigrations(); err != nil {
		return fmt.Errorf("failed to run SQL migrations: %w", err)
	}

	// Then run GORM AutoMigrate for encrypted models
	if err := db.MasterDB.AutoMigrate(
		&models.EncryptedConsent{},
		&models.EncryptedBreachNotification{},
		&models.ConsentLink{},
		&models.ReviewToken{},
		&models.ConsentHistory{},
	); err != nil {
		return fmt.Errorf("failed to migrate encrypted models: %w", err)
	}

	return nil
}

func SetReviewDefaults(cfg *config.Config) error {
	db.InitDB(*cfg)

	result := db.MasterDB.Model(&models.Tenant{}).
		Where("review_frequency_months IS NULL OR review_frequency_months <= 0").
		Update("review_frequency_months", 6)

	if result.Error != nil {
		return fmt.Errorf("failed to update tenants: %w", result.Error)
	}

	err := db.MasterDB.AutoMigrate(&models.ConsentForm{}, &models.ConsentFormPurpose{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	fmt.Printf("✅ Updated %d tenants to default review frequency of 6 months\n", result.RowsAffected)
	return nil
}

func runSQLMigrations() error {
	migrationsDir := "./migrations"

	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".up.sql" {
			if err := executeSQLFile(filepath.Join(migrationsDir, file.Name())); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
			}
			log.Printf("✅ Applied migration: %s\n", file.Name())
		}
	}

	return nil
}

func executeSQLFile(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read SQL file %s: %w", filename, err)
	}

	tx := db.MasterDB.Exec(string(content))
	if tx.Error != nil {
		return fmt.Errorf("failed to execute SQL from %s: %w", filename, tx.Error)
	}

	return nil
}

