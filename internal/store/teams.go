package store

import "github.com/AkMo3/simplify/internal/core"

func (s *Store) CreateTeam(team *core.Team) error {
	return s.genericCreate(BucketTeams, team.ID, team)
}

func (s *Store) GetTeam(id string) (*core.Team, error) {
	return genericGet[core.Team](s, BucketTeams, id)
}

func (s *Store) ListTeams() ([]core.Team, error) {
	return genericList[core.Team](s, BucketTeams)
}

func (s *Store) DeleteTeam(id string) error {
	return s.genericDelete(BucketTeams, id)
}
