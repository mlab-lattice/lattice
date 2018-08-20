package rest

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type CreateSystemRequest struct {
	// TBD
	ID v1.SystemID `json:"id"`
	// TBD
	DefinitionURL string `json:"definitionUrl" example:"git://github.com/foo/foo.git"`
}

type BuildRequest struct {
	// TBD
	Version v1.SystemVersion `json:"version"`
}

type DeployRequest struct {
	// TBD
	Version *v1.SystemVersion `json:"version,omitempty"`
	// TBD
	BuildID *v1.BuildID `json:"buildId,omitempty"`
}

type RunJobRequest struct {
	// TBD
	Path tree.NodePath `json:"path"`
	// TBD
	Command []string `json:"command"`
	// TBD
	Environment definitionv1.ContainerEnvironment `json:"environment"`
}

type SetSecretRequest struct {
	// TBD
	Value string `json:"value"`
}
