package store

import "github.com/AkMo3/simplify/internal/core"

func (s *Store) CreateApplication(app *core.Application) error {
	return s.genericCreate(BucketApplications, app.ID, app)
}

func (s *Store) GetApplication(id string) (*core.Application, error) {
	return genericGet[core.Application](s, BucketApplications, id)
}

func (s *Store) ListApplications() ([]core.Application, error) {
	return genericList[core.Application](s, BucketApplications)
}

func (s *Store) DeleteApplication(id string) error {
	return s.genericDelete(BucketApplications, id)
}
