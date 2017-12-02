package block

import "errors"

type GitRepository struct {
	Url    string  `json:"url"`
	Tag    *string `json:"tag,omitempty"`
	Commit *string `json:"commit,omitempty"`
}

// Implement Interface
func (gr *GitRepository) Validate(interface{}) error {
	// TODO: validate url here

	if gr.Tag == nil && gr.Commit == nil {
		return errors.New("must specify either tag or commit")
	}

	if gr.Tag != nil && gr.Commit != nil {
		return errors.New("cannot specify both tag and commit")
	}

	return nil
}
