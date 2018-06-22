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
	Version v1.SystemVersion `json:"version"`
}

type DeployRequest struct {
	Version *v1.SystemVersion `json:"version,omitempty"`
	BuildID *v1.BuildID       `json:"buildId,omitempty"`
}

type RunJobRequest struct {
	Path        tree.NodePath                     `json:"path"`
	Command     []string                          `json:"command"`
	Environment definitionv1.ContainerEnvironment `json:"environment"`
}

type SetSecretRequest struct {
	Value string `json:"value"`
}
