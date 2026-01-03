package store

import (
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
)

// genericCreate stores an item in the specified bucket using the provided ID key.
func (s *Store) genericCreate(bucketName, id string, item any) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %w", err)
		}

		if err := b.Put([]byte(id), data); err != nil {
			return fmt.Errorf("failed to put item: %w", err)
		}

		return nil
	})
}

// genericGet retrieves a single item by ID from the specified bucket.
func genericGet[T any](s *Store, bucketName, id string) (*T, error) {
	var item T
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("item not found: %s", id)
		}

		if err := json.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("failed to unmarshal item: %w", err)
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
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		return b.ForEach(func(k, v []byte) error {
			var item T
			if err := json.Unmarshal(v, &item); err != nil {
				return fmt.Errorf("failed to unmarshal item (id=%s): %w", string(k), err)
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
func (s *Store) genericDelete(bucketName, id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		return b.Delete([]byte(id))
	})
}
