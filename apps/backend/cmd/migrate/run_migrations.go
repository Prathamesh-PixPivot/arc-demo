package main

import (
	"pixpivot/arc/config"
	"pixpivot/arc/internal/migrations"
	"log"
)

func RunMigrations() {
	cfg := config.LoadConfig()
	
	if err := migrations.RunMigrations(&cfg); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	
	log.Println("✅ Database migration completed successfully")
}

