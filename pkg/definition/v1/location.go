package v1

import (
	"encoding/json"
	"fmt"
)

const (
	LocationTypeGitRepository = "git_repository"
)

type Location struct {
	GitRepository *GitRepository `json:"git_repository,omitempty"`
}

// XXX <GEB>: should this be locationDecoder? following the convention in container.go for now
type locationEncoder struct {
	Type string `json:"type"`
}

type locationGitRepositoryEncoder struct {
	Type string `json:"type"`
	*GitRepository
}

func (l *Location) UnmarshalJSON(data []byte) error {
	var e *locationEncoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	switch e.Type {
	case LocationTypeGitRepository:
		var g *GitRepository
		if err := json.Unmarshal(data, &g); err != nil {
			return err
		}

		l.GitRepository = g
		return nil

	default:
		return fmt.Errorf("unrecognized location type: %v", e.Type)
	}
}

func (l *Location) MarshalJSON() ([]byte, error) {
	var e interface{}
	switch {
	case l.GitRepository != nil:
		e = &locationGitRepositoryEncoder{
			Type:          LocationTypeGitRepository,
			GitRepository: l.GitRepository,
		}

	default:
		return nil, fmt.Errorf("location must have a type")
	}

	return json.Marshal(&e)
}
