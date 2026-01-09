package server

import (
	"encoding/json"
	"net/http"

	"github.com/AkMo3/simplify/internal/errors"
	"github.com/AkMo3/simplify/internal/logger"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the detailed error information
type ErrorDetail struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Resource string `json:"resource,omitempty"`
	ID       string `json:"id,omitempty"`
	Field    string `json:"field,omitempty"`
}

// AppHandler is a handler function that returns an error
// This allows centralized error handling via WrapHandler
type AppHandler func(w http.ResponseWriter, r *http.Request) error

// WrapHandler wraps an AppHandler and converts returned errors to HTTP responses
func WrapHandler(h AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		// Log the error with request context
		logger.Error("Request failed",
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
		)

		// Map error to HTTP response
		statusCode, response := mapErrorToResponse(err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
			logger.Error("Failed to encode error response", "error", encodeErr)
		}
	}
}

// mapErrorToResponse converts a custom error to HTTP status code and response body
func mapErrorToResponse(err error) (int, ErrorResponse) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    errors.CodeInternal,
			Message: "An internal error occurred",
		},
	}

	// Check for NotFoundError
	if errors.IsNotFound(err) {
		if base := errors.GetBaseError(err); base != nil {
			response.Error = ErrorDetail{
				Code:     base.Code,
				Message:  base.Message,
				Resource: base.Resource,
				ID:       base.ID,
			}
		}
		return http.StatusNotFound, response
	}

	// Check for AlreadyExistsError
	if errors.IsAlreadyExists(err) {
		if base := errors.GetBaseError(err); base != nil {
			response.Error = ErrorDetail{
				Code:     base.Code,
				Message:  base.Message,
				Resource: base.Resource,
				ID:       base.ID,
			}
		}
		return http.StatusConflict, response
	}

	// Check for InvalidInputError
	var invalidInputErr *errors.InvalidInputError
	if errors.IsInvalidInput(err) {
		if base := errors.GetBaseError(err); base != nil {
			response.Error = ErrorDetail{
				Code:    base.Code,
				Message: base.Message,
			}
		}
		// Add field if present
		if asErr, ok := err.(*errors.InvalidInputError); ok {
			response.Error.Field = asErr.Field
		} else if stdErr, ok := interface{}(err).(*errors.InvalidInputError); ok {
			response.Error.Field = stdErr.Field
		}
		_ = invalidInputErr // suppress unused warning
		return http.StatusBadRequest, response
	}

	// Check for PermissionError
	if errors.IsPermissionError(err) {
		if base := errors.GetBaseError(err); base != nil {
			response.Error = ErrorDetail{
				Code:    base.Code,
				Message: base.Message,
			}
		}
		return http.StatusForbidden, response
	}

	// Check for InternalError or unknown errors
	if errors.IsInternal(err) {
		if base := errors.GetBaseError(err); base != nil {
			response.Error = ErrorDetail{
				Code:    base.Code,
				Message: base.Message,
			}
		}
		return http.StatusInternalServerError, response
	}

	// Default: unknown error type, treat as internal error
	response.Error.Message = err.Error()
	return http.StatusInternalServerError, response
}

// writeJSON is a helper to write JSON responses
func writeJSON(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// writeSuccess writes a successful JSON response with status 200
func writeSuccess(w http.ResponseWriter, data any) error {
	return writeJSON(w, http.StatusOK, data)
}

// writeCreated writes a successful JSON response with status 201
func writeCreated(w http.ResponseWriter, data any) error {
	return writeJSON(w, http.StatusCreated, data)
}

// writeNoContent writes a successful response with status 204 (no body)
func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
