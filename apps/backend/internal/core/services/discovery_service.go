package services

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"

	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // Postgres driver
	"gorm.io/gorm"
)

type DataDiscoveryService struct {
	DB *gorm.DB
}

func NewDataDiscoveryService(db *gorm.DB) *DataDiscoveryService {
	return &DataDiscoveryService{DB: db}
}

// Regex patterns for Indian PII
var PIIPatterns = map[string]*regexp.Regexp{
	"aadhaar":         regexp.MustCompile(`^\d{4}\s\d{4}\s\d{4}$|^\d{12}$`),
	"pan":             regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`),
	"passport":        regexp.MustCompile(`^[A-Z]{1}[0-9]{7}$`),
	"voter_id":        regexp.MustCompile(`^[A-Z]{3}[0-9]{7}$`),
	"gstin":           regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`),
	"email":           regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	"phone_in":        regexp.MustCompile(`^(\+91[\-\s]?)?[6789]\d{9}$`),
	"credit_card":     regexp.MustCompile(`^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|6(?:011|5[0-9]{2})[0-9]{12}|(?:2131|1800|35\d{3})\d{11})$`),
}

func (s *DataDiscoveryService) CreateDataSource(ds *models.DataSource) error {
	// TODO: Encrypt password before saving
	return s.DB.Create(ds).Error
}

func (s *DataDiscoveryService) ListDataSources(tenantID uuid.UUID) ([]models.DataSource, error) {
	var sources []models.DataSource
	err := s.DB.Where("tenant_id = ?", tenantID).Find(&sources).Error
	return sources, err
}

func (s *DataDiscoveryService) StartScan(tenantID, dataSourceID uuid.UUID) (*models.DiscoveryJob, error) {
	var ds models.DataSource
	if err := s.DB.First(&ds, "id = ? AND tenant_id = ?", dataSourceID, tenantID).Error; err != nil {
		return nil, fmt.Errorf("data source not found")
	}

	job := &models.DiscoveryJob{
		ID:           uuid.New(),
		TenantID:     tenantID,
		DataSourceID: dataSourceID,
		Status:       "running",
		StartTime:    nowPtr(),
	}

	if err := s.DB.Create(job).Error; err != nil {
		return nil, err
	}

	// Run scan asynchronously
	go s.runScan(job, &ds)

	return job, nil
}

func (s *DataDiscoveryService) runScan(job *models.DiscoveryJob, ds *models.DataSource) {
	log.Printf("Starting scan for job %s on source %s", job.ID, ds.Name)
	
	var err error
	defer func() {
		job.EndTime = nowPtr()
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = err.Error()
		} else {
			job.Status = "completed"
		}
		s.DB.Save(job)
		
		// Update LastScanned on DataSource
		ds.LastScanned = nowPtr()
		s.DB.Save(ds)
	}()

	if ds.Type == "postgres" {
		err = s.scanPostgres(job, ds)
	} else {
		err = fmt.Errorf("unsupported database type: %s", ds.Type)
	}
}

func (s *DataDiscoveryService) scanPostgres(job *models.DiscoveryJob, ds *models.DataSource) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		ds.Host, ds.Port, ds.Username, ds.Password, ds.Database)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Get tables
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}
	job.TotalTables = len(tables)
	s.DB.Save(job) // Update progress

	for _, table := range tables {
		if err := s.scanTable(db, job, table); err != nil {
			log.Printf("Error scanning table %s: %v", table, err)
		}
	}

	return nil
}

func (s *DataDiscoveryService) scanTable(db *sql.DB, job *models.DiscoveryJob, tableName string) error {
	// Get columns
	rows, err := db.Query(fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s'", tableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	var columns []struct{ Name, Type string }
	for rows.Next() {
		var name, dtype string
		if err := rows.Scan(&name, &dtype); err != nil {
			continue
		}
		columns = append(columns, struct{ Name, Type string }{name, dtype})
	}
	job.TotalColumns += len(columns)

	// Fetch sample data (first 50 rows)
	// WARNING: Be careful with large text fields
	query := fmt.Sprintf("SELECT * FROM \"%s\" LIMIT 50", tableName)
	dataRows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer dataRows.Close()

	// Prepare to scan rows
	colCount := len(columns)
	rawResult := make([][]byte, colCount)
	dest := make([]interface{}, colCount)
	for i := range rawResult {
		dest[i] = &rawResult[i]
	}

	for dataRows.Next() {
		if err := dataRows.Scan(dest...); err != nil {
			continue
		}

		for i, raw := range rawResult {
			if raw == nil {
				continue
			}
			val := string(raw)
			
			// Check against patterns
			for piiType, regex := range PIIPatterns {
				if regex.MatchString(val) {
					// Found PII
					s.saveResult(job, tableName, columns[i].Name, columns[i].Type, piiType, val)
				}
			}
		}
	}

	return nil
}

func (s *DataDiscoveryService) saveResult(job *models.DiscoveryJob, table, col, dtype, piiType, val string) {
	// Check if already recorded for this column/type to avoid duplicates
	var count int64
	s.DB.Model(&models.DiscoveryResult{}).
		Where("job_id = ? AND table_name = ? AND column_name = ? AND pii_type = ?", 
			job.ID, table, col, piiType).
		Count(&count)
	
	if count > 0 {
		return
	}

	result := &models.DiscoveryResult{
		ID:             uuid.New(),
		JobID:          job.ID,
		TenantID:       job.TenantID,
		DataSourceID:   job.DataSourceID,
		TableName:      table,
		ColumnName:     col,
		DataType:       dtype,
		PIIType:        piiType,
		Confidence:     0.9, // Regex match is high confidence
		SampleData:     maskData(val),
		Classification: "confidential",
	}
	
	if err := s.DB.Create(result).Error; err == nil {
		job.PIIFound++
		s.DB.Save(job)
	}
}

func maskData(val string) string {
	if len(val) <= 4 {
		return "****"
	}
	return val[:2] + "****" + val[len(val)-2:]
}

func nowPtr() *time.Time {
	t := time.Now()
	return &t
}

func (s *DataDiscoveryService) GetJobResults(jobID uuid.UUID) ([]models.DiscoveryResult, error) {
	var results []models.DiscoveryResult
	err := s.DB.Where("job_id = ?", jobID).Find(&results).Error
	return results, err
}

func (s *DataDiscoveryService) GetDashboardStats(tenantID uuid.UUID) (*models.DataClassificationStats, error) {
	stats := &models.DataClassificationStats{
		ByPIIType:  make(map[string]int),
		BySeverity: make(map[string]int),
	}

	var count int64
	s.DB.Model(&models.DataSource{}).Where("tenant_id = ?", tenantID).Count(&count)
	stats.TotalDataSources = int(count)

	s.DB.Model(&models.DiscoveryResult{}).Where("tenant_id = ?", tenantID).Count(&count)
	stats.TotalPIIColumns = int(count)

	// Group by PII Type
	rows, err := s.DB.Model(&models.DiscoveryResult{}).
		Select("pii_type, count(*)").
		Where("tenant_id = ?", tenantID).
		Group("pii_type").
		Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var pType string
			var cnt int
			rows.Scan(&pType, &cnt)
			stats.ByPIIType[pType] = cnt
		}
	}

	return stats, nil
}

