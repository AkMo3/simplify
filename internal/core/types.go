// Package core provides all the core domain types for Simplify
package core

import "time"

// Team represents a group of users (e.g. "Engineering", "Platform")
type Team struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
}

// Project represents a specific codebase or service group (e.g., "simplify-api")
type Project struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	Name      string    `json:"name"`
	RepoURL   string    `json:"repo_url"`
}

// Environment represents a deployment target (e.g., "prod", "staging")
type Environment struct {
	CreatedAt time.Time         `json:"created_at"`
	Config    map[string]string `json:"config"`
	ID        string            `json:"id"`
	ProjectID string            `json:"project_id"`
	Name      string            `json:"name"`
}

// Application represents a running service configuration
type Application struct {
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Ports         map[string]string `json:"ports"`
	EnvVars       map[string]string `json:"env_vars"`
	ID            string            `json:"id"`
	EnvironmentID string            `json:"environment_id"`
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	Replicas      int               `json:"replicas"`
}
