package command

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type Context struct {
	URL    string       `json:"url"`
	System v1.SystemID  `json:"system,omitempty"`
	Auth   *AuthContext `json:"auth"`
}

type AuthContext struct {
	BearerToken *string `json:"bearerToken"`
}
