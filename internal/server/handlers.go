package server

import (
	"encoding/json"
	"net/http"

	"github.com/AkMo3/simplify/internal/core"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// handleCreateApplication accepts a JSON body and creates a new application
func (s *Server) handleCreateApplication(w http.ResponseWriter, r *http.Request) {
	var app core.Application
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation: ensure ID is present
	if app.ID == "" {
		http.Error(w, "Application ID is required", http.StatusBadRequest)
		return
	}

	if err := s.store.CreateApplication(&app); err != nil {
		logger.Error("Failed to create application", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(app); err != nil {
		logger.Error("Failed to encode response", zap.Error(err))
	}
}

// handleListApplications returns all applications
func (s *Server) handleListApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := s.store.ListApplications()
	if err != nil {
		logger.Error("Failed to list applications", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return empty list instead of null if empty
	if apps == nil {
		apps = []core.Application{}
	}

	if err := json.NewEncoder(w).Encode(apps); err != nil {
		logger.Error("Failed to encode response", zap.Error(err))
	}
}

// handleGetApplication returns a single application by ID
func (s *Server) handleGetApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	app, err := s.store.GetApplication(id)
	if err != nil {
		// In a real app, check if err is "NotFound" vs "InternalError"
		// For now, assume error means not found if it comes from BoltDB Get
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(app); err != nil {
		logger.Error("Failed to encode response", zap.Error(err))
	}
}

// handleDeleteApplication removes an application
func (s *Server) handleDeleteApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.store.DeleteApplication(id); err != nil {
		logger.Error("Failed to delete application", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
