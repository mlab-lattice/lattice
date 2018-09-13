package v1

// ContainerLogOptions represents options for retrieving log files
type ContainerLogOptions struct {
	Follow     bool   `json:"follow,omitempty"`
	Tail       *int64 `json:"tail,omitempty"`
	Previous   bool   `json:"previous,omitempty"`
	Since      string `json:"since,omitempty"`
	SinceTime  string `json:"sinceTime,omitempty"`
	Timestamps bool   `json:"timestamps,omitempty"`
}
