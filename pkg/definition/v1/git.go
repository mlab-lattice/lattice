package v1

import "github.com/mlab-lattice/lattice/pkg/definition/tree"

type GitRepository struct {
	URL string `json:"url"`

	Branch  *string `json:"branch,omitempty"`
	Commit  *string `json:"commit,omitempty"`
	Tag     *string `json:"tag,omitempty"`
	Version *string `json:"version,omitempty"`

	SSHKey *tree.PathSubcomponent `json:"ssh_key,omitempty"`
}
