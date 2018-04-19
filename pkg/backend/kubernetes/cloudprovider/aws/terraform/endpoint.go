package terraform

type ExternalNameEndpoint struct {
	Source string `json:"source"`

	Region string `json:"region"`

	ZoneID       string `json:"zone_id"`
	Name         string `json:"name"`
	ExternalName string `json:"external_name"`
}

func NewExternalNameEndpointModule(moduleRoot, region, zoneID, name, externalName string) *ExternalNameEndpoint {
	return &ExternalNameEndpoint{
		Source: moduleRoot + ModulePathEndpointExternalName,

		Region: region,

		ZoneID:       zoneID,
		Name:         name,
		ExternalName: externalName,
	}
}

type IPEndpoint struct {
	Source string `json:"source"`

	Region string `json:"region"`

	ZoneID string `json:"zone_id"`
	Name   string `json:"name"`
	IP     string `json:"ip"`
}

func NewIPEndpointModule(moduleRoot, region, zoneID, name, ip string) *IPEndpoint {
	return &IPEndpoint{
		Source: moduleRoot + ModulePathEndpointIP,

		Region: region,

		ZoneID: zoneID,
		Name:   name,
		IP:     ip,
	}
}
