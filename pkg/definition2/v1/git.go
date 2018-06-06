package v1

type GitRepository struct {
	URL string `json:"url"`

	Branch *string `json:"branch,omitempty"`
	Commit *string `json:"commit,omitempty"`
	Tag    *string `json:"tag,omitempty"`
}
