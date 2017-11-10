package types

type Cluster struct {
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	ConnectTimeoutMs int32     `json:"connect_timeout_ms"`
	LBType           string    `json:"lb_type"`
	ServiceName      string    `json:"service_name"`
	Hosts            []StaticHost `json:"hosts,omitempty"`
	// TODO: reexamine other fields
}
