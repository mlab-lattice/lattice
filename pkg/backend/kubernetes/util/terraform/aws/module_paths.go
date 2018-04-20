package aws

const (
	modulePathRoot                    = "/aws"
	modulePathEndpoint                = modulePathRoot + "/endpoint"
	modulePathEndpointExternalName    = modulePathEndpoint + "/external-name"
	modulePathEndpointIP              = modulePathEndpoint + "/ip"
	modulePathLoadBalancer            = modulePathRoot + "/load-balancer"
	modulePathApplicationLoadBalancer = modulePathLoadBalancer + "/application"
	modulePathNodePool                = modulePathRoot + "/node-pool"
)
