// Package errors defines project wide custom error definition
package errors

import (
	"errors"
	"fmt"
)

// Error codes for consistent error identification
const (
	CodeNotFound      = "NOT_FOUND"
	CodeAlreadyExists = "ALREADY_EXISTS"
	CodeInvalidInput  = "INVALID_INPUT"
	CodeInternal      = "INTERNAL_ERROR"
	CodePermission    = "PERMISSION_DENIED"
)

// BaseError contains common fields for all custom errors
type BaseError struct {
	Cause    error  // Underlying error
	Code     string // Machine-readable error code
	Message  string // Human-readable message
	Resource string // Resource type (e.g., "application", "team")
	ID       string // Resource identifier
}

// Error implements the error interface
func (e *BaseError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s: %s (resource=%s, id=%s)", e.Code, e.Message, e.Resource, e.ID)
	}
	if e.Resource != "" {
		return fmt.Sprintf("%s: %s (resource=%s)", e.Code, e.Message, e.Resource)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for error chain support
func (e *BaseError) Unwrap() error {
	return e.Cause
}

// NotFoundError indicates a resource was not found
type NotFoundError struct {
	BaseError
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		BaseError: BaseError{
			Code:     CodeNotFound,
			Message:  fmt.Sprintf("%s not found", resource),
			Resource: resource,
			ID:       id,
		},
	}
}

// NewNotFoundErrorWithCause creates a NotFoundError with an underlying cause
func NewNotFoundErrorWithCause(resource, id string, cause error) *NotFoundError {
	err := NewNotFoundError(resource, id)
	err.Cause = cause
	return err
}

// AlreadyExistsError indicates a resource already exists
type AlreadyExistsError struct {
	BaseError
}

// NewAlreadyExistsError creates a new AlreadyExistsError
func NewAlreadyExistsError(resource, id string) *AlreadyExistsError {
	return &AlreadyExistsError{
		BaseError: BaseError{
			Code:     CodeAlreadyExists,
			Message:  fmt.Sprintf("%s already exists", resource),
			Resource: resource,
			ID:       id,
		},
	}
}

// NewAlreadyExistsErrorWithCause creates an AlreadyExistsError with an underlying cause
func NewAlreadyExistsErrorWithCause(resource, id string, cause error) *AlreadyExistsError {
	err := NewAlreadyExistsError(resource, id)
	err.Cause = cause
	return err
}

// InvalidInputError indicates invalid input was provided
type InvalidInputError struct {
	BaseError
	Field string // Specific field that is invalid (optional)
}

// NewInvalidInputError creates a new InvalidInputError
func NewInvalidInputError(message string) *InvalidInputError {
	return &InvalidInputError{
		BaseError: BaseError{
			Code:    CodeInvalidInput,
			Message: message,
		},
	}
}

// NewInvalidInputErrorWithField creates an InvalidInputError for a specific field
func NewInvalidInputErrorWithField(field, message string) *InvalidInputError {
	return &InvalidInputError{
		BaseError: BaseError{
			Code:    CodeInvalidInput,
			Message: message,
		},
		Field: field,
	}
}

// NewInvalidInputErrorWithCause creates an InvalidInputError with an underlying cause
func NewInvalidInputErrorWithCause(message string, cause error) *InvalidInputError {
	err := NewInvalidInputError(message)
	err.Cause = cause
	return err
}

// Error implements the error interface with field information
func (e *InvalidInputError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field=%s)", e.Code, e.Message, e.Field)
	}
	return e.BaseError.Error()
}

// InternalError indicates an internal server error
type InternalError struct {
	BaseError
}

// NewInternalError creates a new InternalError
func NewInternalError(message string) *InternalError {
	return &InternalError{
		BaseError: BaseError{
			Code:    CodeInternal,
			Message: message,
		},
	}
}

// NewInternalErrorWithCause creates an InternalError with an underlying cause
func NewInternalErrorWithCause(message string, cause error) *InternalError {
	return &InternalError{
		BaseError: BaseError{
			Code:    CodeInternal,
			Message: message,
			Cause:   cause,
		},
	}
}

// PermissionError indicates a permission-related error
type PermissionError struct {
	BaseError
	Path string // File or directory path
}

// NewPermissionError creates a new PermissionError
func NewPermissionError(path, message string) *PermissionError {
	return &PermissionError{
		BaseError: BaseError{
			Code:    CodePermission,
			Message: message,
		},
		Path: path,
	}
}

// NewPermissionErrorWithCause creates a PermissionError with an underlying cause
func NewPermissionErrorWithCause(path, message string, cause error) *PermissionError {
	return &PermissionError{
		BaseError: BaseError{
			Code:    CodePermission,
			Message: message,
			Cause:   cause,
		},
		Path: path,
	}
}

// Error implements the error interface with path information
func (e *PermissionError) Error() string {
	return fmt.Sprintf("%s: %s (path=%s)", e.Code, e.Message, e.Path)
}

// Type checking helper functions

// IsNotFound checks if an error is a NotFoundError
func IsNotFound(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}

// IsAlreadyExists checks if an error is an AlreadyExistsError
func IsAlreadyExists(err error) bool {
	var alreadyExistsErr *AlreadyExistsError
	return errors.As(err, &alreadyExistsErr)
}

// IsInvalidInput checks if an error is an InvalidInputError
func IsInvalidInput(err error) bool {
	var invalidInputErr *InvalidInputError
	return errors.As(err, &invalidInputErr)
}

// IsInternal checks if an error is an InternalError
func IsInternal(err error) bool {
	var internalErr *InternalError
	return errors.As(err, &internalErr)
}

// IsPermissionError checks if an error is a PermissionError
func IsPermissionError(err error) bool {
	var permErr *PermissionError
	return errors.As(err, &permErr)
}

// GetErrorCode extracts the error code from a custom error, or returns INTERNAL_ERROR
func GetErrorCode(err error) string {
	var base *BaseError

	var notFound *NotFoundError
	if errors.As(err, &notFound) {
		return notFound.Code
	}

	var alreadyExists *AlreadyExistsError
	if errors.As(err, &alreadyExists) {
		return alreadyExists.Code
	}

	var invalidInput *InvalidInputError
	if errors.As(err, &invalidInput) {
		return invalidInput.Code
	}

	var internal *InternalError
	if errors.As(err, &internal) {
		return internal.Code
	}

	var permission *PermissionError
	if errors.As(err, &permission) {
		return permission.Code
	}

	if errors.As(err, &base) {
		return base.Code
	}

	return CodeInternal
}

// GetBaseError extracts the BaseError from any custom error type
func GetBaseError(err error) *BaseError {
	var notFound *NotFoundError
	if errors.As(err, &notFound) {
		return &notFound.BaseError
	}

	var alreadyExists *AlreadyExistsError
	if errors.As(err, &alreadyExists) {
		return &alreadyExists.BaseError
	}

	var invalidInput *InvalidInputError
	if errors.As(err, &invalidInput) {
		return &invalidInput.BaseError
	}

	var internal *InternalError
	if errors.As(err, &internal) {
		return &internal.BaseError
	}

	var permission *PermissionError
	if errors.As(err, &permission) {
		return &permission.BaseError
	}

	return nil
}
