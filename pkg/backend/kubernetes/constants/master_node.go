package constants

const (
    MasterNodeComponentLatticeControllerManager = "lattice-controller-manager"
    MasterNodeComponentManagerAPI               = "manager-api"

    MasterNodeDNSSController			        = "local-dns-controller"
    MasterNodeDNSServer	                        = "local-dnsmasq-server"

    masterNodeLabel   = "node.master.lattice.mlab.com"
    MasterNodeLabelID = masterNodeLabel + "/id"

    MasterNodeLabelComponent = masterNodeLabel + "/component"
)
