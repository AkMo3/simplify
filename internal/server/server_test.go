package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/AkMo3/simplify/internal/config"
	"github.com/AkMo3/simplify/internal/core"
	"github.com/AkMo3/simplify/internal/errors"
	"github.com/AkMo3/simplify/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a test server with a temporary database
func setupTestServer(t *testing.T) (srv *Server, cleanup func()) {
	t.Helper()

	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "simplify-server-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Create store
	s, err := store.New(dbPath)
	require.NoError(t, err)

	// Create config with defaults
	cfg := &config.Config{
		Env: config.EnvDevelopment,
		Server: config.ServerConfig{
			Port:            8080,
			ReadTimeout:     30,
			WriteTimeout:    30,
			IdleTimeout:     120,
			ShutdownTimeout: 30,
		},
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
	}

	// Create server without container client (nil is ok for most tests)
	srv = New(cfg, s, nil)

	cleanup = func() {
		s.Close()
		os.RemoveAll(tmpDir)
	}

	return srv, cleanup
}

// =============================================================================
// Application Handler Tests
// =============================================================================

func TestCreateApplication(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		body           map[string]any
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "valid application",
			body: map[string]any{
				"name":  "test-app",
				"image": "nginx:latest",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var app core.Application
				err := json.Unmarshal(body, &app)
				require.NoError(t, err)
				assert.Equal(t, "test-app", app.Name)
				assert.Equal(t, "nginx:latest", app.Image)
				assert.NotEmpty(t, app.ID)
				assert.False(t, app.CreatedAt.IsZero())
			},
		},
		{
			name: "missing name",
			body: map[string]any{
				"image": "nginx:latest",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var errResp ErrorResponse
				err := json.Unmarshal(body, &errResp)
				require.NoError(t, err)
				assert.Equal(t, errors.CodeInvalidInput, errResp.Error.Code)
				assert.Equal(t, "name", errResp.Error.Field)
			},
		},
		{
			name: "missing image",
			body: map[string]any{
				"name": "test-app",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var errResp ErrorResponse
				err := json.Unmarshal(body, &errResp)
				require.NoError(t, err)
				assert.Equal(t, errors.CodeInvalidInput, errResp.Error.Code)
				assert.Equal(t, "image", errResp.Error.Field)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.body)
			assert.Nil(t, err, "expected errror to be nil while marshaling body")
			req := httptest.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			srv.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestGetApplication(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// Create an application first
	app := &core.Application{
		ID:    "test-app-1",
		Name:  "Test App",
		Image: "nginx:latest",
	}
	err := srv.store.CreateApplication(app)
	require.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "existing application",
			id:             "test-app-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent application",
			id:             "non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/applications/"+tt.id, http.NoBody)
			w := httptest.NewRecorder()
			srv.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestListApplications(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// Initially empty
	req := httptest.NewRequest(http.MethodGet, "/api/v1/applications", http.NoBody)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var apps []core.Application
	err := json.Unmarshal(w.Body.Bytes(), &apps)
	require.NoError(t, err)
	assert.Empty(t, apps)

	// Add some applications
	for i := range 3 {
		app := &core.Application{
			ID:    "app-" + string(rune('a'+i)),
			Name:  "App " + string(rune('A'+i)),
			Image: "nginx:latest",
		}
		err := srv.store.CreateApplication(app)
		require.NoError(t, err)
	}

	// List again
	req = httptest.NewRequest(http.MethodGet, "/api/v1/applications", http.NoBody)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &apps)
	require.NoError(t, err)
	assert.Len(t, apps, 3)
}

func TestUpdateApplication(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// Create an application first
	app := &core.Application{
		ID:    "test-app-update",
		Name:  "Original Name",
		Image: "nginx:1.0",
	}
	err := srv.store.CreateApplication(app)
	require.NoError(t, err)

	// Update it
	updateBody := map[string]any{
		"name":  "Updated Name",
		"image": "nginx:2.0",
	}
	body, err := json.Marshal(updateBody)
	assert.Nil(t, err, "expected no error while marshaling update body")
	req := httptest.NewRequest(http.MethodPut, "/api/v1/applications/test-app-update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated core.Application
	err = json.Unmarshal(w.Body.Bytes(), &updated)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "nginx:2.0", updated.Image)
}

func TestUpdateApplicationNotFound(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	updateBody := map[string]any{
		"name":  "Updated Name",
		"image": "nginx:latest",
	}
	body, err := json.Marshal(updateBody)
	assert.Nil(t, err, "expected no error while marshaling update body")
	req := httptest.NewRequest(http.MethodPut, "/api/v1/applications/non-existent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeNotFound, errResp.Error.Code)
}

func TestDeleteApplication(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// Create an application first
	app := &core.Application{
		ID:    "test-app-delete",
		Name:  "To Delete",
		Image: "nginx:latest",
	}
	err := srv.store.CreateApplication(app)
	require.NoError(t, err)

	// Delete it
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/applications/test-app-delete", http.NoBody)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify it's gone
	_, err = srv.store.GetApplication("test-app-delete")
	assert.True(t, errors.IsNotFound(err))
}

// =============================================================================
// Health Check Tests
// =============================================================================

func TestHealthzEndpoint(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var status HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &status)
	require.NoError(t, err)
	assert.Equal(t, "healthy", status.Status)
}

func TestReadyzEndpoint(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/readyz", http.NoBody)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	// Should be unhealthy because container client is nil
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var status HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &status)
	require.NoError(t, err)
	assert.Equal(t, "unhealthy", status.Status)
	assert.Contains(t, status.Checks, "database")
	assert.Contains(t, status.Checks, "podman")
	assert.Equal(t, "healthy", status.Checks["database"].Status)
	assert.Equal(t, "unhealthy", status.Checks["podman"].Status)
}

// =============================================================================
// Middleware Tests
// =============================================================================

func TestRequireJSONContentType(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "valid JSON content type",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest, // Bad request due to empty body, but passes content type check
		},
		{
			name:           "JSON with charset",
			contentType:    "application/json; charset=utf-8",
			expectedStatus: http.StatusBadRequest, // Bad request due to empty body
		},
		{
			name:           "missing content type",
			contentType:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "wrong content type",
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewReader([]byte("{}")))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			srv.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
}

func TestNoCacheHeaders(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, "no-store, no-cache, must-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))
}

// =============================================================================
// Error Mapping Tests
// =============================================================================

func TestMapErrorToResponse(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "not found error",
			err:            errors.NewNotFoundError("applications", "test-id"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   errors.CodeNotFound,
		},
		{
			name:           "already exists error",
			err:            errors.NewAlreadyExistsError("applications", "test-id"),
			expectedStatus: http.StatusConflict,
			expectedCode:   errors.CodeAlreadyExists,
		},
		{
			name:           "invalid input error",
			err:            errors.NewInvalidInputError("invalid data"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   errors.CodeInvalidInput,
		},
		{
			name:           "internal error",
			err:            errors.NewInternalError("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   errors.CodeInternal,
		},
		{
			name:           "permission error",
			err:            errors.NewPermissionError("access denied"),
			expectedStatus: http.StatusForbidden,
			expectedCode:   errors.CodePermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, response := mapErrorToResponse(tt.err)
			assert.Equal(t, tt.expectedStatus, status)
			assert.Equal(t, tt.expectedCode, response.Error.Code)
		})
	}
}

// =============================================================================
// Team Handler Tests (Basic coverage)
// =============================================================================

func TestTeamCRUD(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// Create
	createBody := map[string]any{"name": "Engineering", "slug": "eng"}
	body, err := json.Marshal(createBody)
	assert.Nil(t, err, "expected no error while marshaling create body")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var team core.Team
	err = json.Unmarshal(w.Body.Bytes(), &team)
	assert.Nil(t, err, "expected no error while unmarshaling body")
	teamID := team.ID

	// Get
	req = httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID, http.NoBody)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// List
	req = httptest.NewRequest(http.MethodGet, "/api/v1/teams", http.NoBody)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update
	updateBody := map[string]any{"name": "Platform Engineering", "slug": "platform"}
	body, err = json.Marshal(updateBody)
	assert.Nil(t, err, "expected no error while marshaling update body")
	req = httptest.NewRequest(http.MethodPut, "/api/v1/teams/"+teamID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/teams/"+teamID, http.NoBody)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

// =============================================================================
// Graceful Shutdown Test
// =============================================================================

func TestServerGracefulShutdown(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Give server time to start
	// In real tests, you'd wait for the server to be ready

	// Cancel context to trigger shutdown
	cancel()

	// Server should shut down gracefully
	err := <-errCh
	assert.NoError(t, err)
}
