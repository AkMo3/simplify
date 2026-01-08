package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotFoundError(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		err := NewNotFoundError("application", "app-123")

		assert.Equal(t, CodeNotFound, err.Code)
		assert.Equal(t, "application", err.Resource)
		assert.Equal(t, "app-123", err.ID)
		assert.Contains(t, err.Error(), "NOT_FOUND")
		assert.Contains(t, err.Error(), "application")
		assert.Contains(t, err.Error(), "app-123")
	})

	t.Run("with cause", func(t *testing.T) {
		cause := fmt.Errorf("database error")
		err := NewNotFoundErrorWithCause("team", "team-1", cause)

		assert.Equal(t, cause, err.Cause)
		assert.True(t, errors.Is(err, cause))
	})

	t.Run("IsNotFound helper", func(t *testing.T) {
		err := NewNotFoundError("project", "proj-1")
		assert.True(t, IsNotFound(err))
		assert.False(t, IsAlreadyExists(err))
		assert.False(t, IsInvalidInput(err))
		assert.False(t, IsInternal(err))
	})
}

func TestAlreadyExistsError(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		err := NewAlreadyExistsError("application", "app-123")

		assert.Equal(t, CodeAlreadyExists, err.Code)
		assert.Equal(t, "application", err.Resource)
		assert.Equal(t, "app-123", err.ID)
		assert.Contains(t, err.Error(), "ALREADY_EXISTS")
	})

	t.Run("with cause", func(t *testing.T) {
		cause := fmt.Errorf("unique constraint violation")
		err := NewAlreadyExistsErrorWithCause("team", "team-1", cause)

		assert.Equal(t, cause, err.Cause)
		assert.True(t, errors.Is(err, cause))
	})

	t.Run("IsAlreadyExists helper", func(t *testing.T) {
		err := NewAlreadyExistsError("project", "proj-1")
		assert.True(t, IsAlreadyExists(err))
		assert.False(t, IsNotFound(err))
	})
}

func TestInvalidInputError(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		err := NewInvalidInputError("name is required")

		assert.Equal(t, CodeInvalidInput, err.Code)
		assert.Equal(t, "name is required", err.Message)
		assert.Contains(t, err.Error(), "INVALID_INPUT")
	})

	t.Run("with field", func(t *testing.T) {
		err := NewInvalidInputErrorWithField("email", "invalid email format")

		assert.Equal(t, "email", err.Field)
		assert.Contains(t, err.Error(), "field=email")
	})

	t.Run("with cause", func(t *testing.T) {
		cause := fmt.Errorf("json unmarshal error")
		err := NewInvalidInputErrorWithCause("invalid request body", cause)

		assert.Equal(t, cause, err.Cause)
		assert.True(t, errors.Is(err, cause))
	})

	t.Run("IsInvalidInput helper", func(t *testing.T) {
		err := NewInvalidInputError("bad input")
		assert.True(t, IsInvalidInput(err))
		assert.False(t, IsNotFound(err))
	})
}

func TestInternalError(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		err := NewInternalError("unexpected error occurred")

		assert.Equal(t, CodeInternal, err.Code)
		assert.Contains(t, err.Error(), "INTERNAL_ERROR")
	})

	t.Run("with cause", func(t *testing.T) {
		cause := fmt.Errorf("database connection failed")
		err := NewInternalErrorWithCause("failed to process request", cause)

		assert.Equal(t, cause, err.Cause)
		assert.True(t, errors.Is(err, cause))
	})

	t.Run("IsInternal helper", func(t *testing.T) {
		err := NewInternalError("internal error")
		assert.True(t, IsInternal(err))
		assert.False(t, IsNotFound(err))
	})
}

func TestPermissionError(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		err := NewPermissionError("/var/lib/simplify", "cannot write to directory")

		assert.Equal(t, CodePermission, err.Code)
		assert.Equal(t, "/var/lib/simplify", err.Path)
		assert.Contains(t, err.Error(), "PERMISSION_DENIED")
		assert.Contains(t, err.Error(), "/var/lib/simplify")
	})

	t.Run("with cause", func(t *testing.T) {
		cause := fmt.Errorf("operation not permitted")
		err := NewPermissionErrorWithCause("/etc/simplify", "cannot create directory", cause)

		assert.Equal(t, cause, err.Cause)
		assert.True(t, errors.Is(err, cause))
	})

	t.Run("IsPermissionError helper", func(t *testing.T) {
		err := NewPermissionError("/path", "no access")
		assert.True(t, IsPermissionError(err))
		assert.False(t, IsNotFound(err))
	})
}

func TestErrorUnwrapping(t *testing.T) {
	t.Run("unwrap chain", func(t *testing.T) {
		rootCause := fmt.Errorf("root cause")
		err := NewInternalErrorWithCause("wrapper", rootCause)

		// errors.Unwrap should return the cause
		unwrapped := errors.Unwrap(err)
		assert.Equal(t, rootCause, unwrapped)
	})

	t.Run("errors.Is with wrapped error", func(t *testing.T) {
		sentinel := fmt.Errorf("sentinel error")
		err := NewNotFoundErrorWithCause("resource", "id", sentinel)

		assert.True(t, errors.Is(err, sentinel))
	})

	t.Run("errors.As with custom type", func(t *testing.T) {
		err := NewNotFoundError("application", "app-1")

		var notFoundErr *NotFoundError
		require.True(t, errors.As(err, &notFoundErr))
		assert.Equal(t, "application", notFoundErr.Resource)
		assert.Equal(t, "app-1", notFoundErr.ID)
	})
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "NotFoundError",
			err:      NewNotFoundError("app", "1"),
			expected: CodeNotFound,
		},
		{
			name:     "AlreadyExistsError",
			err:      NewAlreadyExistsError("app", "1"),
			expected: CodeAlreadyExists,
		},
		{
			name:     "InvalidInputError",
			err:      NewInvalidInputError("bad"),
			expected: CodeInvalidInput,
		},
		{
			name:     "InternalError",
			err:      NewInternalError("failed"),
			expected: CodeInternal,
		},
		{
			name:     "PermissionError",
			err:      NewPermissionError("/path", "denied"),
			expected: CodePermission,
		},
		{
			name:     "unknown error returns internal",
			err:      fmt.Errorf("some random error"),
			expected: CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetErrorCode(tt.err)
			assert.Equal(t, tt.expected, code)
		})
	}
}

func TestGetBaseError(t *testing.T) {
	t.Run("from NotFoundError", func(t *testing.T) {
		err := NewNotFoundError("app", "123")
		base := GetBaseError(err)

		require.NotNil(t, base)
		assert.Equal(t, CodeNotFound, base.Code)
		assert.Equal(t, "app", base.Resource)
	})

	t.Run("from unknown error returns nil", func(t *testing.T) {
		err := fmt.Errorf("random error")
		base := GetBaseError(err)

		assert.Nil(t, base)
	})
}

func TestErrorMessages(t *testing.T) {
	t.Run("NotFoundError with all fields", func(t *testing.T) {
		err := NewNotFoundError("application", "app-123")
		msg := err.Error()

		assert.Equal(t, "NOT_FOUND: application not found (resource=application, id=app-123)", msg)
	})

	t.Run("BaseError without ID", func(t *testing.T) {
		err := &InternalError{
			BaseError: BaseError{
				Code:     CodeInternal,
				Message:  "database unavailable",
				Resource: "database",
			},
		}
		msg := err.Error()

		assert.Equal(t, "INTERNAL_ERROR: database unavailable (resource=database)", msg)
	})

	t.Run("BaseError without resource", func(t *testing.T) {
		err := NewInternalError("unexpected error")
		msg := err.Error()

		assert.Equal(t, "INTERNAL_ERROR: unexpected error", msg)
	})

	t.Run("InvalidInputError with field", func(t *testing.T) {
		err := NewInvalidInputErrorWithField("port", "must be between 1 and 65535")
		msg := err.Error()

		assert.Equal(t, "INVALID_INPUT: must be between 1 and 65535 (field=port)", msg)
	})

	t.Run("PermissionError with path", func(t *testing.T) {
		err := NewPermissionError("/var/lib/simplify", "cannot create directory")
		msg := err.Error()

		assert.Equal(t, "PERMISSION_DENIED: cannot create directory (path=/var/lib/simplify)", msg)
	})
}
