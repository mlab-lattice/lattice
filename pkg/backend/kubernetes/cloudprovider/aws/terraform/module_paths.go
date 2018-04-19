package terraform

const (
	ModulePathEndpoint                = "/endpoint"
	ModulePathEndpointExternalName    = ModulePathEndpoint + "/external-name"
	ModulePathEndpointIP              = ModulePathEndpoint + "/ip"
	ModulePathLoadBalancer            = "/load-balancer"
	ModulePathApplicationLoadBalancer = ModulePathLoadBalancer + "/application"
	ModulePathNodePool                = "/node-pool"
)
