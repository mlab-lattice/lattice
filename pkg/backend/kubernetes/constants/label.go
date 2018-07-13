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
)
