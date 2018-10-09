package deepcopy

type Interface interface {
	DeepCopyInterface() Interface
}
