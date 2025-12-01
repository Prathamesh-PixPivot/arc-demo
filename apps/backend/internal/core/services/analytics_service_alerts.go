package services

import (
	"fmt"
	"time"

	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
)

// GenerateAlerts generates analytics-based alerts (continuation)
func (s *AnalyticsService) GenerateAlertsComplete(dashboard *models.AnalyticsDashboard) []models.AnalyticsAlert {
	alerts := []models.AnalyticsAlert{}

	// Check consent rate
	if dashboard.ConsentAnalytics != nil && dashboard.ConsentAnalytics.ConsentRate < 70 {
		alerts = append(alerts, models.AnalyticsAlert{
			ID:          uuid.New(),
			Type:        "warning",
			Category:    "consent",
			Message:     "Consent rate is below threshold",
			Metric:      "consent_rate",
			Threshold:   70.0,
			ActualValue: dashboard.ConsentAnalytics.ConsentRate,
			CreatedAt:   time.Now(),
		})
	}

	// Check DSR response time
	if dashboard.DSRAnalytics != nil && dashboard.DSRAnalytics.AvgResponseTime > 48 {
		alerts = append(alerts, models.AnalyticsAlert{
			ID:          uuid.New(),
			Type:        "critical",
			Category:    "dsr",
			Message:     "DSR response time exceeds 48 hours",
			Metric:      "dsr_response_time",
			Threshold:   48.0,
			ActualValue: dashboard.DSRAnalytics.AvgResponseTime,
			CreatedAt:   time.Now(),
		})
	}

	// Check compliance score
	if dashboard.ComplianceAnalytics != nil && dashboard.ComplianceAnalytics.OverallComplianceScore < 90 {
		alerts = append(alerts, models.AnalyticsAlert{
			ID:          uuid.New(),
			Type:        "warning",
			Category:    "compliance",
			Message:     "Compliance score is below 90%",
			Metric:      "compliance_score",
			Threshold:   90.0,
			ActualValue: dashboard.ComplianceAnalytics.OverallComplianceScore,
			CreatedAt:   time.Now(),
		})
	}

	// Check API error rate
	if dashboard.PerformanceAnalytics != nil && dashboard.PerformanceAnalytics.APIErrorRate > 1.0 {
		alerts = append(alerts, models.AnalyticsAlert{
			ID:          uuid.New(),
			Type:        "warning",
			Category:    "performance",
			Message:     "API error rate is above 1%",
			Metric:      "api_error_rate",
			Threshold:   1.0,
			ActualValue: dashboard.PerformanceAnalytics.APIErrorRate,
			CreatedAt:   time.Now(),
		})
	}

	// Check churn rate
	if dashboard.RevenueAnalytics != nil && dashboard.RevenueAnalytics.SubscriptionChurnRate > 5.0 {
		alerts = append(alerts, models.AnalyticsAlert{
			ID:          uuid.New(),
			Type:        "warning",
			Category:    "revenue",
			Message:     "Subscription churn rate is high",
			Metric:      "churn_rate",
			Threshold:   5.0,
			ActualValue: dashboard.RevenueAnalytics.SubscriptionChurnRate,
			CreatedAt:   time.Now(),
		})
	}

	return alerts
}

// GenerateRecommendations generates actionable recommendations based on analytics
func (s *AnalyticsService) GenerateRecommendations(dashboard *models.AnalyticsDashboard) []string {
	recommendations := []string{}

	if dashboard.ConsentAnalytics != nil {
		// Low consent rate recommendation
		if dashboard.ConsentAnalytics.ConsentRate < 70 {
			recommendations = append(recommendations,
				"Consider simplifying your consent forms to improve the consent rate. Current rate is below 70%.")
		}

		// High abandonment rate
		if dashboard.ConsentAnalytics.AbandonmentRate > 30 {
			recommendations = append(recommendations,
				"High abandonment rate detected. Review your consent flow for potential friction points.")
		}

		// Purpose optimization
		if dashboard.ConsentAnalytics.LeastConsentedPurpose != "" {
			recommendations = append(recommendations,
				"Purpose '"+dashboard.ConsentAnalytics.LeastConsentedPurpose+"' has the lowest consent rate. Consider reviewing its necessity or description.")
		}
	}

	if dashboard.DSRAnalytics != nil {
		// DSR response time
		if dashboard.DSRAnalytics.AvgResponseTime > 48 {
			recommendations = append(recommendations,
				"DSR response time exceeds 48 hours. Consider automating more of the DSR workflow.")
		}

		// Pending requests
		if dashboard.DSRAnalytics.PendingRequests > 10 {
			recommendations = append(recommendations,
				"You have "+fmt.Sprintf("%d", dashboard.DSRAnalytics.PendingRequests)+" pending DSR requests. Prioritize processing to maintain compliance.")
		}
	}

	if dashboard.UserEngagement != nil {
		// High bounce rate
		if dashboard.UserEngagement.BounceRate > 40 {
			recommendations = append(recommendations,
				"Bounce rate is above 40%. Improve page load times and user experience.")
		}

		// Mobile optimization
		if dashboard.UserEngagement.MobileUsers > dashboard.UserEngagement.DesktopUsers {
			recommendations = append(recommendations,
				"Majority of users are on mobile. Ensure mobile experience is optimized.")
		}
	}

	if dashboard.CookieAnalytics != nil {
		// Unknown cookies
		if dashboard.CookieAnalytics.UnknownCookies > 0 {
			recommendations = append(recommendations,
				"Found "+fmt.Sprintf("%d", dashboard.CookieAnalytics.UnknownCookies)+" unknown cookies. Review and categorize them for compliance.")
		}

		// Low marketing consent
		if rate, exists := dashboard.CookieAnalytics.CategoryConsentRates["marketing"]; exists && rate < 50 {
			recommendations = append(recommendations,
				"Marketing cookie consent rate is low. Consider explaining the benefits more clearly.")
		}
	}

	if dashboard.ComplianceAnalytics != nil {
		// High risk areas
		if dashboard.ComplianceAnalytics.HighRiskAreas > 0 {
			recommendations = append(recommendations,
				"Address "+fmt.Sprintf("%d", dashboard.ComplianceAnalytics.HighRiskAreas)+" high-risk compliance areas immediately.")
		}

		// Policy violations
		if dashboard.ComplianceAnalytics.PolicyViolations > 0 {
			recommendations = append(recommendations,
				"Review and address "+fmt.Sprintf("%d", dashboard.ComplianceAnalytics.PolicyViolations)+" policy violations.")
		}
	}

	if dashboard.PerformanceAnalytics != nil {
		// High CPU usage
		if dashboard.PerformanceAnalytics.CPUUsage > 80 {
			recommendations = append(recommendations,
				"CPU usage is above 80%. Consider scaling up your infrastructure.")
		}

		// Failed jobs
		if dashboard.PerformanceAnalytics.FailedJobs > 10 {
			recommendations = append(recommendations,
				"High number of failed background jobs. Investigate and fix the root cause.")
		}
	}

	if dashboard.RevenueAnalytics != nil {
		// High CAC
		if dashboard.RevenueAnalytics.CustomerAcquisitionCost > dashboard.RevenueAnalytics.AverageRevenuePerUser*3 {
			recommendations = append(recommendations,
				"Customer acquisition cost is too high relative to ARPU. Optimize marketing spend.")
		}

		// Churn rate
		if dashboard.RevenueAnalytics.SubscriptionChurnRate > 5 {
			recommendations = append(recommendations,
				"Subscription churn rate is above 5%. Implement retention strategies.")
		}
	}

	// Add general recommendations if no specific issues
	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"All metrics are within acceptable ranges. Continue monitoring for changes.")
		recommendations = append(recommendations,
			"Consider setting up automated alerts for key metrics.")
	}

	return recommendations
}

