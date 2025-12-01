package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"pixpivot/arc/internal/models"
	"gorm.io/gorm"
)

// AnalyticsRepository handles database operations for analytics
type AnalyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// SaveMetric saves an analytics metric to the database
func (r *AnalyticsRepository) SaveMetric(ctx context.Context, metric *models.AnalyticsMetric) error {
	if metric.ID == uuid.Nil {
		metric.ID = uuid.New()
	}
	
	err := r.db.WithContext(ctx).Create(metric).Error
	if err != nil {
		return fmt.Errorf("failed to save metric: %w", err)
	}
	
	return nil
}

// GetMetrics retrieves metrics based on filters
func (r *AnalyticsRepository) GetMetrics(ctx context.Context, filter *models.AnalyticsFilter) ([]models.AnalyticsMetric, error) {
	query := r.db.WithContext(ctx).Model(&models.AnalyticsMetric{})
	
	if filter.TenantID != uuid.Nil {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	
	if filter.StartDate != nil {
		query = query.Where("period_start >= ?", *filter.StartDate)
	}
	
	if filter.EndDate != nil {
		query = query.Where("period_end <= ?", *filter.EndDate)
	}
	
	if len(filter.Metrics) > 0 {
		query = query.Where("metric_name IN ?", filter.Metrics)
	}
	
	var metrics []models.AnalyticsMetric
	err := query.Order("period_start DESC").Find(&metrics).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	
	return metrics, nil
}

// GetMetricByType retrieves a specific metric type
func (r *AnalyticsRepository) GetMetricByType(ctx context.Context, tenantID uuid.UUID, metricType string, period string) (*models.AnalyticsMetric, error) {
	var metric models.AnalyticsMetric
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND metric_type = ? AND period = ?", tenantID, metricType, period).
		Order("created_at DESC").
		First(&metric).Error
	
	if err != nil {
		return nil, fmt.Errorf("metric not found: %w", err)
	}
	
	return &metric, nil
}

// AggregateMetrics aggregates metrics over a time period
func (r *AnalyticsRepository) AggregateMetrics(ctx context.Context, tenantID uuid.UUID, metricName string, aggregation string, startDate, endDate time.Time) (float64, error) {
	var result float64
	
	query := r.db.WithContext(ctx).
		Model(&models.AnalyticsMetric{}).
		Where("tenant_id = ? AND metric_name = ? AND period_start >= ? AND period_end <= ?",
			tenantID, metricName, startDate, endDate)
	
	switch aggregation {
	case "sum":
		query = query.Select("COALESCE(SUM(value), 0)")
	case "avg":
		query = query.Select("COALESCE(AVG(value), 0)")
	case "max":
		query = query.Select("COALESCE(MAX(value), 0)")
	case "min":
		query = query.Select("COALESCE(MIN(value), 0)")
	case "count":
		query = query.Select("COALESCE(COUNT(*), 0)")
	default:
		return 0, fmt.Errorf("unsupported aggregation: %s", aggregation)
	}
	
	err := query.Scan(&result).Error
	if err != nil {
		return 0, fmt.Errorf("failed to aggregate metrics: %w", err)
	}
	
	return result, nil
}

// GetConsentMetrics retrieves consent-related metrics
func (r *AnalyticsRepository) GetConsentMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	
	// Total consents
	var totalConsents int64
	r.db.WithContext(ctx).
		Model(&models.UserConsent{}).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
		Count(&totalConsents)
	metrics["total_consents"] = totalConsents
	
	// Active consents
	var activeConsents int64
	r.db.WithContext(ctx).
		Model(&models.UserConsent{}).
		Where("tenant_id = ? AND status = ? AND (expires_at IS NULL OR expires_at > ?)", 
			tenantID, "active", time.Now()).
		Count(&activeConsents)
	metrics["active_consents"] = activeConsents
	
	// Consent by status
	type statusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []statusCount
	r.db.WithContext(ctx).
		Model(&models.UserConsent{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&statusCounts)
	
	statusMap := make(map[string]int64)
	for _, sc := range statusCounts {
		statusMap[sc.Status] = sc.Count
	}
	metrics["consent_by_status"] = statusMap
	
	return metrics, nil
}

// GetDSRMetrics retrieves DSR-related metrics
func (r *AnalyticsRepository) GetDSRMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	
	// Total DSR requests
	var totalRequests int64
	r.db.WithContext(ctx).
		Model(&models.DSRRequest{}).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
		Count(&totalRequests)
	metrics["total_requests"] = totalRequests
	
	// DSR by type
	type typeCount struct {
		RequestType string
		Count       int64
	}
	var typeCounts []typeCount
	r.db.WithContext(ctx).
		Model(&models.DSRRequest{}).
		Select("request_type, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("request_type").
		Scan(&typeCounts)
	
	typeMap := make(map[string]int64)
	for _, tc := range typeCounts {
		typeMap[tc.RequestType] = tc.Count
	}
	metrics["dsr_by_type"] = typeMap
	
	// Average response time
	var avgResponseTime float64
	r.db.WithContext(ctx).
		Model(&models.DSRRequest{}).
		Select("AVG(EXTRACT(EPOCH FROM (updated_at - created_at))/3600)").
		Where("tenant_id = ? AND status = ?", tenantID, "completed").
		Scan(&avgResponseTime)
	metrics["avg_response_time_hours"] = avgResponseTime
	
	return metrics, nil
}

// GetUserMetrics retrieves user-related metrics
func (r *AnalyticsRepository) GetUserMetrics(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	
	// Total users
	var totalUsers int64
	r.db.WithContext(ctx).
		Model(&models.DataPrincipal{}).
		Count(&totalUsers)
	metrics["total_users"] = totalUsers
	
	// New users in period
	var newUsers int64
	r.db.WithContext(ctx).
		Model(&models.DataPrincipal{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&newUsers)
	metrics["new_users"] = newUsers
	
	// Active users (users with consent in last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var activeUsers int64
	r.db.WithContext(ctx).
		Model(&models.DataPrincipal{}).
		Joins("JOIN user_consents ON data_principals.id = user_consents.data_principal_id").
		Where("user_consents.created_at >= ?", thirtyDaysAgo).
		Distinct("data_principals.id").
		Count(&activeUsers)
	metrics["active_users"] = activeUsers
	
	return metrics, nil
}

// GetCookieMetrics retrieves cookie-related metrics
func (r *AnalyticsRepository) GetCookieMetrics(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	
	// Total cookies
	var totalCookies int64
	r.db.WithContext(ctx).
		Model(&models.Cookie{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalCookies)
	metrics["total_cookies"] = totalCookies
	
	// Cookies by category
	type categoryCount struct {
		Category string
		Count    int64
	}
	var categoryCounts []categoryCount
	r.db.WithContext(ctx).
		Model(&models.Cookie{}).
		Select("category, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("category").
		Scan(&categoryCounts)
	
	categoryMap := make(map[string]int64)
	for _, cc := range categoryCounts {
		categoryMap[cc.Category] = cc.Count
	}
	metrics["cookies_by_category"] = categoryMap
	
	// First party vs third party
	var firstPartyCookies int64
	r.db.WithContext(ctx).
		Model(&models.Cookie{}).
		Where("tenant_id = ? AND is_first_party = ?", tenantID, true).
		Count(&firstPartyCookies)
	metrics["first_party_cookies"] = firstPartyCookies
	metrics["third_party_cookies"] = totalCookies - firstPartyCookies
	
	return metrics, nil
}

// GetComplianceMetrics retrieves compliance-related metrics
func (r *AnalyticsRepository) GetComplianceMetrics(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	
	// Consent forms compliance
	var totalForms int64
	r.db.WithContext(ctx).
		Model(&models.ConsentForm{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalForms)
	
	var publishedForms int64
	r.db.WithContext(ctx).
		Model(&models.ConsentForm{}).
		Where("tenant_id = ? AND status = ?", tenantID, "published").
		Count(&publishedForms)
	
	metrics["total_consent_forms"] = totalForms
	metrics["published_consent_forms"] = publishedForms
	
	if totalForms > 0 {
		metrics["form_compliance_rate"] = float64(publishedForms) / float64(totalForms) * 100
	} else {
		metrics["form_compliance_rate"] = 0.0
	}
	
	// Purpose compliance
	var totalPurposes int64
	r.db.WithContext(ctx).
		Model(&models.Purpose{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalPurposes)
	
	var compliantPurposes int64
	r.db.WithContext(ctx).
		Model(&models.Purpose{}).
		Where("tenant_id = ? AND legal_basis IS NOT NULL AND retention_period > 0", tenantID).
		Count(&compliantPurposes)
	
	metrics["total_purposes"] = totalPurposes
	metrics["compliant_purposes"] = compliantPurposes
	
	if totalPurposes > 0 {
		metrics["purpose_compliance_rate"] = float64(compliantPurposes) / float64(totalPurposes) * 100
	} else {
		metrics["purpose_compliance_rate"] = 0.0
	}
	
	return metrics, nil
}

// GetPerformanceMetrics retrieves system performance metrics
func (r *AnalyticsRepository) GetPerformanceMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	
	// Database connection stats
	sqlDB, err := r.db.DB()
	if err == nil {
		stats := sqlDB.Stats()
		metrics["db_open_connections"] = stats.OpenConnections
		metrics["db_in_use"] = stats.InUse
		metrics["db_idle"] = stats.Idle
		metrics["db_wait_count"] = stats.WaitCount
		metrics["db_wait_duration_ms"] = stats.WaitDuration.Milliseconds()
	}
	
	// Query performance (example)
	var avgQueryTime float64
	r.db.WithContext(ctx).
		Raw("SELECT AVG(query_time_ms) FROM query_logs WHERE created_at > NOW() - INTERVAL '1 hour'").
		Scan(&avgQueryTime)
	metrics["avg_query_time_ms"] = avgQueryTime
	
	return metrics, nil
}

// SaveDashboard saves a complete analytics dashboard snapshot
func (r *AnalyticsRepository) SaveDashboard(ctx context.Context, dashboard *models.AnalyticsDashboard) error {
	// This could save the dashboard to a separate table for historical tracking
	// For now, we'll just validate it
	if dashboard.ID == uuid.Nil {
		dashboard.ID = uuid.New()
	}
	
	// Could implement actual persistence here if needed
	return nil
}

// GetHistoricalDashboard retrieves a historical dashboard snapshot
func (r *AnalyticsRepository) GetHistoricalDashboard(ctx context.Context, dashboardID uuid.UUID) (*models.AnalyticsDashboard, error) {
	// Placeholder for retrieving historical dashboard data
	// Would query from a dashboards table if implemented
	return nil, fmt.Errorf("historical dashboard not found")
}

// CleanupOldMetrics removes metrics older than retention period
func (r *AnalyticsRepository) CleanupOldMetrics(ctx context.Context, retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	
	err := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&models.AnalyticsMetric{}).Error
	
	if err != nil {
		return fmt.Errorf("failed to cleanup old metrics: %w", err)
	}
	
	return nil
}

// GetTopPurposes retrieves the most consented purposes
func (r *AnalyticsRepository) GetTopPurposes(ctx context.Context, tenantID uuid.UUID, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT p.id, p.name, COUNT(uc.id) as consent_count
			FROM purposes p
			JOIN user_consents uc ON p.id = ANY(uc.purposes)
			WHERE p.tenant_id = ?
			GROUP BY p.id, p.name
			ORDER BY consent_count DESC
			LIMIT ?
		`, tenantID, limit).
		Scan(&results).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get top purposes: %w", err)
	}
	
	return results, nil
}

// GetConsentTrendData retrieves consent trend data for charts
func (r *AnalyticsRepository) GetConsentTrendData(ctx context.Context, tenantID uuid.UUID, days int) ([]models.TrendPoint, error) {
	var trends []models.TrendPoint
	
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DATE(created_at) as timestamp, COUNT(*) as value
			FROM user_consents
			WHERE tenant_id = ? AND created_at >= NOW() - INTERVAL '? days'
			GROUP BY DATE(created_at)
			ORDER BY timestamp
		`, tenantID, days).
		Scan(&trends).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get consent trend data: %w", err)
	}
	
	return trends, nil
}

