package store

import "github.com/AkMo3/simplify/internal/core"

func (s *Store) CreateProject(p *core.Project) error {
	return s.genericCreate(BucketProjects, p.ID, p)
}

func (s *Store) GetProject(id string) (*core.Project, error) {
	return genericGet[core.Project](s, BucketProjects, id)
}

func (s *Store) ListProjects() ([]core.Project, error) {
	return genericList[core.Project](s, BucketProjects)
}

func (s *Store) DeleteProject(id string) error {
	return s.genericDelete(BucketProjects, id)
}
