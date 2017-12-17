package constants

const (
	LabelKeyKubernetesNodeRole = "node-role.kubernetes.io"
	LabelKeyMasterNode         = LabelKeyKubernetesNodeRole + "/lattice-master"
	LabelKeyBuildNode          = LabelKeyKubernetesNodeRole + "/lattice-build"
	LabelKeyServiceNode        = LabelKeyKubernetesNodeRole + "/lattice-service"
	LabelKeyNodeRoleNodePool   = LabelKeyKubernetesNodeRole + "/lattice-node-pool"

	LabelKeyComponentBuildID = "component.build.lattice.mlab.com/id"

	LabelKeyInternalComponent = "component.lattice.mlab.com/internal"

	LabelKeyNodePoolID = "node-pool.lattice.mlab.com/id"

	LabelKeySystemRolloutVersion = "rollout.system.lattice.mlab.com/version"
	LabelKeySystemRolloutBuildID = "rollout.system.lattice.mlab.com/build"

	LabelKeyServiceID         = "service.lattice.mlab.com/id"
	LabelKeyServicePathDomain = "service.lattice.mlab.com/path-domain"

	LabelKeySystemBuildID      = "system.build.lattice.mlab.com/id"
	LabelKeySystemBuildVersion = "system.build.lattice.mlab.com/version"
	LabelKeySystemVersion      = "system.lattice.mlab.com/version"
)
