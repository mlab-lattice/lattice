package types

type ServiceBuildID string
type ServiceBuildState string

type ServiceBuild struct {
	ID    ServiceBuildID    `json:"id"`
	State ServiceBuildState `json:"state"`

	// Components maps the component name to the build for that component.
	Components map[string]ComponentBuild `json:"componentBuilds"`
}
