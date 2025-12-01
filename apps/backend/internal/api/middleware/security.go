package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"pixpivot/arc/internal/models"
)

// SecurityMiddleware provides comprehensive security features
type SecurityMiddleware struct {
	db        *gorm.DB
	jwtSecret []byte
	hmacKey   []byte
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(db *gorm.DB, jwtSecret, hmacKey string) *SecurityMiddleware {
	return &SecurityMiddleware{
		db:        db,
		jwtSecret: []byte(jwtSecret),
		hmacKey:   []byte(hmacKey),
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self' wss: ws:; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"

		w.Header().Set("Content-Security-Policy", csp)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS with strict origin checking
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Allowed origins (should be configurable)
		allowedOrigins := []string{
			"http://localhost:3000",
			"https://consent-manager.com",
			"https://www.consent-manager.com",
		}

		// Check if origin is allowed
		originAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				originAllowed = true
				break
			}
		}

		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-API-Key, X-Platform, X-Signature, X-Timestamp")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates JWT tokens with OAuth 2.0 support
func (sm *SecurityMiddleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return sm.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Extract user information
		userID, _ := uuid.Parse(claims["sub"].(string))
		tenantID, _ := uuid.Parse(claims["tenant_id"].(string))
		platform := claims["platform"].(string)
		roles := claims["roles"].([]interface{})

		// Validate session in database
		var session models.PlatformSession
		err = sm.db.Where("user_id = ? AND platform = ? AND expires_at > ?", 
			userID, platform, time.Now()).First(&session).Error
		if err != nil {
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}

		// Add user context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)
		ctx = context.WithValue(ctx, "platform", platform)
		ctx = context.WithValue(ctx, "roles", roles)
		ctx = context.WithValue(ctx, "session_id", session.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// APIKeyMiddleware validates API keys for integration endpoints
func (sm *SecurityMiddleware) APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}

		// Hash the API key for lookup
		hasher := sha256.New()
		hasher.Write([]byte(apiKey))
		keyHash := hex.EncodeToString(hasher.Sum(nil))

		// Validate API key in database
		var apiKeyRecord models.APIKey
		err := sm.db.Where("key_hash = ? AND (expires_at IS NULL OR expires_at > ?)", 
			keyHash, time.Now()).First(&apiKeyRecord).Error
		if err != nil {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// Update last used timestamp
		sm.db.Model(&apiKeyRecord).Update("last_used_at", time.Now())

		// Add API key context
		ctx := context.WithValue(r.Context(), "api_key_id", apiKeyRecord.KeyID)
		ctx = context.WithValue(ctx, "tenant_id", apiKeyRecord.TenantID)
		ctx = context.WithValue(ctx, "api_scopes", apiKeyRecord.Scopes)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminRoleMiddleware ensures user has admin role
func AdminRoleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roles, ok := r.Context().Value("roles").([]interface{})
		if !ok {
			http.Error(w, "No roles found", http.StatusForbidden)
			return
		}

		hasAdminRole := false
		for _, role := range roles {
			if roleStr, ok := role.(string); ok && (roleStr == "admin" || roleStr == "super_admin") {
				hasAdminRole = true
				break
			}
		}

		if !hasAdminRole {
			http.Error(w, "Admin role required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// PlatformMiddleware ensures request is from correct platform
func PlatformMiddleware(allowedPlatforms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			platform := r.Header.Get("X-Platform")
			if platform == "" {
				// Try to get from context (for authenticated requests)
				if ctxPlatform := r.Context().Value("platform"); ctxPlatform != nil {
					platform = ctxPlatform.(string)
				}
			}

			if platform == "" {
				http.Error(w, "Platform header required", http.StatusBadRequest)
				return
			}

			allowed := false
			for _, allowedPlatform := range allowedPlatforms {
				if platform == allowedPlatform {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Platform not allowed for this endpoint", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestSigningMiddleware validates HMAC signatures for critical operations
func (sm *SecurityMiddleware) RequestSigningMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("X-Signature")
		timestamp := r.Header.Get("X-Timestamp")

		if signature == "" || timestamp == "" {
			http.Error(w, "Request signature required", http.StatusBadRequest)
			return
		}

		// Parse timestamp
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			http.Error(w, "Invalid timestamp", http.StatusBadRequest)
			return
		}

		// Check timestamp is within 5 minutes
		now := time.Now().Unix()
		if abs(now-ts) > 300 {
			http.Error(w, "Request timestamp too old", http.StatusBadRequest)
			return
		}

		// Read request body for signature verification
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)

		// Create signature payload
		payload := fmt.Sprintf("%s\n%s\n%s\n%s", r.Method, r.URL.Path, timestamp, string(body))

		// Calculate expected signature
		mac := hmac.New(sha256.New, sm.hmacKey)
		mac.Write([]byte(payload))
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		// Compare signatures
		if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) != 1 {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// TenantIsolationMiddleware ensures tenant isolation
func TenantIsolationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tenantID uuid.UUID

		// Try to get tenant ID from various sources
		if tid := r.Header.Get("X-Tenant-ID"); tid != "" {
			if parsed, err := uuid.Parse(tid); err == nil {
				tenantID = parsed
			}
		}

		// Try from context (authenticated users)
		if tenantID == uuid.Nil {
			if ctxTenantID := r.Context().Value("tenant_id"); ctxTenantID != nil {
				tenantID = ctxTenantID.(uuid.UUID)
			}
		}

		// Try from subdomain
		if tenantID == uuid.Nil {
			host := r.Host
			if strings.Contains(host, ".") {
				subdomain := strings.Split(host, ".")[0]
				// Look up tenant by subdomain (implement as needed)
				_ = subdomain
			}
		}

		// Try from path parameter
		if tenantID == uuid.Nil {
			vars := mux.Vars(r)
			if tid := vars["tenantId"]; tid != "" {
				if parsed, err := uuid.Parse(tid); err == nil {
					tenantID = parsed
				}
			}
		}

		if tenantID == uuid.Nil {
			http.Error(w, "Tenant ID required", http.StatusBadRequest)
			return
		}

		// Add tenant ID to context
		ctx := context.WithValue(r.Context(), "tenant_id", tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimitMiddleware implements rate limiting per IP/user
func RateLimitMiddleware(next http.Handler) http.Handler {
	// Simple in-memory rate limiter (use Redis in production)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := getClientIP(r)
		
		// Check rate limit (implement with Redis/memory store)
		// For now, just pass through
		_ = clientIP

		next.ServeHTTP(w, r)
	})
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

