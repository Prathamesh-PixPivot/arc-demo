package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"pixpivot/arc/config"
	"pixpivot/arc/pkg/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type BackupService struct {
	MasterDB *gorm.DB
	Cfg      config.Config
	Cron     *cron.Cron
	S3Client *s3.S3
}

func NewBackupService(masterDB *gorm.DB, cfg config.Config) *BackupService {
	// Initialize S3 Session
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Endpoint:         aws.String(cfg.S3Endpoint),
		Region:           aws.String(cfg.S3Region),
		DisableSSL:       aws.Bool(!cfg.S3UseSSL),
		S3ForcePathStyle: aws.Bool(cfg.S3ForcePathStyle),
	}
	sess, err := session.NewSession(s3Config)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to create S3 session for backups")
	}

	return &BackupService{
		MasterDB: masterDB,
		Cfg:      cfg,
		Cron:     cron.New(),
		S3Client: s3.New(sess),
	}
}

func (s *BackupService) Start() {
	// Schedule backups based on config or default
	// Default: Daily at 2 AM
	_, err := s.Cron.AddFunc("0 2 * * *", func() {
		s.PerformBackup("daily")
	})
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to schedule daily backup")
	}

	// Weekly: Sunday at 3 AM
	_, err = s.Cron.AddFunc("0 3 * * 0", func() {
		s.PerformBackup("weekly")
	})
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to schedule weekly backup")
	}

	// Monthly: 1st of month at 4 AM
	_, err = s.Cron.AddFunc("0 4 1 * *", func() {
		s.PerformBackup("monthly")
	})
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to schedule monthly backup")
	}

	s.Cron.Start()
	log.Logger.Info().Msg("Backup service started")
}

func (s *BackupService) Stop() {
	s.Cron.Stop()
}

func (s *BackupService) PerformBackup(tag string) {
	log.Logger.Info().Str("type", tag).Msg("Starting backup sequence")

	// 1. Backup Global DB (Master)
	s.backupDatabase(s.Cfg.DBName, tag)

	// 2. Get list of Tenant DBs
	var tenants []struct {
		TenantID string
	}
	// Assuming "tenants" table exists in MasterDB and has "tenant_id" column
	if err := s.MasterDB.Table("tenants").Select("tenant_id").Scan(&tenants).Error; err != nil {
		log.Logger.Error().Err(err).Msg("Failed to fetch tenants for backup")
		return
	}

	for _, t := range tenants {
		dbName := "tenant_" + strings.ReplaceAll(t.TenantID, "-", "")
		s.backupDatabase(dbName, tag)
	}
}

func (s *BackupService) backupDatabase(dbName, tag string) {
	log.Logger.Info().Str("db", dbName).Msg("Backing up database")
	
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s_%s_%s.sql", tag, dbName, timestamp)
	filepath := filepath.Join(os.TempDir(), filename)

	os.Setenv("PGPASSWORD", s.Cfg.DBPassword)
	cmd := exec.Command("pg_dump",
		"-h", s.Cfg.DBHost,
		"-p", s.Cfg.DBPort,
		"-U", s.Cfg.DBUser,
		"-d", dbName,
		"-f", filepath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Logger.Error().Err(err).Str("db", dbName).Str("output", string(output)).Msg("pg_dump failed")
		return
	}

	// Upload to S3
	file, err := os.Open(filepath)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to open backup file")
		return
	}
	defer file.Close()
	defer os.Remove(filepath)

	key := fmt.Sprintf("backups/%s/%s/%s", tag, dbName, filename)
	_, err = s.S3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.Cfg.S3Bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to upload backup to S3")
		return
	}
	log.Logger.Info().Str("db", dbName).Msg("Backup uploaded")
}

// ManualBackup triggers a backup immediately
func (s *BackupService) ManualBackup() {
	go s.PerformBackup("manual")
}

