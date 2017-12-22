package constants

const (
	LabelKeyNodeRoleLattice  = "node-role.lattice.mlab.com"
	LabelKeyMasterNode       = LabelKeyNodeRoleLattice + "/master"
	LabelKeyBuildNode        = LabelKeyNodeRoleLattice + "/build"
	LabelKeyServiceNode      = LabelKeyNodeRoleLattice + "/service"
	LabelKeyNodeRoleNodePool = LabelKeyNodeRoleLattice + "/node-pool"

	LabelKeyComponentBuildID = "component.build.lattice.mlab.com/id"

	LabelKeyInternalComponent = "component.lattice.mlab.com/internal"

	LabelKeyNodePoolID        = "node-pool.lattice.mlab.com/id"
	LabelKeyNodePoolNamespace = "node-pool.lattice.mlab.com/namespace"

	LabelKeySystemRolloutVersion = "rollout.system.lattice.mlab.com/version"
	LabelKeySystemRolloutBuildID = "rollout.system.lattice.mlab.com/build"

	LabelKeyServiceID         = "service.lattice.mlab.com/id"
	LabelKeyServicePathDomain = "service.lattice.mlab.com/path-domain"

	LabelKeySystemBuildID      = "system.build.lattice.mlab.com/id"
	LabelKeySystemBuildVersion = "system.build.lattice.mlab.com/version"
	LabelKeySystemVersion      = "system.lattice.mlab.com/version"
)
