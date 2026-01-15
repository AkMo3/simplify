package store

import (
	"encoding/json"

	"github.com/AkMo3/simplify/internal/core"
	"github.com/AkMo3/simplify/internal/errors"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

// ListNetworks retrieves all networks from the database
func (s *Store) ListNetworks() ([]core.Network, error) {
	var networks []core.Network

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketNetworks))
		if b == nil {
			return nil // Bucket might not exist yet if freshly migrated
		}

		return b.ForEach(func(k, v []byte) error {
			var network core.Network
			if err := json.Unmarshal(v, &network); err != nil {
				return errors.NewInternalErrorWithCause("failed to unmarshal network", err)
			}
			networks = append(networks, network)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return networks, nil
}

// GetNetwork retrieves a single network by ID
func (s *Store) GetNetwork(id string) (*core.Network, error) {
	var network core.Network

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketNetworks))
		v := b.Get([]byte(id))
		if v == nil {
			return errors.NewNotFoundError("network", id)
		}

		if err := json.Unmarshal(v, &network); err != nil {
			return errors.NewInternalErrorWithCause("failed to unmarshal network", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &network, nil
}

// CreateNetwork stores a new network in the database
func (s *Store) CreateNetwork(network *core.Network) error {
	if network.ID == "" {
		network.ID = uuid.New().String()
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketNetworks))

		// Check for duplicate name
		err := b.ForEach(func(k, v []byte) error {
			var n core.Network
			if err := json.Unmarshal(v, &n); err != nil {
				return nil
			}
			if n.Name == network.Name {
				return errors.NewAlreadyExistsError("network", network.Name)
			}
			return nil
		})
		if err != nil {
			return err
		}

		data, err := json.Marshal(network)
		if err != nil {
			return errors.NewInternalErrorWithCause("failed to marshal network", err)
		}

		if err := b.Put([]byte(network.ID), data); err != nil {
			return errors.NewInternalErrorWithCause("failed to save network", err)
		}

		return nil
	})
}

// DeleteNetwork removes a network from the database
func (s *Store) DeleteNetwork(id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketNetworks))
		if b.Get([]byte(id)) == nil {
			return errors.NewNotFoundError("network", id)
		}

		if err := b.Delete([]byte(id)); err != nil {
			return errors.NewInternalErrorWithCause("failed to delete network", err)
		}
		return nil
	})
}
