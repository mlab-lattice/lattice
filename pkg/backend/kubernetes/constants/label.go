package constants

const (
	LabelKeyKubernetesNodeRole   = "node-role.kubernetes.io"
	LabelKeyMasterNode           = LabelKeyKubernetesNodeRole + "/lattice-master"
	LabelKeyBuildNode            = LabelKeyKubernetesNodeRole + "/lattice-build"
	LabelKeyServiceNode          = LabelKeyKubernetesNodeRole + "/lattice-service"
	LabelKeyNodeRoleNodePool     = LabelKeyKubernetesNodeRole + "/lattice-node-pool"
	LabelKeyComponentBuildID     = "component.build.lattice.mlab.com/id"
	LabelKeyInternalComponent    = "component.lattice.mlab.com/internal"
	LabelKeySystemRolloutVersion = "rollout.system.lattice.mlab.com/version"
	LabelKeySystemRolloutBuildID = "rollout.system.lattice.mlab.com/build"
	LabelKeyDeploymentServiceID  = "service.lattice.mlab.com/id"
	LabelKeySystemBuildVersion   = "system.build.lattice.mlab.com/version"
	LabelKeySystemVersion        = "system.lattice.mlab.com/version"
)
