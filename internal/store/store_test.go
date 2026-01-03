package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AkMo3/simplify/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestStore creates a temporary database for testing and returns the store
// along with a cleanup function to remove the file afterwards.
func setupTestStore(t *testing.T) (s *Store, cleanup func()) {
	// Create a temporary directory for the test db
	tmpDir, err := os.MkdirTemp("", "simplify-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test.db")

	s, err = New(dbPath)
	require.NoError(t, err)

	cleanup = func() {
		s.Close()
		os.RemoveAll(tmpDir)
	}

	return s, cleanup
}

func TestApplicationCRUD(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// 1. Define a sample application
	app := &core.Application{
		ID:            "app-123",
		EnvironmentID: "env-prod",
		Name:          "web-server",
		Image:         "nginx:latest",
		Ports:         map[string]string{"80": "80"},
		Replicas:      3,
		EnvVars:       map[string]string{"ENV": "production"},
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// 2. Test Create
	err := s.CreateApplication(app)
	require.NoError(t, err, "failed to create application")

	// 3. Test Get
	fetchedApp, err := s.GetApplication(app.ID)
	require.NoError(t, err, "failed to get application")
	assert.Equal(t, app.ID, fetchedApp.ID)
	assert.Equal(t, app.Name, fetchedApp.Name)
	assert.Equal(t, app.Replicas, fetchedApp.Replicas)
	// Compare times carefully due to serialization precision
	assert.WithinDuration(t, app.CreatedAt, fetchedApp.CreatedAt, time.Second)

	// 4. Test List
	apps, err := s.ListApplications()
	require.NoError(t, err, "failed to list applications")
	assert.Len(t, apps, 1)
	assert.Equal(t, app.ID, apps[0].ID)

	// 5. Test Update (Create with same ID)
	app.Replicas = 5
	err = s.CreateApplication(app)
	require.NoError(t, err)

	updatedApp, err := s.GetApplication(app.ID)
	require.NoError(t, err)
	assert.Equal(t, 5, updatedApp.Replicas)

	// 6. Test Delete
	err = s.DeleteApplication(app.ID)
	require.NoError(t, err, "failed to delete application")

	// 7. Verify Deletion
	_, err = s.GetApplication(app.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	apps, err = s.ListApplications()
	require.NoError(t, err)
	assert.Len(t, apps, 0)
}

func TestTeamCRUD(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	team := &core.Team{
		ID:        "team-1",
		Name:      "Platform Engineering",
		Slug:      "platform",
		CreatedAt: time.Now(),
	}

	// Create
	err := s.CreateTeam(team)
	require.NoError(t, err)

	// Get
	fetched, err := s.GetTeam(team.ID)
	require.NoError(t, err)
	assert.Equal(t, team.Name, fetched.Name)

	// List
	teams, err := s.ListTeams()
	require.NoError(t, err)
	assert.Len(t, teams, 1)
}

func TestProjectAndEnvironmentWiring(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Test Project
	proj := &core.Project{
		ID:     "proj-1",
		TeamID: "team-1",
		Name:   "API Gateway",
	}
	require.NoError(t, s.CreateProject(proj))
	fetchedProj, err := s.GetProject("proj-1")
	require.NoError(t, err)
	assert.Equal(t, "API Gateway", fetchedProj.Name)

	// Test Environment
	env := &core.Environment{
		ID:        "env-1",
		ProjectID: "proj-1",
		Name:      "Staging",
	}
	require.NoError(t, s.CreateEnvironment(env))
	fetchedEnv, err := s.GetEnvironment("env-1")
	require.NoError(t, err)
	assert.Equal(t, "Staging", fetchedEnv.Name)
}

func TestMissingItem(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Test getting a non-existent item
	_, err := s.GetApplication("non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
