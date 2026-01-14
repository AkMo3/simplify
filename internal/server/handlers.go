package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/core"
	"github.com/AkMo3/simplify/internal/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// =============================================================================
// Application Handlers
// =============================================================================

// handleCreateApplication creates a new application
func (s *Server) handleCreateApplication(w http.ResponseWriter, r *http.Request) error {
	var app core.Application
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		return errors.NewInvalidInputErrorWithCause("invalid request body", err)
	}

	// Generate ID if not provided
	if app.ID == "" {
		app.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now().UTC()
	app.CreatedAt = now
	app.UpdatedAt = now

	// Validate required fields
	if app.Name == "" {
		return errors.NewInvalidInputErrorWithField("name", "name is required")
	}
	if app.Image == "" {
		return errors.NewInvalidInputErrorWithField("image", "image is required")
	}

	if err := s.store.CreateApplication(&app); err != nil {
		return err
	}

	return writeCreated(w, app)
}

// handleListApplications returns all applications
func (s *Server) handleListApplications(w http.ResponseWriter, r *http.Request) error {
	apps, err := s.store.ListApplications()
	if err != nil {
		return err
	}

	// Fetch from podman the status of containers
	containers, err := s.container.List(r.Context(), true)
	if err != nil {
		// Log error but return apps with unknown status is better than failing?
		// For now, let's return error as before but we could improve this.
		return err
	}

	// Create a map of AppID -> ContainerInfo
	containerMap := make(map[string]container.ContainerInfo)
	for _, c := range containers {
		if appID, ok := c.Labels["simplify.app.id"]; ok {
			containerMap[appID] = c
		}
	}

	for i := range apps {
		if info, ok := containerMap[apps[i].ID]; ok {
			apps[i].Status = info.Status
			apps[i].Ports = info.Ports
		} else {
			apps[i].Status = "stopped" // Or "unknown" or empty
		}
	}

	// Return empty array instead of null
	if apps == nil {
		apps = []core.Application{}
	}

	return writeSuccess(w, apps)
}

// handleGetApplication returns a single application by ID
func (s *Server) handleGetApplication(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	app, err := s.store.GetApplication(id)
	if err != nil {
		return err
	}

	return writeSuccess(w, app)
}

// handleUpdateApplication updates an existing application
func (s *Server) handleUpdateApplication(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	var app core.Application
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		return errors.NewInvalidInputErrorWithCause("invalid request body", err)
	}

	// Ensure ID matches URL
	app.ID = id
	app.UpdatedAt = time.Now().UTC()

	// Validate required fields
	if app.Name == "" {
		return errors.NewInvalidInputErrorWithField("name", "name is required")
	}
	if app.Image == "" {
		return errors.NewInvalidInputErrorWithField("image", "image is required")
	}

	if err := s.store.UpdateApplication(&app); err != nil {
		return err
	}

	return writeSuccess(w, app)
}

// handleDeleteApplication removes an application
func (s *Server) handleDeleteApplication(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	if err := s.store.DeleteApplication(id); err != nil {
		return err
	}

	writeNoContent(w)
	return nil
}

// =============================================================================
// Team Handlers
// =============================================================================

// handleCreateTeam creates a new team
func (s *Server) handleCreateTeam(w http.ResponseWriter, r *http.Request) error {
	var team core.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		return errors.NewInvalidInputErrorWithCause("invalid request body", err)
	}

	if team.ID == "" {
		team.ID = uuid.New().String()
	}
	team.CreatedAt = time.Now().UTC()

	if team.Name == "" {
		return errors.NewInvalidInputErrorWithField("name", "name is required")
	}

	if err := s.store.CreateTeam(&team); err != nil {
		return err
	}

	return writeCreated(w, team)
}

// handleListTeams returns all teams
func (s *Server) handleListTeams(w http.ResponseWriter, r *http.Request) error {
	teams, err := s.store.ListTeams()
	if err != nil {
		return err
	}

	if teams == nil {
		teams = []core.Team{}
	}

	return writeSuccess(w, teams)
}

// handleGetTeam returns a single team by ID
func (s *Server) handleGetTeam(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	team, err := s.store.GetTeam(id)
	if err != nil {
		return err
	}

	return writeSuccess(w, team)
}

// handleUpdateTeam updates an existing team
func (s *Server) handleUpdateTeam(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	var team core.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		return errors.NewInvalidInputErrorWithCause("invalid request body", err)
	}

	team.ID = id

	if team.Name == "" {
		return errors.NewInvalidInputErrorWithField("name", "name is required")
	}

	if err := s.store.UpdateTeam(&team); err != nil {
		return err
	}

	return writeSuccess(w, team)
}

// handleDeleteTeam removes a team
func (s *Server) handleDeleteTeam(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	if err := s.store.DeleteTeam(id); err != nil {
		return err
	}

	writeNoContent(w)
	return nil
}

// =============================================================================
// Project Handlers
// =============================================================================

// handleCreateProject creates a new project
func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) error {
	var project core.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		return errors.NewInvalidInputErrorWithCause("invalid request body", err)
	}

	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	project.CreatedAt = time.Now().UTC()

	if project.Name == "" {
		return errors.NewInvalidInputErrorWithField("name", "name is required")
	}

	if err := s.store.CreateProject(&project); err != nil {
		return err
	}

	return writeCreated(w, project)
}

// handleListProjects returns all projects
func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) error {
	projects, err := s.store.ListProjects()
	if err != nil {
		return err
	}

	if projects == nil {
		projects = []core.Project{}
	}

	return writeSuccess(w, projects)
}

// handleGetProject returns a single project by ID
func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	project, err := s.store.GetProject(id)
	if err != nil {
		return err
	}

	return writeSuccess(w, project)
}

// handleUpdateProject updates an existing project
func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	var project core.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		return errors.NewInvalidInputErrorWithCause("invalid request body", err)
	}

	project.ID = id

	if project.Name == "" {
		return errors.NewInvalidInputErrorWithField("name", "name is required")
	}

	if err := s.store.UpdateProject(&project); err != nil {
		return err
	}

	return writeSuccess(w, project)
}

// handleDeleteProject removes a project
func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	if err := s.store.DeleteProject(id); err != nil {
		return err
	}

	writeNoContent(w)
	return nil
}

// =============================================================================
// Environment Handlers
// =============================================================================

// handleCreateEnvironment creates a new environment
func (s *Server) handleCreateEnvironment(w http.ResponseWriter, r *http.Request) error {
	var env core.Environment
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		return errors.NewInvalidInputErrorWithCause("invalid request body", err)
	}

	if env.ID == "" {
		env.ID = uuid.New().String()
	}
	env.CreatedAt = time.Now().UTC()

	if env.Name == "" {
		return errors.NewInvalidInputErrorWithField("name", "name is required")
	}

	if err := s.store.CreateEnvironment(&env); err != nil {
		return err
	}

	return writeCreated(w, env)
}

// handleListEnvironments returns all environments
func (s *Server) handleListEnvironments(w http.ResponseWriter, r *http.Request) error {
	envs, err := s.store.ListEnvironments()
	if err != nil {
		return err
	}

	if envs == nil {
		envs = []core.Environment{}
	}

	return writeSuccess(w, envs)
}

// handleGetEnvironment returns a single environment by ID
func (s *Server) handleGetEnvironment(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	env, err := s.store.GetEnvironment(id)
	if err != nil {
		return err
	}

	return writeSuccess(w, env)
}

// handleUpdateEnvironment updates an existing environment
func (s *Server) handleUpdateEnvironment(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	var env core.Environment
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		return errors.NewInvalidInputErrorWithCause("invalid request body", err)
	}

	env.ID = id

	if env.Name == "" {
		return errors.NewInvalidInputErrorWithField("name", "name is required")
	}

	if err := s.store.UpdateEnvironment(&env); err != nil {
		return err
	}

	return writeSuccess(w, env)
}

// handleDeleteEnvironment removes an environment
func (s *Server) handleDeleteEnvironment(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return errors.NewInvalidInputErrorWithField("id", "id is required")
	}

	if err := s.store.DeleteEnvironment(id); err != nil {
		return err
	}

	writeNoContent(w)
	return nil
}
