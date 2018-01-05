package aws

type ExternalNameEndpoint struct {
	Source string `json:"source"`

	Region string `json:"region"`

	ZoneID       string `json:"zone_id"`
	Name         string `json:"name"`
	ExternalName string `json:"external_name"`
}

func NewExternalNameEndpointModule(accountID, region, zoneID, name, externalName string) *ExternalNameEndpoint {
	return &ExternalNameEndpoint{
		Source: modulePathEndpointExternalName,

		Region: region,

		ZoneID:       zoneID,
		Name:         name,
		ExternalName: externalName,
	}
}

type IPEndpoint struct {
	Source string `json:"source"`

	AWSAccountID string `json:"aws_account_id"`
	Region       string `json:"region"`

	ZoneID string `json:"zone_id"`
	Name   string `json:"name"`
	IP     string `json:"ip"`
}

func NewIPEndpointModule(accountID, region, zoneID, name, ip string) *IPEndpoint {
	return &IPEndpoint{
		Source: modulePathEndpointExternalName,

		AWSAccountID: accountID,
		Region:       region,

		ZoneID: zoneID,
		Name:   name,
		IP:     ip,
	}
}
