package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/component"
)

const ComponentTypeReference = "reference"

var ReferenceType = component.Type{
	APIVersion: APIVersion,
	Type:       ComponentTypeReference,
}

type Reference struct {
	GitRepository *GitRepositoryReference `json:"git_repository,omitempty"`
	File          *string                 `json:"file,omitempty"`
}

type GitRepositoryReference struct {
	File string `json:"file"`
	*GitRepository
}

func (r *Reference) Type() component.Type {
	return ReferenceType
}
