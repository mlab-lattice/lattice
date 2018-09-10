package rest

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type CreateSystemRequest struct {
	// System ID
	ID v1.SystemID `json:"id"`
	// URL for for where the system definition resides in.
	DefinitionURL string `json:"definitionUrl" example:"git://github.com/foo/foo.git"`
}

type BuildRequest struct {
	// Version of system to build
	Version v1.SystemVersion `json:"version"`
}

type DeployRequest struct {
	// Version of system to deploy
	Version *v1.SystemVersion `json:"version,omitempty"`
	// BuildID to deploy
	BuildID *v1.BuildID `json:"buildId,omitempty"`
}

type RunJobRequest struct {
	// Path to run the job agsinst
	Path tree.Path `json:"path"`
	// Command to run
	Command []string `json:"command"`
	// Container environment
	Environment definitionv1.ContainerEnvironment `json:"environment"`
}

type SetSecretRequest struct {
	// Secret Value
	Value string `json:"value"`
}
