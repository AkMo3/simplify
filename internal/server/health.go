package server

import (
	"context"
	"net/http"
	"time"

	"github.com/AkMo3/simplify/internal/logger"
	"go.uber.org/zap"
)

// HealthStatus represents the health check response
type HealthStatus struct {
	Checks map[string]ComponentHealth `json:"checks,omitempty"`
	Status string                     `json:"status"`
}

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

const (
	statusHealthy   = "healthy"
	statusUnhealthy = "unhealthy"
)

// handleHealthz handles liveness probe requests.
// Returns 200 if the server is running (always healthy if we can respond).
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status: statusHealthy,
	}
	err := writeJSON(w, http.StatusOK, status)
	if err != nil {
		logger.Error("failed to write json", zap.Error(err))
	}
}

// handleReadyz handles readiness probe requests.
// Returns 200 only if all dependencies (database, Podman) are accessible.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]ComponentHealth)
	allHealthy := true

	// Check database connectivity
	dbHealth := s.checkDatabase()
	checks["database"] = dbHealth
	if dbHealth.Status != statusHealthy {
		allHealthy = false
	}

	// Check Podman connectivity
	podmanHealth := s.checkPodman(r.Context())
	checks["podman"] = podmanHealth
	if podmanHealth.Status != statusHealthy {
		allHealthy = false
	}

	status := HealthStatus{
		Status: statusHealthy,
		Checks: checks,
	}

	httpStatus := http.StatusOK
	if !allHealthy {
		status.Status = statusUnhealthy
		httpStatus = http.StatusServiceUnavailable
	}

	err := writeJSON(w, httpStatus, status)
	if err != nil {
		logger.Error("failed to write json", zap.Error(err))
	}
}

// checkDatabase verifies database connectivity
func (s *Server) checkDatabase() ComponentHealth {
	if s.store == nil {
		return ComponentHealth{
			Status:  statusUnhealthy,
			Message: "store not initialized",
		}
	}

	if err := s.store.Ping(); err != nil {
		return ComponentHealth{
			Status:  statusUnhealthy,
			Message: err.Error(),
		}
	}

	return ComponentHealth{
		Status: statusHealthy,
	}
}

// checkPodman verifies Podman socket connectivity
func (s *Server) checkPodman(ctx context.Context) ComponentHealth {
	if s.container == nil {
		return ComponentHealth{
			Status:  statusUnhealthy,
			Message: "container client not initialized",
		}
	}

	// Use a timeout for the health check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to list containers as a connectivity check
	_, err := s.container.List(checkCtx, false)
	if err != nil {
		return ComponentHealth{
			Status:  statusUnhealthy,
			Message: err.Error(),
		}
	}

	return ComponentHealth{
		Status: statusHealthy,
	}
}
