package provisioner

type Interface interface {
	Provision(name string, initialSystemDefinitionURL *string) (clusterAddress string, err error)
	Deprovision(name string, force bool) error
}
