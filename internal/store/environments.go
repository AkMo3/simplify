package store

import "github.com/AkMo3/simplify/internal/core"

func (s *Store) CreateEnvironment(env *core.Environment) error {
	return s.genericCreate(BucketEnvironments, env.ID, env)
}

func (s *Store) GetEnvironment(id string) (*core.Environment, error) {
	return genericGet[core.Environment](s, BucketEnvironments, id)
}

func (s *Store) ListEnvironments() ([]core.Environment, error) {
	return genericList[core.Environment](s, BucketEnvironments)
}

func (s *Store) DeleteEnvironment(id string) error {
	return s.genericDelete(BucketEnvironments, id)
}
