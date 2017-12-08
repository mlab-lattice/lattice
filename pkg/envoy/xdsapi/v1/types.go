package v1

type Service struct {
	EgressPort  int32
	Components  map[string]Component
	IPAddresses []string
}

type Component struct {
	// Ports maps the Component's ports to their envoy ports.
	Ports map[int32]int32
}
