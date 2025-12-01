package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// writeJSON writes a JSON response with the given status code and data
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeError writes an error response with the given status code and message
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  statusCode,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("Error encoding error response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

/*
// writeValidationError writes a validation error response
func writeValidationError(w http.ResponseWriter, errors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	errorResponse := map[string]interface{}{
		"error":            true,
		"message":          "Validation failed",
		"validation_errors": errors,
		"status":           http.StatusBadRequest,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("Error encoding validation error response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
*/

// getClientIP extracts the client IP address from the request
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

// getUserAgent extracts the user agent from the request
func getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

/*
// getTenantIDFromHeader extracts tenant ID from request headers
func getTenantIDFromHeader(r *http.Request) string {
	return r.Header.Get("X-Tenant-ID")
}

// getLanguageFromHeader extracts language preference from request headers
func getLanguageFromHeader(r *http.Request) string {
	// Check Accept-Language header
	if lang := r.Header.Get("Accept-Language"); lang != "" {
		// Parse the first language preference
		if strings.Contains(lang, ",") {
			return strings.Split(lang, ",")[0]
		}
		return lang
	}

	// Check custom language header
	if lang := r.Header.Get("X-Language"); lang != "" {
		return lang
	}

	// Default to English
	return "en"
}
*/

// corsHeaders sets CORS headers for the response
func corsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-Language")
}

/*
// handlePreflight handles OPTIONS requests for CORS
func handlePreflight(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		corsHeaders(w)
		w.WriteHeader(http.StatusOK)
		return
	}
}

// logRequest logs the incoming request for debugging
func logRequest(r *http.Request) {
	log.Printf("[%s] %s %s - IP: %s, User-Agent: %s",
		r.Method, r.URL.Path, r.URL.RawQuery, getClientIP(r), getUserAgent(r))
}
*/

