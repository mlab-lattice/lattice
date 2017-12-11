package local_dns

//TODO :: Should this be called backend
type backend interface {

	//Placeholder
	Ready() bool
	Services() bool
}
