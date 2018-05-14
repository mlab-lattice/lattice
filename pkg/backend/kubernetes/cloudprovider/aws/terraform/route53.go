package terraform

type Route53Record struct {
	Source string `json:"source"`

	Region string `json:"region"`

	ZoneID string `json:"zone_id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}
