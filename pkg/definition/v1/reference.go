package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/resource"
)

const ResourceTypeReference = "reference"

var ReferenceType = resource.Type{
	APIVersion: APIVersion,
	Type:       ResourceTypeReference,
}

type Reference struct {
	GitRepository *GitRepository `json:"git_repository,omitempty"`
	File          *string        `json:"file,omitempty"`
}

type ReferenceOrResource struct {
	Reference *Reference
	Resource  resource.Interface
}
