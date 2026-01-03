// Package store provides database operations
package store

import (
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

// Bucket names constants to avoid types
const (
	BucketTeams        = "teams"
	BucketProjects     = "projects"
	BucketEnvironments = "environments"
	BucketApplications = "applications"
)

// Store holds the database connection
type Store struct {
	db *bbolt.DB
}

// New creates a new Store and initializes the database buckets
func New(dbPath string) (*Store, error) {
	// Open the database with a 1-second timeout to prevent hanging if locked
	db, err := bbolt.Open(dbPath, 0o600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt db: %w", err)
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
		}

		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}

		return nil
	})
}

// Close ensures the database file is released
func (s *Store) Close() error {
	return s.db.Close()
}
