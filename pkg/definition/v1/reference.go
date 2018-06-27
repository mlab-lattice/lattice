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
	GitRepository *GitRepository `json:"git_repository,omitempty"`
	File          *string        `json:"file,omitempty"`
}

type ReferenceOrResource struct {
	Reference *Reference
	Resource  component.Interface
}
