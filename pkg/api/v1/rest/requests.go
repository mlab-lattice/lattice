package rest

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type CreateSystemRequest struct {
	ID            v1.SystemID `json:"id"`
	DefinitionURL string      `json:"definitionUrl"`
}

type BuildRequest struct {
	Path    *tree.Path        `json:"path,omitempty"`
	Version *v1.SystemVersion `json:"version,omitempty"`
}

type DeployRequest struct {
	BuildID *v1.BuildID       `json:"buildId,omitempty"`
	Path    *tree.Path        `json:"path,omitempty"`
	Version *v1.SystemVersion `json:"version,omitempty"`
}

type RunJobRequest struct {
	Path        tree.Path                         `json:"path"`
	Command     []string                          `json:"command,omitempty"`
	Environment definitionv1.ContainerEnvironment `json:"environment,omitempty"`
}

type SetSecretRequest struct {
	Value string `json:"value"`
}
