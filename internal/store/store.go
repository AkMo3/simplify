// Package store provides database operations using BoltDB
package store

import (
	"fmt"
	"time"

	"github.com/AkMo3/simplify/internal/errors"
	"github.com/AkMo3/simplify/internal/permissions"
	"go.etcd.io/bbolt"
)

// Bucket names constants to avoid typos
const (
	BucketTeams        = "teams"
	BucketProjects     = "projects"
	BucketEnvironments = "environments"
	BucketApplications = "applications"
	BucketPods         = "pods"
	BucketNetworks     = "networks"
)

// Store holds the database connection
type Store struct {
	db *bbolt.DB
}

// New creates a new Store and initializes the database buckets.
// It ensures the database directory exists and is writable before opening.
func New(dbPath string) (*Store, error) {
	// Ensure the database directory exists and is writable
	if err := permissions.EnsureFileWritable(dbPath); err != nil {
		return nil, err
	}

	// Open the database with a 1-second timeout to prevent hanging if locked
	db, err := bbolt.Open(dbPath, 0o600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, errors.NewInternalErrorWithCause(
			fmt.Sprintf("failed to open database at %s", dbPath), err)
	}

	s := &Store{db: db}

	// Initialize buckets immediately
	if err := s.initBuckets(); err != nil {
		s.Close()
		return nil, err
	}

	return s, nil
}

// initBuckets creates the necessary buckets if they don't exist
func (s *Store) initBuckets() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		buckets := []string{
			BucketTeams,
			BucketProjects,
			BucketEnvironments,
			BucketApplications,
			BucketPods,
			BucketNetworks,
		}

		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return errors.NewInternalErrorWithCause(
					fmt.Sprintf("failed to create bucket %s", bucket), err)
			}
		}

		return nil
	})
}

// Close ensures the database file is released
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB returns the underlying BoltDB instance for advanced operations.
// Use with caution - prefer the typed methods when possible.
func (s *Store) DB() *bbolt.DB {
	return s.db
}

// Ping verifies the database connection is healthy.
// Used for health checks.
func (s *Store) Ping() error {
	return s.db.View(func(tx *bbolt.Tx) error {
		// Just verify we can start a transaction
		return nil
	})
}
