package types

type ServiceBuildID string
type ServiceBuildState string

type ServiceBuild struct {
	ID    ServiceBuildID    `json:"id"`
	State ServiceBuildState `json:"state"`

	// ComponentBuilds maps the component name to the build for that component.
	ComponentBuilds map[string]*ComponentBuild `json:"componentBuilds"`
}

func (sb ServiceBuild) GetRenderMap() map[string]string {
	return map[string]string{
		"ID":    string(sb.ID),
		"State": string(sb.State),
	}
}
