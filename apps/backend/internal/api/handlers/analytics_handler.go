package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/core/services"

	"github.com/google/uuid"
)

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	service      *services.AnalyticsService
	auditService *services.AuditService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(service *services.AnalyticsService, auditService *services.AuditService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service:      service,
		auditService: auditService,
	}
}

// GetDashboard handles GET /api/v1/fiduciary/analytics/dashboard
func (h *AnalyticsHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	// Parse query parameters for filters
	filter := &models.AnalyticsFilter{
		TenantID: tenantID,
	}

	// Parse date range
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			filter.StartDate = &t
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			filter.EndDate = &t
		}
	}

	// Parse granularity
	if granularity := r.URL.Query().Get("granularity"); granularity != "" {
		filter.Granularity = granularity
	}

	// Parse specific metrics if requested
	if metrics := r.URL.Query()["metrics"]; len(metrics) > 0 {
		filter.Metrics = metrics
	}

	// Generate dashboard
	dashboard, err := h.service.GetDashboard(r.Context(), tenantID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate analytics dashboard")
		return
	}

	writeJSON(w, http.StatusOK, dashboard)
}

// GetConsentAnalytics handles GET /api/v1/fiduciary/analytics/consent
func (h *AnalyticsHandler) GetConsentAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	analytics, err := h.service.GetConsentAnalytics(r.Context(), tenantID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get consent analytics")
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

// GetDSRAnalytics handles GET /api/v1/fiduciary/analytics/dsr
func (h *AnalyticsHandler) GetDSRAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	analytics, err := h.service.GetDSRAnalytics(r.Context(), tenantID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get DSR analytics")
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

// GetUserEngagementAnalytics handles GET /api/v1/fiduciary/analytics/engagement
func (h *AnalyticsHandler) GetUserEngagementAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	analytics, err := h.service.GetUserEngagementAnalytics(r.Context(), tenantID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get user engagement analytics")
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

// GetCookieAnalytics handles GET /api/v1/fiduciary/analytics/cookies
func (h *AnalyticsHandler) GetCookieAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	analytics, err := h.service.GetCookieAnalytics(r.Context(), tenantID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get cookie analytics")
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

// GetComplianceAnalytics handles GET /api/v1/fiduciary/analytics/compliance
func (h *AnalyticsHandler) GetComplianceAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	analytics, err := h.service.GetComplianceAnalytics(r.Context(), tenantID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get compliance analytics")
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

// GetPerformanceAnalytics handles GET /api/v1/fiduciary/analytics/performance
func (h *AnalyticsHandler) GetPerformanceAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	analytics, err := h.service.GetPerformanceAnalytics(r.Context(), tenantID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get performance analytics")
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

// GetRevenueAnalytics handles GET /api/v1/fiduciary/analytics/revenue
func (h *AnalyticsHandler) GetRevenueAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	analytics, err := h.service.GetRevenueAnalytics(r.Context(), tenantID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get revenue analytics")
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

// GetConsentTrend handles GET /api/v1/fiduciary/analytics/trends/consent
func (h *AnalyticsHandler) GetConsentTrend(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	trend := h.service.GenerateConsentTrend(r.Context(), tenantID, filter)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"trend":  trend,
		"period": filter.Granularity,
	})
}

// GetDSRTrend handles GET /api/v1/fiduciary/analytics/trends/dsr
func (h *AnalyticsHandler) GetDSRTrend(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	trend := h.service.GenerateDSRTrend(r.Context(), tenantID, filter)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"trend":  trend,
		"period": filter.Granularity,
	})
}

// GetComplianceTrend handles GET /api/v1/fiduciary/analytics/trends/compliance
func (h *AnalyticsHandler) GetComplianceTrend(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	filter := h.parseAnalyticsFilter(r, tenantID)

	trend := h.service.GenerateComplianceTrend(r.Context(), tenantID, filter)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"trend":  trend,
		"period": filter.Granularity,
	})
}

// ExportAnalytics handles GET /api/v1/fiduciary/analytics/export
func (h *AnalyticsHandler) ExportAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	// Parse export format
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	filter := h.parseAnalyticsFilter(r, tenantID)

	// Generate complete dashboard
	dashboard, err := h.service.GetDashboard(r.Context(), tenantID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate analytics for export")
		return
	}

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=analytics_export.json")
		json.NewEncoder(w).Encode(dashboard)

	case "csv":
		// Convert to CSV format (simplified example)
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=analytics_export.csv")
		h.exportAsCSV(w, dashboard)

	default:
		writeError(w, http.StatusBadRequest, "Unsupported export format")
	}
}

// GetRealTimeMetrics handles GET /api/v1/fiduciary/analytics/realtime
func (h *AnalyticsHandler) GetRealTimeMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	// This would typically connect to a real-time data source
	// For now, return current snapshot
	_ = tenantID // TODO: Use tenantID for real-time metrics
	metrics := map[string]interface{}{
		"timestamp":               time.Now(),
		"active_users":            125,
		"active_sessions":         89,
		"consents_last_hour":      23,
		"dsr_pending":             5,
		"api_requests_per_minute": 450,
		"avg_response_time_ms":    145,
	}

	writeJSON(w, http.StatusOK, metrics)
}

// GetCustomReport handles POST /api/v1/fiduciary/analytics/custom-report
func (h *AnalyticsHandler) GetCustomReport(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	var request struct {
		Metrics    []string               `json:"metrics"`
		Dimensions []string               `json:"dimensions"`
		Filters    map[string]interface{} `json:"filters"`
		StartDate  string                 `json:"start_date"`
		EndDate    string                 `json:"end_date"`
		GroupBy    string                 `json:"group_by"`
		OrderBy    string                 `json:"order_by"`
		Limit      int                    `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate and process custom report request
	if len(request.Metrics) == 0 {
		writeError(w, http.StatusBadRequest, "At least one metric is required")
		return
	}

	// Generate custom report (simplified)
	report := map[string]interface{}{
		"tenant_id":  tenantID,
		"generated":  time.Now(),
		"metrics":    request.Metrics,
		"dimensions": request.Dimensions,
		"data": []map[string]interface{}{
			// This would contain actual data based on the request
			{"dimension": "value1", "metric": 100},
			{"dimension": "value2", "metric": 200},
		},
	}

	writeJSON(w, http.StatusOK, report)
}

// GetAlerts handles GET /api/v1/fiduciary/analytics/alerts
func (h *AnalyticsHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(uuid.UUID)

	// Parse alert type filter
	alertType := r.URL.Query().Get("type")
	category := r.URL.Query().Get("category")

	// Generate current dashboard to check for alerts
	dashboard, err := h.service.GetDashboard(r.Context(), tenantID, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate alerts")
		return
	}

	alerts := dashboard.ActiveAlerts

	// TODO: Implement proper alert filtering when alert models are defined
	// For now, just return the alerts as-is
	_ = alertType // Suppress unused variable warning
	_ = category  // Suppress unused variable warning

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// Helper function to parse analytics filter from request
func (h *AnalyticsHandler) parseAnalyticsFilter(r *http.Request, tenantID uuid.UUID) *models.AnalyticsFilter {
	filter := &models.AnalyticsFilter{
		TenantID: tenantID,
	}

	// Parse date range
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			filter.StartDate = &t
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			filter.EndDate = &t
		}
	}

	// Parse granularity
	if granularity := r.URL.Query().Get("granularity"); granularity != "" {
		filter.Granularity = granularity
	}

	// Parse dimensions
	if dimensions := r.URL.Query()["dimensions"]; len(dimensions) > 0 {
		filter.Dimensions = dimensions
	}

	return filter
}

// Helper function to export dashboard as CSV
func (h *AnalyticsHandler) exportAsCSV(w http.ResponseWriter, dashboard *services.Dashboard) {
	// Simplified CSV export - would need proper CSV library in production
	w.Write([]byte("Metric,Value,Unit\n"))

	w.Write([]byte("Total Consents," + strconv.FormatInt(dashboard.TotalConsents, 10) + ",count\n"))
	w.Write([]byte("Active Consents," + strconv.FormatInt(dashboard.ActiveConsents, 10) + ",count\n"))
	w.Write([]byte("Withdrawn Consents," + strconv.FormatInt(dashboard.WithdrawnConsents, 10) + ",count\n"))
	w.Write([]byte("Consent Rate," + strconv.FormatFloat(dashboard.ConsentRate, 'f', 2, 64) + ",%\n"))
}

