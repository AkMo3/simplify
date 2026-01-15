package store

import "github.com/AkMo3/simplify/internal/core"

// =============================================================================
// Pod Methods
// =============================================================================

// CreatePod stores a new pod. Overwrites if ID exists.
func (s *Store) CreatePod(pod *core.Pod) error {
	return s.genericCreate(BucketPods, pod.ID, pod)
}

// GetPod retrieves a pod by ID.
// Returns NotFoundError if the pod doesn't exist.
func (s *Store) GetPod(id string) (*core.Pod, error) {
	return genericGet[core.Pod](s, BucketPods, id)
}

// ListPods returns all pods.
func (s *Store) ListPods() ([]core.Pod, error) {
	return genericList[core.Pod](s, BucketPods)
}

// UpdatePod updates an existing pod.
// Returns NotFoundError if the pod doesn't exist.
func (s *Store) UpdatePod(pod *core.Pod) error {
	return s.genericUpdate(BucketPods, pod.ID, pod)
}

// DeletePod removes a pod by ID.
func (s *Store) DeletePod(id string) error {
	return s.genericDelete(BucketPods, id)
}

// PodExists checks if a pod exists.
func (s *Store) PodExists(id string) (bool, error) {
	return s.genericExists(BucketPods, id)
}

// =============================================================================
// Application Methods
// =============================================================================

// CreateApplication stores a new application. Overwrites if ID exists.
func (s *Store) CreateApplication(app *core.Application) error {
	return s.genericCreate(BucketApplications, app.ID, app)
}

// GetApplication retrieves an application by ID.
// Returns NotFoundError if the application doesn't exist.
func (s *Store) GetApplication(id string) (*core.Application, error) {
	return genericGet[core.Application](s, BucketApplications, id)
}

// ListApplications returns all applications.
func (s *Store) ListApplications() ([]core.Application, error) {
	return genericList[core.Application](s, BucketApplications)
}

// UpdateApplication updates an existing application.
// Returns NotFoundError if the application doesn't exist.
func (s *Store) UpdateApplication(app *core.Application) error {
	return s.genericUpdate(BucketApplications, app.ID, app)
}

// DeleteApplication removes an application by ID.
func (s *Store) DeleteApplication(id string) error {
	return s.genericDelete(BucketApplications, id)
}

// ApplicationExists checks if an application exists.
func (s *Store) ApplicationExists(id string) (bool, error) {
	return s.genericExists(BucketApplications, id)
}

// =============================================================================
// Team Methods
// =============================================================================

// CreateTeam stores a new team. Overwrites if ID exists.
func (s *Store) CreateTeam(team *core.Team) error {
	return s.genericCreate(BucketTeams, team.ID, team)
}

// GetTeam retrieves a team by ID.
// Returns NotFoundError if the team doesn't exist.
func (s *Store) GetTeam(id string) (*core.Team, error) {
	return genericGet[core.Team](s, BucketTeams, id)
}

// ListTeams returns all teams.
func (s *Store) ListTeams() ([]core.Team, error) {
	return genericList[core.Team](s, BucketTeams)
}

// UpdateTeam updates an existing team.
// Returns NotFoundError if the team doesn't exist.
func (s *Store) UpdateTeam(team *core.Team) error {
	return s.genericUpdate(BucketTeams, team.ID, team)
}

// DeleteTeam removes a team by ID.
func (s *Store) DeleteTeam(id string) error {
	return s.genericDelete(BucketTeams, id)
}

// TeamExists checks if a team exists.
func (s *Store) TeamExists(id string) (bool, error) {
	return s.genericExists(BucketTeams, id)
}

// =============================================================================
// Project Methods
// =============================================================================

// CreateProject stores a new project. Overwrites if ID exists.
func (s *Store) CreateProject(p *core.Project) error {
	return s.genericCreate(BucketProjects, p.ID, p)
}

// GetProject retrieves a project by ID.
// Returns NotFoundError if the project doesn't exist.
func (s *Store) GetProject(id string) (*core.Project, error) {
	return genericGet[core.Project](s, BucketProjects, id)
}

// ListProjects returns all projects.
func (s *Store) ListProjects() ([]core.Project, error) {
	return genericList[core.Project](s, BucketProjects)
}

// UpdateProject updates an existing project.
// Returns NotFoundError if the project doesn't exist.
func (s *Store) UpdateProject(p *core.Project) error {
	return s.genericUpdate(BucketProjects, p.ID, p)
}

// DeleteProject removes a project by ID.
func (s *Store) DeleteProject(id string) error {
	return s.genericDelete(BucketProjects, id)
}

// ProjectExists checks if a project exists.
func (s *Store) ProjectExists(id string) (bool, error) {
	return s.genericExists(BucketProjects, id)
}

// =============================================================================
// Environment Methods
// =============================================================================

// CreateEnvironment stores a new environment. Overwrites if ID exists.
func (s *Store) CreateEnvironment(env *core.Environment) error {
	return s.genericCreate(BucketEnvironments, env.ID, env)
}

// GetEnvironment retrieves an environment by ID.
// Returns NotFoundError if the environment doesn't exist.
func (s *Store) GetEnvironment(id string) (*core.Environment, error) {
	return genericGet[core.Environment](s, BucketEnvironments, id)
}

// ListEnvironments returns all environments.
func (s *Store) ListEnvironments() ([]core.Environment, error) {
	return genericList[core.Environment](s, BucketEnvironments)
}

// UpdateEnvironment updates an existing environment.
// Returns NotFoundError if the environment doesn't exist.
func (s *Store) UpdateEnvironment(env *core.Environment) error {
	return s.genericUpdate(BucketEnvironments, env.ID, env)
}

// DeleteEnvironment removes an environment by ID.
func (s *Store) DeleteEnvironment(id string) error {
	return s.genericDelete(BucketEnvironments, id)
}

// EnvironmentExists checks if an environment exists.
func (s *Store) EnvironmentExists(id string) (bool, error) {
	return s.genericExists(BucketEnvironments, id)
}
