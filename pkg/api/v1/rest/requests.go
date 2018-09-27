package rest

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

// swagger:model CreateSystemRequest
type CreateSystemRequest struct {
	ID            v1.SystemID `json:"id"`
	DefinitionURL string      `json:"definitionUrl"`
}

// swagger:model BuildRequest
type BuildRequest struct {
	Path    *tree.Path  `json:"path,omitempty"`
	Version *v1.Version `json:"version,omitempty"`
}

// swagger:model DeployRequest
type DeployRequest struct {
	BuildID *v1.BuildID `json:"buildId,omitempty"`
	Path    *tree.Path  `json:"path,omitempty"`
	Version *v1.Version `json:"version,omitempty"`
}

// swagger:model JobRequest
type RunJobRequest struct {
	Path        tree.Path                         `json:"path"`
	Command     []string                          `json:"command,omitempty"`
	Environment definitionv1.ContainerEnvironment `json:"environment,omitempty"`
}

// swagger:model SetSecretRequest
type SetSecretRequest struct {
	Value string `json:"value"`
}
