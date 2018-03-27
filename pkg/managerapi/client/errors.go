package client

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"
)

type InvalidSystemOptionsError struct {
	Reason string
}

func (e *InvalidSystemOptionsError) Error() string {
	return fmt.Sprintf("invalid system: %v", e.Reason)
}

type SystemAlreadyExistsError struct {
	ID types.SystemID
}

func (e *SystemAlreadyExistsError) Error() string {
	return fmt.Sprintf("system %v already exists", e.ID)
}

type InvalidSystemIDError struct {
	ID types.SystemID
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
	ID types.SystemBuildID
}

func (e *InvalidBuildIDError) Error() string {
	return fmt.Sprintf("invalid build %v", e.ID)
}

type InvalidRolloutIDError struct {
	ID types.SystemRolloutID
}

func (e *InvalidRolloutIDError) Error() string {
	return fmt.Sprintf("invalid rollout %v", e.ID)
}

type InvalidTeardownIDError struct {
	ID types.SystemTeardownID
}

func (e *InvalidTeardownIDError) Error() string {
	return fmt.Sprintf("invalid teardown %v", e.ID)
}

type InvalidServiceIDError struct {
	ID types.ServiceID
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
