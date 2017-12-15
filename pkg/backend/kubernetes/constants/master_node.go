package constants

const (
	MasterNodeComponentLatticeControllerManager = "lattice-controller-manager"
	MasterNodeComponentManagerAPI               = "manager-api"
	// TODO :: Need to check that this beloongs in master node.
	MasterNodeDNSServer							= "local-dns-server"

	masterNodeLabel   = "node.master.lattice.mlab.com"
	MasterNodeLabelID = masterNodeLabel + "/id"

	MasterNodeLabelComponent = masterNodeLabel + "/component"
)
