package constants

const (
	MasterNodeComponentLatticeControllerMaster = "lattice-controller-master"
	MasterNodeComponentManagementApi           = "management-api"

	masterNodeLabel   = "node.master.lattice.mlab.com"
	MasterNodeLabelID = masterNodeLabel + "/id"

	masterNodeLabelComponent                        = masterNodeLabel + "/component"
	MasterNodeLabelComponentLatticeControllerMaster = masterNodeLabelComponent + MasterNodeComponentLatticeControllerMaster
	MasterNodeLabelComponentManagementApi           = masterNodeLabelComponent + MasterNodeComponentManagementApi
)
