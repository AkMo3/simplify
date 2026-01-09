package store

import (
	"encoding/json"

	"github.com/AkMo3/simplify/internal/errors"
	"go.etcd.io/bbolt"
)

// genericCreate stores an item in the specified bucket using the provided ID key.
func (s *Store) genericCreate(bucketName, id string, item any) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return errors.NewInternalError("bucket " + bucketName + " not found")
		}

		data, err := json.Marshal(item)
		if err != nil {
			return errors.NewInternalErrorWithCause("failed to marshal item", err)
		}

		if err := b.Put([]byte(id), data); err != nil {
			return errors.NewInternalErrorWithCause("failed to store item", err)
		}

		return nil
	})
}

// genericGet retrieves a single item by ID from the specified bucket.
// Returns NotFoundError if the item doesn't exist.
func genericGet[T any](s *Store, bucketName, id string) (*T, error) {
	var item T
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return errors.NewInternalError("bucket " + bucketName + " not found")
		}

		data := b.Get([]byte(id))
		if data == nil {
			return errors.NewNotFoundError(bucketName, id)
		}

		if err := json.Unmarshal(data, &item); err != nil {
			return errors.NewInternalErrorWithCause("failed to unmarshal item", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// genericList retrieves all items from the specified bucket.
func genericList[T any](s *Store, bucketName string) ([]T, error) {
	var items []T

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return errors.NewInternalError("bucket " + bucketName + " not found")
		}

		return b.ForEach(func(k, v []byte) error {
			var item T
			if err := json.Unmarshal(v, &item); err != nil {
				return errors.NewInternalErrorWithCause(
					"failed to unmarshal item with id "+string(k), err)
			}
			items = append(items, item)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

// genericDelete removes an item by ID from the specified bucket.
// Note: BoltDB Delete is idempotent - it doesn't error if the key doesn't exist.
func (s *Store) genericDelete(bucketName, id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return errors.NewInternalError("bucket " + bucketName + " not found")
		}

		return b.Delete([]byte(id))
	})
}

// genericExists checks if an item exists in the specified bucket.
func (s *Store) genericExists(bucketName, id string) (bool, error) {
	var exists bool
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return errors.NewInternalError("bucket " + bucketName + " not found")
		}

		exists = b.Get([]byte(id)) != nil
		return nil
	})
	return exists, err
}

// genericUpdate updates an existing item in the specified bucket.
// Returns NotFoundError if the item doesn't exist.
func (s *Store) genericUpdate(bucketName, id string, item any) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return errors.NewInternalError("bucket " + bucketName + " not found")
		}

		// Check if item exists
		if b.Get([]byte(id)) == nil {
			return errors.NewNotFoundError(bucketName, id)
		}

		data, err := json.Marshal(item)
		if err != nil {
			return errors.NewInternalErrorWithCause("failed to marshal item", err)
		}

		if err := b.Put([]byte(id), data); err != nil {
			return errors.NewInternalErrorWithCause("failed to update item", err)
		}

		return nil
	})
}

// genericCreateIfNotExists creates an item only if it doesn't already exist.
// Returns AlreadyExistsError if the item exists.
func (s *Store) genericCreateIfNotExists(bucketName, id string, item any) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return errors.NewInternalError("bucket " + bucketName + " not found")
		}

		// Check if item already exists
		if b.Get([]byte(id)) != nil {
			return errors.NewAlreadyExistsError(bucketName, id)
		}

		data, err := json.Marshal(item)
		if err != nil {
			return errors.NewInternalErrorWithCause("failed to marshal item", err)
		}

		if err := b.Put([]byte(id), data); err != nil {
			return errors.NewInternalErrorWithCause("failed to store item", err)
		}

		return nil
	})
}
