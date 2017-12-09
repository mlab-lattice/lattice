package constants

const (
	LabelKeyKubernetesNodeRole   = "node-role.kubernetes.io"
	LabelKeyMasterNode           = LabelKeyKubernetesNodeRole + "/lattice-master"
	LabelKeyBuildNode            = LabelKeyKubernetesNodeRole + "/lattice-build"
	LabelKeyServiceNode          = LabelKeyKubernetesNodeRole + "/lattice-service"
	LabelKeyComponentBuildID     = "component.build.lattice.mlab.com/id"
	LabelKeyInternalComponent    = "component.lattice.mlab.com/internal"
	LabelKeySystemRolloutVersion = "rollout.system.lattice.mlab.com/version"
	LabelKeySystemRolloutBuildID = "rollout.system.lattice.mlab.com/build"
	LabelKeyServiceDeployment    = "service.lattice.mlab.com"
	LabelKeySystemBuildVersion   = "system.build.lattice.mlab.com/version"
	LabelKeySystemVersion        = "system.lattice.mlab.com/version"
)
