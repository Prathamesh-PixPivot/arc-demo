package middleware

import (
	"net/http"
	"pixpivot/arc/internal/claims"
	contextKey "pixpivot/arc/internal/contextkeys"
)

// RequireSuperAdmin blocks any admin who is not a superadmin.
func RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(contextKey.FiduciaryClaimsKey).(*claims.FiduciaryClaims)
		if !ok || claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		if !claims.IsSuperAdmin {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"forbidden - superadmin only"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
