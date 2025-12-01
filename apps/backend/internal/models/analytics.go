package models

import (
	"time"

	"github.com/google/uuid"
)

// AnalyticsMetric represents a single analytics metric
type AnalyticsMetric struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TenantID    uuid.UUID              `json:"tenant_id" gorm:"type:uuid;not null;index"`
	MetricType  string                 `json:"metric_type" gorm:"type:varchar(100);not null;index"`
	MetricName  string                 `json:"metric_name" gorm:"type:varchar(255);not null"`
	Value       float64                `json:"value" gorm:"not null"`
	Unit        string                 `json:"unit" gorm:"type:varchar(50)"`
	Period      string                 `json:"period" gorm:"type:varchar(50)"` // daily, weekly, monthly, yearly
	PeriodStart time.Time              `json:"period_start"`
	PeriodEnd   time.Time              `json:"period_end"`
	Dimensions  map[string]interface{} `json:"dimensions" gorm:"type:jsonb"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ConsentAnalytics represents consent-specific analytics
type ConsentAnalytics struct {
	// Overview Metrics
	TotalConsents           int64   `json:"total_consents"`
	ActiveConsents          int64   `json:"active_consents"`
	ExpiredConsents         int64   `json:"expired_consents"`
	RevokedConsents         int64   `json:"revoked_consents"`
	PendingConsents         int64   `json:"pending_consents"`
	ConsentRate             float64 `json:"consent_rate"` // Percentage of users who consented
	
	// Time-based Metrics
	ConsentsToday           int64   `json:"consents_today"`
	ConsentsThisWeek        int64   `json:"consents_this_week"`
	ConsentsThisMonth       int64   `json:"consents_this_month"`
	ConsentsThisYear        int64   `json:"consents_this_year"`
	
	// Average Metrics
	AvgConsentDuration      float64 `json:"avg_consent_duration_days"`
	AvgTimeToConsent        float64 `json:"avg_time_to_consent_minutes"`
	AvgPurposesPerConsent   float64 `json:"avg_purposes_per_consent"`
	
	// Conversion Metrics
	ConversionRate          float64 `json:"conversion_rate"`
	AbandonmentRate         float64 `json:"abandonment_rate"`
	RetentionRate           float64 `json:"retention_rate"`
	ChurnRate               float64 `json:"churn_rate"`
	
	// Purpose-based Metrics
	PurposeBreakdown        map[string]int64   `json:"purpose_breakdown"`
	MostConsentedPurpose    string             `json:"most_consented_purpose"`
	LeastConsentedPurpose   string             `json:"least_consented_purpose"`
	PurposeConsentRates     map[string]float64 `json:"purpose_consent_rates"`
	
	// Channel Metrics
	ConsentsByChannel       map[string]int64   `json:"consents_by_channel"` // web, mobile, api, etc.
	ChannelConversionRates  map[string]float64 `json:"channel_conversion_rates"`
	
	// Geographic Metrics
	ConsentsByCountry       map[string]int64   `json:"consents_by_country"`
	ConsentsByState         map[string]int64   `json:"consents_by_state"`
	ConsentsByCity          map[string]int64   `json:"consents_by_city"`
	
	// Demographic Metrics
	ConsentsByAgeGroup      map[string]int64   `json:"consents_by_age_group"`
	ConsentsByGender        map[string]int64   `json:"consents_by_gender"`
	
	// Compliance Metrics
	ComplianceScore         float64            `json:"compliance_score"`
	DataBreaches            int64              `json:"data_breaches"`
	DSRRequests             int64              `json:"dsr_requests"`
	DSRCompletionRate       float64            `json:"dsr_completion_rate"`
	AvgDSRResponseTime      float64            `json:"avg_dsr_response_time_hours"`
}

// DSRAnalytics represents DSR-specific analytics
type DSRAnalytics struct {
	// Request Metrics
	TotalRequests           int64              `json:"total_requests"`
	PendingRequests         int64              `json:"pending_requests"`
	CompletedRequests       int64              `json:"completed_requests"`
	RejectedRequests        int64              `json:"rejected_requests"`
	
	// Type Breakdown
	AccessRequests          int64              `json:"access_requests"`
	CorrectionRequests      int64              `json:"correction_requests"`
	ErasureRequests         int64              `json:"erasure_requests"`
	PortabilityRequests     int64              `json:"portability_requests"`
	ObjectionRequests       int64              `json:"objection_requests"`
	
	// Performance Metrics
	AvgResponseTime         float64            `json:"avg_response_time_hours"`
	CompletionRate          float64            `json:"completion_rate"`
	SLACompliance           float64            `json:"sla_compliance_rate"`
	
	// Appeal Metrics
	TotalAppeals            int64              `json:"total_appeals"`
	AppealSuccessRate       float64            `json:"appeal_success_rate"`
	
	// Channel Metrics
	RequestsByChannel       map[string]int64   `json:"requests_by_channel"`
}

// UserEngagementAnalytics represents user engagement metrics
type UserEngagementAnalytics struct {
	// User Metrics
	TotalUsers              int64              `json:"total_users"`
	ActiveUsers             int64              `json:"active_users"`
	NewUsers                int64              `json:"new_users"`
	ReturningUsers          int64              `json:"returning_users"`
	
	// Session Metrics
	TotalSessions           int64              `json:"total_sessions"`
	AvgSessionDuration      float64            `json:"avg_session_duration_minutes"`
	BounceRate              float64            `json:"bounce_rate"`
	
	// Interaction Metrics
	ConsentFormViews        int64              `json:"consent_form_views"`
	ConsentFormCompletions  int64              `json:"consent_form_completions"`
	PreferenceCenterViews   int64              `json:"preference_center_views"`
	PreferenceUpdates       int64              `json:"preference_updates"`
	
	// Device Metrics
	DesktopUsers            int64              `json:"desktop_users"`
	MobileUsers             int64              `json:"mobile_users"`
	TabletUsers             int64              `json:"tablet_users"`
	
	// Browser Metrics
	BrowserBreakdown        map[string]int64   `json:"browser_breakdown"`
}

// CookieAnalytics represents cookie management analytics
type CookieAnalytics struct {
	// Cookie Metrics
	TotalCookies            int64              `json:"total_cookies"`
	FirstPartyCookies       int64              `json:"first_party_cookies"`
	ThirdPartyCookies       int64              `json:"third_party_cookies"`
	
	// Category Breakdown
	NecessaryCookies        int64              `json:"necessary_cookies"`
	FunctionalCookies       int64              `json:"functional_cookies"`
	AnalyticsCookies        int64              `json:"analytics_cookies"`
	MarketingCookies        int64              `json:"marketing_cookies"`
	
	// Consent Metrics
	CookieConsentRate       float64            `json:"cookie_consent_rate"`
	CategoryConsentRates    map[string]float64 `json:"category_consent_rates"`
	
	// Scanning Metrics
	WebsitesScanned         int64              `json:"websites_scanned"`
	NewCookiesDetected      int64              `json:"new_cookies_detected"`
	UnknownCookies          int64              `json:"unknown_cookies"`
	
	// Blocking Metrics
	CookiesBlocked          int64              `json:"cookies_blocked"`
	BlockingEffectiveness   float64            `json:"blocking_effectiveness"`
}

// ComplianceAnalytics represents compliance-related analytics
type ComplianceAnalytics struct {
	// Overall Compliance
	OverallComplianceScore  float64            `json:"overall_compliance_score"`
	DPDPAComplianceScore    float64            `json:"dpdpa_compliance_score"`
	
	// Audit Metrics
	TotalAudits             int64              `json:"total_audits"`
	PassedAudits            int64              `json:"passed_audits"`
	FailedAudits            int64              `json:"failed_audits"`
	AuditPassRate           float64            `json:"audit_pass_rate"`
	
	// Risk Metrics
	HighRiskAreas           int64              `json:"high_risk_areas"`
	MediumRiskAreas         int64              `json:"medium_risk_areas"`
	LowRiskAreas            int64              `json:"low_risk_areas"`
	RiskScore               float64            `json:"risk_score"`
	
	// Breach Metrics
	DataBreaches            int64              `json:"data_breaches"`
	AffectedUsers           int64              `json:"affected_users"`
	BreachNotificationTime  float64            `json:"avg_breach_notification_hours"`
	
	// Policy Metrics
	ActivePolicies          int64              `json:"active_policies"`
	PolicyViolations        int64              `json:"policy_violations"`
	PolicyComplianceRate    float64            `json:"policy_compliance_rate"`
}

// PerformanceAnalytics represents system performance metrics
type PerformanceAnalytics struct {
	// API Metrics
	APIRequests             int64              `json:"api_requests"`
	APIResponseTime         float64            `json:"avg_api_response_ms"`
	APIErrorRate            float64            `json:"api_error_rate"`
	APIUptime               float64            `json:"api_uptime_percentage"`
	
	// Database Metrics
	DatabaseQueries         int64              `json:"database_queries"`
	DatabaseResponseTime    float64            `json:"avg_database_response_ms"`
	DatabaseConnections     int64              `json:"active_database_connections"`
	
	// System Metrics
	CPUUsage                float64            `json:"cpu_usage_percentage"`
	MemoryUsage             float64            `json:"memory_usage_percentage"`
	DiskUsage               float64            `json:"disk_usage_percentage"`
	
	// Queue Metrics
	QueuedJobs              int64              `json:"queued_jobs"`
	ProcessedJobs           int64              `json:"processed_jobs"`
	FailedJobs              int64              `json:"failed_jobs"`
	JobProcessingTime       float64            `json:"avg_job_processing_seconds"`
}

// RevenueAnalytics represents revenue and business metrics
type RevenueAnalytics struct {
	// Revenue Metrics
	TotalRevenue            float64            `json:"total_revenue"`
	RecurringRevenue        float64            `json:"recurring_revenue"`
	AverageRevenuePerUser   float64            `json:"arpu"`
	CustomerLifetimeValue   float64            `json:"clv"`
	
	// Subscription Metrics
	ActiveSubscriptions     int64              `json:"active_subscriptions"`
	NewSubscriptions        int64              `json:"new_subscriptions"`
	CancelledSubscriptions  int64              `json:"cancelled_subscriptions"`
	SubscriptionChurnRate   float64            `json:"subscription_churn_rate"`
	
	// Cost Metrics
	CustomerAcquisitionCost float64            `json:"cac"`
	OperationalCost         float64            `json:"operational_cost"`
	ComplianceCost          float64            `json:"compliance_cost"`
	
	// ROI Metrics
	ReturnOnInvestment      float64            `json:"roi"`
	PaybackPeriod           float64            `json:"payback_period_months"`
}

// AnalyticsDashboard represents the complete analytics dashboard
type AnalyticsDashboard struct {
	ID                      uuid.UUID                `json:"id"`
	TenantID                uuid.UUID                `json:"tenant_id"`
	GeneratedAt             time.Time                `json:"generated_at"`
	PeriodStart             time.Time                `json:"period_start"`
	PeriodEnd               time.Time                `json:"period_end"`
	
	// Core Analytics
	ConsentAnalytics        *ConsentAnalytics        `json:"consent_analytics"`
	DSRAnalytics            *DSRAnalytics            `json:"dsr_analytics"`
	UserEngagement          *UserEngagementAnalytics `json:"user_engagement"`
	CookieAnalytics         *CookieAnalytics         `json:"cookie_analytics"`
	ComplianceAnalytics     *ComplianceAnalytics     `json:"compliance_analytics"`
	PerformanceAnalytics    *PerformanceAnalytics    `json:"performance_analytics"`
	RevenueAnalytics        *RevenueAnalytics        `json:"revenue_analytics"`
	
	// Trends
	ConsentTrend            []TrendPoint             `json:"consent_trend"`
	DSRTrend                []TrendPoint             `json:"dsr_trend"`
	ComplianceTrend         []TrendPoint             `json:"compliance_trend"`
	
	// Alerts
	ActiveAlerts            []AnalyticsAlert         `json:"active_alerts"`
	
	// Recommendations
	Recommendations         []string                 `json:"recommendations"`
}

// TrendPoint represents a point in a trend line
type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

// AnalyticsAlert represents an analytics-based alert
type AnalyticsAlert struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"` // warning, critical, info
	Category    string    `json:"category"`
	Message     string    `json:"message"`
	Metric      string    `json:"metric"`
	Threshold   float64   `json:"threshold"`
	ActualValue float64   `json:"actual_value"`
	CreatedAt   time.Time `json:"created_at"`
}

// MetricDefinition defines a metric configuration
type MetricDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Type        string                 `json:"type"` // counter, gauge, histogram
	Unit        string                 `json:"unit"`
	Query       string                 `json:"query"` // SQL query to calculate metric
	Threshold   *MetricThreshold       `json:"threshold,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// MetricThreshold defines thresholds for alerts
type MetricThreshold struct {
	Warning  float64 `json:"warning"`
	Critical float64 `json:"critical"`
	Operator string  `json:"operator"` // gt, lt, eq, gte, lte
}

// AnalyticsFilter represents filters for analytics queries
type AnalyticsFilter struct {
	TenantID    uuid.UUID  `json:"tenant_id,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Granularity string     `json:"granularity,omitempty"` // hour, day, week, month, year
	Metrics     []string   `json:"metrics,omitempty"`
	Dimensions  []string   `json:"dimensions,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
}

