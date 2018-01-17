package aws

const (
	modulePathRoot                           = "/aws"
	modulePathEndpoint                       = modulePathRoot + "/endpoint"
	modulePathEndpointExternalName           = modulePathEndpoint + "/external-name"
	modulePathEndpointIP                     = modulePathEndpoint + "/ip"
	modulePathLoadBalancer                   = modulePathRoot + "/load-balancer"
	modulePathApplicationLoadBalancer        = modulePathLoadBalancer + "/application"
	modulePathNodePool                       = modulePathRoot + "/node-pool"
	modulePathMasterNode                     = modulePathRoot + "/master-node"
	modulePathMasterNodeEtcdVolumeAttachment = modulePathMasterNode + "/etcd-volume-attachment"
	modulePathMasterNodeDNS                  = modulePathMasterNode + "/dns"
)
