package block

type DockerImage struct {
	Registry   string `json:"registry"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

// Validate implements Interface
func (di *DockerImage) Validate(interface{}) error {
	return nil
}
