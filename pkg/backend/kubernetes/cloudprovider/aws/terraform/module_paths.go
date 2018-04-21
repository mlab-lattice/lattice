package terraform

const (
	ModulePathRoute53                 = "/route53"
	ModulePathRoute53Record           = ModulePathRoute53 + "/record"
	ModulePathLoadBalancer            = "/load-balancer"
	ModulePathApplicationLoadBalancer = ModulePathLoadBalancer + "/application"
	ModulePathNodePool                = "/node-pool"
)
