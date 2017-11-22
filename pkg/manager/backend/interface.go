package backend

import (
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
	coretypes "github.com/mlab-lattice/core/pkg/types"
)

type Interface interface {
	// Utils
	GetSystemUrl(ln coretypes.LatticeNamespace) (url string, err error)

	// Builds
	BuildSystem(ln coretypes.LatticeNamespace, definitionRoot systemtree.Node, v coretypes.SystemVersion) (bid coretypes.SystemBuildId, err error)
	ListSystemBuilds(ln coretypes.LatticeNamespace) (b []coretypes.SystemBuild, err error)
	GetSystemBuild(ln coretypes.LatticeNamespace, buildId coretypes.SystemBuildId) (b *coretypes.SystemBuild, exists bool, err error)

	// Rollouts
	RollOutSystemBuild(ln coretypes.LatticeNamespace, bid coretypes.SystemBuildId) (rid coretypes.SystemRolloutId, err error)
	RollOutSystem(ln coretypes.LatticeNamespace, definitionRoot systemtree.Node, v coretypes.SystemVersion) (rid coretypes.SystemRolloutId, err error)
	ListSystemRollouts(ln coretypes.LatticeNamespace) (r []coretypes.SystemRollout, err error)
	GetSystemRollout(ln coretypes.LatticeNamespace, rid coretypes.SystemRolloutId) (r *coretypes.SystemRollout, exists bool, err error)

	// Teardowns
	TearDownSystem(ln coretypes.LatticeNamespace) (tid coretypes.SystemTeardownId, err error)
	ListSystemTeardowns(ln coretypes.LatticeNamespace) (t []coretypes.SystemTeardown, err error)
	GetSystemTeardown(ln coretypes.LatticeNamespace, tid coretypes.SystemTeardownId) (t *coretypes.SystemTeardown, exists bool, err error)

	// Services
	ListSystemServices(ln coretypes.LatticeNamespace) (svcs []coretypes.Service, err error)
	GetSystemService(ln coretypes.LatticeNamespace, p systemtree.NodePath) (svc *coretypes.Service, err error)
}
