package block

type Interface interface {
	Validate(interface{}) error
}
