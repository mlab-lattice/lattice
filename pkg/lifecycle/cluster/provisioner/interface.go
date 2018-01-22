package provisioner

type Interface interface {
	Provision(name, url string) (clusterAddress string, err error)
	Deprovision(name string, force bool) error
}
