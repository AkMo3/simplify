package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AkMo3/simplify/internal/core"
	"github.com/AkMo3/simplify/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestStore creates a temporary database for testing and returns the store
// along with a cleanup function to remove the file afterwards.
func setupTestStore(t *testing.T) (s *Store, cleanup func()) {
	t.Helper()

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

func TestNew(t *testing.T) {
	t.Run("creates database in new directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "subdir", "nested", "test.db")

		s, err := New(dbPath)
		require.NoError(t, err)
		defer s.Close()

		// Verify database file exists
		_, err = os.Stat(dbPath)
		assert.NoError(t, err)
	})

	t.Run("opens existing database", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		// Create and close first store
		s1, err := New(dbPath)
		require.NoError(t, err)
		s1.Close()

		// Open same database
		s2, err := New(dbPath)
		require.NoError(t, err)
		defer s2.Close()

		assert.NotNil(t, s2)
	})
}

func TestStore_Ping(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	err := s.Ping()
	assert.NoError(t, err)
}

func TestApplicationCRUD(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Define a sample application
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

	t.Run("Create", func(t *testing.T) {
		err := s.CreateApplication(app)
		require.NoError(t, err, "failed to create application")
	})

	t.Run("Get", func(t *testing.T) {
		fetchedApp, err := s.GetApplication(app.ID)
		require.NoError(t, err, "failed to get application")
		assert.Equal(t, app.ID, fetchedApp.ID)
		assert.Equal(t, app.Name, fetchedApp.Name)
		assert.Equal(t, app.Replicas, fetchedApp.Replicas)
		assert.WithinDuration(t, app.CreatedAt, fetchedApp.CreatedAt, time.Second)
	})

	t.Run("List", func(t *testing.T) {
		apps, err := s.ListApplications()
		require.NoError(t, err, "failed to list applications")
		assert.Len(t, apps, 1)
		assert.Equal(t, app.ID, apps[0].ID)
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := s.ApplicationExists(app.ID)
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = s.ApplicationExists("non-existent")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Update", func(t *testing.T) {
		app.Replicas = 5
		app.UpdatedAt = time.Now().UTC()

		err := s.UpdateApplication(app)
		require.NoError(t, err)

		updatedApp, err := s.GetApplication(app.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, updatedApp.Replicas)
	})

	t.Run("Delete", func(t *testing.T) {
		err := s.DeleteApplication(app.ID)
		require.NoError(t, err, "failed to delete application")

		// Verify deletion
		_, err = s.GetApplication(app.ID)
		assert.Error(t, err)
		assert.True(t, errors.IsNotFound(err), "expected NotFoundError")
	})
}

func TestApplicationErrors(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	t.Run("Get non-existent returns NotFoundError", func(t *testing.T) {
		_, err := s.GetApplication("non-existent-id")

		require.Error(t, err)
		assert.True(t, errors.IsNotFound(err), "expected NotFoundError, got: %T", err)

		// Verify error details
		var notFoundErr *errors.NotFoundError
		require.ErrorAs(t, err, &notFoundErr)
		assert.Equal(t, "applications", notFoundErr.Resource)
		assert.Equal(t, "non-existent-id", notFoundErr.ID)
	})

	t.Run("Update non-existent returns NotFoundError", func(t *testing.T) {
		app := &core.Application{
			ID:   "non-existent",
			Name: "test",
		}

		err := s.UpdateApplication(app)

		require.Error(t, err)
		assert.True(t, errors.IsNotFound(err), "expected NotFoundError, got: %T", err)
	})
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

	t.Run("Create", func(t *testing.T) {
		err := s.CreateTeam(team)
		require.NoError(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		fetched, err := s.GetTeam(team.ID)
		require.NoError(t, err)
		assert.Equal(t, team.Name, fetched.Name)
	})

	t.Run("List", func(t *testing.T) {
		teams, err := s.ListTeams()
		require.NoError(t, err)
		assert.Len(t, teams, 1)
	})

	t.Run("Update", func(t *testing.T) {
		team.Name = "Platform Team"
		err := s.UpdateTeam(team)
		require.NoError(t, err)

		updated, err := s.GetTeam(team.ID)
		require.NoError(t, err)
		assert.Equal(t, "Platform Team", updated.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		err := s.DeleteTeam(team.ID)
		require.NoError(t, err)

		_, err = s.GetTeam(team.ID)
		assert.True(t, errors.IsNotFound(err))
	})
}

func TestProjectCRUD(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	proj := &core.Project{
		ID:        "proj-1",
		TeamID:    "team-1",
		Name:      "API Gateway",
		CreatedAt: time.Now(),
	}

	t.Run("Create", func(t *testing.T) {
		err := s.CreateProject(proj)
		require.NoError(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		fetched, err := s.GetProject(proj.ID)
		require.NoError(t, err)
		assert.Equal(t, "API Gateway", fetched.Name)
	})

	t.Run("Update", func(t *testing.T) {
		proj.Name = "API Gateway v2"
		err := s.UpdateProject(proj)
		require.NoError(t, err)

		updated, err := s.GetProject(proj.ID)
		require.NoError(t, err)
		assert.Equal(t, "API Gateway v2", updated.Name)
	})
}

func TestEnvironmentCRUD(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	env := &core.Environment{
		ID:        "env-1",
		ProjectID: "proj-1",
		Name:      "Staging",
		CreatedAt: time.Now(),
	}

	t.Run("Create", func(t *testing.T) {
		err := s.CreateEnvironment(env)
		require.NoError(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		fetched, err := s.GetEnvironment(env.ID)
		require.NoError(t, err)
		assert.Equal(t, "Staging", fetched.Name)
	})

	t.Run("Update", func(t *testing.T) {
		env.Name = "Production"
		err := s.UpdateEnvironment(env)
		require.NoError(t, err)

		updated, err := s.GetEnvironment(env.ID)
		require.NoError(t, err)
		assert.Equal(t, "Production", updated.Name)
	})
}

func TestGenericCreateIfNotExists(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	app := &core.Application{
		ID:   "app-unique",
		Name: "Unique App",
	}

	t.Run("creates when not exists", func(t *testing.T) {
		err := s.genericCreateIfNotExists(BucketApplications, app.ID, app)
		require.NoError(t, err)

		fetched, err := s.GetApplication(app.ID)
		require.NoError(t, err)
		assert.Equal(t, app.Name, fetched.Name)
	})

	t.Run("returns AlreadyExistsError when exists", func(t *testing.T) {
		err := s.genericCreateIfNotExists(BucketApplications, app.ID, app)

		require.Error(t, err)
		assert.True(t, errors.IsAlreadyExists(err), "expected AlreadyExistsError, got: %T", err)

		var alreadyExistsErr *errors.AlreadyExistsError
		require.ErrorAs(t, err, &alreadyExistsErr)
		assert.Equal(t, "applications", alreadyExistsErr.Resource)
		assert.Equal(t, "app-unique", alreadyExistsErr.ID)
	})
}

func TestEmptyList(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// List should return empty slice, not nil
	apps, err := s.ListApplications()
	require.NoError(t, err)
	assert.NotNil(t, apps)
	assert.Len(t, apps, 0)
}

func TestDeleteIdempotent(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Deleting non-existent item should not error (BoltDB behavior)
	err := s.DeleteApplication("non-existent")
	assert.NoError(t, err)
}

func TestMultipleItems(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Create multiple applications
	for i := 0; i < 10; i++ {
		app := &core.Application{
			ID:   "app-" + string(rune('a'+i)),
			Name: "App " + string(rune('A'+i)),
		}
		err := s.CreateApplication(app)
		require.NoError(t, err)
	}

	// List all
	apps, err := s.ListApplications()
	require.NoError(t, err)
	assert.Len(t, apps, 10)
}
