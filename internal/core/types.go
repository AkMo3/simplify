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
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Ports             map[string]string `json:"ports"`
	EnvVars           map[string]string `json:"env_vars"`
	ID                string            `json:"id"`
	EnvironmentID     string            `json:"environment_id"`
	Name              string            `json:"name"`
	Image             string            `json:"image"`
	Status            string            `json:"status"`
	HealthStatus      string            `json:"health_status"`
	Replicas          int               `json:"replicas"`
	PodID             string            `json:"pod_id,omitempty"`
	NetworkID         string            `json:"network_id,omitempty"`
	IPAddress         string            `json:"ip_address,omitempty"`
	ExposedPorts      []string          `json:"exposed_ports,omitempty"`
	ConnectedNetworks []string          `json:"connected_networks,omitempty"`
}

// Pod represents a shared network namespace for multiple applications
type Pod struct {
	CreatedAt time.Time         `json:"created_at"`
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Ports     map[string]string `json:"ports"` // Host:Container
	Status    string            `json:"status"`
}

// Network represents a bridge network for container communication
type Network struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Subnet    string    `json:"subnet"`
	Driver    string    `json:"driver"`
}
