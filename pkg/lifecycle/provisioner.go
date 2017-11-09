package lifecycle

type Provisioner interface {
	Provision(name, url string) error
	Address(name string) (string, error)
	Deprovision(name string) error
}
