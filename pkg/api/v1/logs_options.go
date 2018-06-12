package v1

// ContainerLogOptions represents options for retrieving log files
type ContainerLogOptions struct {
	Follow     bool
	Tail       *int64
	Previous   bool
	Since      string
	SinceTime  string
	Timestamps bool
}

func NewContainerLogOptions() *ContainerLogOptions {
	return &ContainerLogOptions{}
}
