package types

type RouteConfig struct {
	VirtualHosts []VirtualHost `json:"virtual_hosts"`
}

type VirtualHost struct {
	Name    string             `json:"name"`
	Domains []string           `json:"domains"`
	Routes  []VirtualHostRoute `json:"routes"`
	// TODO: reexamine other fields
}

type VirtualHostRoute struct {
	Prefix  string `json:"prefix"`
	Cluster string `json:"cluster"`
	// TODO: reexamine other fields
}
