package provisioner

type Interface interface {
	Provision(name, url string) error
	Address(name string) (string, error)
	Deprovision(name string) error
}
