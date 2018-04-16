package constants

const (
	LabelKeyLatticeID = "lattice.mlab.com/id"

	LabelKeyControlPlane        = "control-plane.lattice.mlab.com"
	LabelKeyControlPlaneService = LabelKeyControlPlane + "/service"

	LabelKeyNodeRollKubernetes       = "node-role.kubernetes.io"
	LabelKeyNodeRollKubernetesMaster = LabelKeyNodeRollKubernetes + "/master"

	LabelKeyNodeRoleLattice         = "node-role.lattice.mlab.com"
	LabelKeyNodeRoleLatticeMaster   = LabelKeyNodeRoleLattice + "/master"
	LabelKeyNodeRoleLatticeBuild    = LabelKeyNodeRoleLattice + "/build"
	LabelKeyNodeRoleLatticeNodePool = LabelKeyNodeRoleLattice + "/node-pool"

	LabelKeyNodePool           = "node-pool.lattice.mlab.com"
	LabelKeyNodePoolPath       = LabelKeyNodePool + "/path"
	LabelKeyNodePoolGeneration = LabelKeyNodePool + "/generation"

	LabelKeyComponentBuildID = "component.build.lattice.mlab.com/id"

	LabelKeySystemRolloutVersion = "rollout.system.lattice.mlab.com/version"
	LabelKeySystemRolloutBuildID = "rollout.system.lattice.mlab.com/build"

	LabelKeyServiceID   = "service.lattice.mlab.com/id"
	LabelKeyServicePath = "service.lattice.mlab.com/path"

	LabelKeySystemBuildID = "system.build.lattice.mlab.com/id"
	LabelKeySystemVersion = "system.lattice.mlab.com/version"

	LabelKeySecret = "secret.lattice.mlab.com"
)
