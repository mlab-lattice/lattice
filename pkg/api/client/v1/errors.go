package v1

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type InvalidSystemOptionsError struct {
	Reason string
}

func (e *InvalidSystemOptionsError) Error() string {
	return fmt.Sprintf("invalid system: %v", e.Reason)
}

type SystemAlreadyExistsError struct {
	ID v1.SystemID
}

func (e *SystemAlreadyExistsError) Error() string {
	return fmt.Sprintf("system %v already exists", e.ID)
}

type InvalidSystemIDError struct {
	ID v1.SystemID
}

func (e *InvalidSystemIDError) Error() string {
	return fmt.Sprintf("invalid system %v", e.ID)
}

type InvalidSystemVersionError struct {
	Version string
}

func (e *InvalidSystemVersionError) Error() string {
	return fmt.Sprintf("invalid system version %v", e.Version)
}

type InvalidBuildIDError struct {
	ID v1.BuildID
}

func (e *InvalidBuildIDError) Error() string {
	return fmt.Sprintf("invalid build %v", e.ID)
}

type InvalidDeployIDError struct {
	ID v1.DeployID
}

func (e *InvalidDeployIDError) Error() string {
	return fmt.Sprintf("invalid rollout %v", e.ID)
}

type InvalidTeardownIDError struct {
	ID v1.TeardownID
}

func (e *InvalidTeardownIDError) Error() string {
	return fmt.Sprintf("invalid teardown %v", e.ID)
}

type InvalidServiceIDError struct {
	ID v1.ServiceID
}

func (e *InvalidServiceIDError) Error() string {
	return fmt.Sprintf("invalid service %v", e.ID)
}

type InvalidSecretError struct {
	Path tree.NodePath
	Name string
}

func (e *InvalidSecretError) Error() string {
	return fmt.Sprintf("invalid secret %v:%v", e.Path, e.Name)
}
