package rest

import (
	"github.com/mlab-lattice/system/pkg/api/v1"
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

type SetSecretRequest struct {
	Value string `json:"value"`
}
