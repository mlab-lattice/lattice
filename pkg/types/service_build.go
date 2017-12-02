package types

type ServiceBuildID string
type ServiceBuildState string

type ServiceBuild struct {
	ID    ServiceBuildID    `json:"id"`
	State ServiceBuildState `json:"state"`

	// ComponentBuilds maps the component name to the build for that component.
	ComponentBuilds map[string]*ComponentBuild `json:"componentBuilds"`
}
