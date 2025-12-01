package middleware

import (
	"net/http"
	"pixpivot/arc/internal/claims"
	"pixpivot/arc/internal/contextkeys"
	"pixpivot/arc/internal/licensing"
)

type LicenseMiddleware struct {
	manager *licensing.LicenseManager
	tracker licensing.UsageTracker
}

func NewLicenseMiddleware(manager *licensing.LicenseManager, tracker licensing.UsageTracker) *LicenseMiddleware {
	return &LicenseMiddleware{
		manager: manager,
		tracker: tracker,
	}
}

func (m *LicenseMiddleware) EnforceLicense(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Check if license is valid
		license, err := m.manager.GetLicense()
		if err != nil {
			// If no license, block everything except admin/license endpoints (handled by routing)
			// For now, return 503 Service Unavailable or 402 Payment Required
			http.Error(w, "License required: "+err.Error(), http.StatusPaymentRequired)
			return
		}

		if license.IsExpired() {
			http.Error(w, "License expired", http.StatusPaymentRequired)
			return
		}

		// 2. SaaS Enforcement
		if license.Type == licensing.LicenseTypeSaaS {
			// Extract Tenant ID
			tenantID := "global" // Default if not authenticated
			if dpClaims := GetClaimsFromContext(r.Context()); dpClaims != nil {
				tenantID = dpClaims.TenantID
			} else if fClaims, ok := r.Context().Value(contextkeys.FiduciaryClaimsKey).(*claims.FiduciaryClaims); ok {
				tenantID = fClaims.TenantID
			}

			// Increment usage
			count, err := m.tracker.IncrementAPIRequest(tenantID)
			if err != nil {
				// Log error but maybe allow request? Or fail closed?
				// Fail open for Redis errors to avoid downtime
				next.ServeHTTP(w, r)
				return
			}

			// Check limit
			if err := m.manager.CheckSaaSLimit("api_requests", count); err != nil {
				http.Error(w, "API rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
