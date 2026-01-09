// Package server provides HTTP API capabilities for Simplify
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AkMo3/simplify/internal/config"
	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/AkMo3/simplify/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server represents the HTTP API server
type Server struct {
	router    *chi.Mux
	server    *http.Server
	store     *store.Store
	container *container.Client
	config    *config.Config
}

// New creates a new Server with the provided dependencies
func New(cfg *config.Config, storeImpl *store.Store, containerClient *container.Client) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		store:     storeImpl,
		container: containerClient,
		config:    cfg,
	}
	s.setupMiddleware()
	s.setupRoutes()
	return s
}

// setupMiddleware configures the middleware stack
func (s *Server) setupMiddleware() {
	// Request ID for tracing
	s.router.Use(middleware.RequestID)

	// Real IP detection (for proxies)
	s.router.Use(middleware.RealIP)

	// Request logging
	s.router.Use(middleware.Logger)

	// Panic recovery
	s.router.Use(middleware.Recoverer)

	// Security headers
	s.router.Use(SecurityHeaders)

	// No cache for API responses
	s.router.Use(NoCacheHeaders)

	// JSON content type validation for POST/PUT/PATCH
	s.router.Use(RequireJSONContentType)
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// Health check endpoints (no auth required)
	s.router.Get("/healthz", s.handleHealthz)
	s.router.Get("/readyz", s.handleReadyz)

	// API routes
	s.router.Route("/api/v1", func(r chi.Router) {
		// Applications
		r.Post("/applications", WrapHandler(s.handleCreateApplication))
		r.Get("/applications", WrapHandler(s.handleListApplications))
		r.Get("/applications/{id}", WrapHandler(s.handleGetApplication))
		r.Put("/applications/{id}", WrapHandler(s.handleUpdateApplication))
		r.Delete("/applications/{id}", WrapHandler(s.handleDeleteApplication))

		// Teams
		r.Post("/teams", WrapHandler(s.handleCreateTeam))
		r.Get("/teams", WrapHandler(s.handleListTeams))
		r.Get("/teams/{id}", WrapHandler(s.handleGetTeam))
		r.Put("/teams/{id}", WrapHandler(s.handleUpdateTeam))
		r.Delete("/teams/{id}", WrapHandler(s.handleDeleteTeam))

		// Projects
		r.Post("/projects", WrapHandler(s.handleCreateProject))
		r.Get("/projects", WrapHandler(s.handleListProjects))
		r.Get("/projects/{id}", WrapHandler(s.handleGetProject))
		r.Put("/projects/{id}", WrapHandler(s.handleUpdateProject))
		r.Delete("/projects/{id}", WrapHandler(s.handleDeleteProject))

		// Environments
		r.Post("/environments", WrapHandler(s.handleCreateEnvironment))
		r.Get("/environments", WrapHandler(s.handleListEnvironments))
		r.Get("/environments/{id}", WrapHandler(s.handleGetEnvironment))
		r.Put("/environments/{id}", WrapHandler(s.handleUpdateEnvironment))
		r.Delete("/environments/{id}", WrapHandler(s.handleDeleteEnvironment))
	})
}

// Start starts the HTTP server and blocks until the context is canceled.
// It handles graceful shutdown when the context is canceled.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Server.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.Server.IdleTimeout) * time.Second,
	}

	// Channel to receive server errors
	errCh := make(chan error, 1)

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server",
			"addr", addr,
			"read_timeout", s.config.Server.ReadTimeout,
			"write_timeout", s.config.Server.WriteTimeout,
		)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		logger.Info("Shutting down HTTP server...")
		return s.shutdown()
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}
}

// shutdown performs a graceful shutdown of the server
func (s *Server) shutdown() error {
	shutdownTimeout := time.Duration(s.config.Server.ShutdownTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info("Graceful shutdown initiated",
		"timeout_seconds", s.config.Server.ShutdownTimeout,
	)

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Graceful shutdown failed", "error", err)
		return fmt.Errorf("shutdown error: %w", err)
	}

	logger.Info("HTTP server stopped gracefully")
	return nil
}

// Router returns the chi router for testing purposes
func (s *Server) Router() *chi.Mux {
	return s.router
}
