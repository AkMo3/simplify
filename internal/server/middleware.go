package server

import (
	"net/http"
	"strings"

	"github.com/AkMo3/simplify/internal/logger"
	"go.uber.org/zap"
)

// JSONContentType sets the Content-Type header to application/json for responses
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// RequireJSONContentType validates that POST/PUT/PATCH requests have JSON content type
func RequireJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check methods that typically have request bodies
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if contentType == "" {
				err := writeJSON(w, http.StatusBadRequest, ErrorResponse{
					Error: ErrorDetail{
						Code:    "INVALID_INPUT",
						Message: "Content-Type header is required",
						Field:   "Content-Type",
					},
				})
				if err != nil {
					logger.Error("failed to write json", zap.Error(err))
				}
				return
			}

			// Check if content type is JSON (handle charset and other params)
			if !strings.HasPrefix(contentType, "application/json") {
				err := writeJSON(w, http.StatusUnsupportedMediaType, ErrorResponse{
					Error: ErrorDetail{
						Code:    "INVALID_INPUT",
						Message: "Content-Type must be application/json",
						Field:   "Content-Type",
					},
				})
				if err != nil {
					logger.Error("failed to write json", zap.Error(err))
				}
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// NoCacheHeaders adds headers to prevent caching of API responses
func NoCacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders adds common security headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		next.ServeHTTP(w, r)
	})
}
