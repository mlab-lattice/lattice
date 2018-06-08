package v1

type Service struct {
	EgressPort  int32
	Containers  map[string]Container
	IPAddresses []string
}

type Container struct {
	// Ports maps the Sidecar's ports to their envoy ports.
	Ports map[int32]int32
}
