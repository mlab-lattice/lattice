package command

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

// Context contains the necessary information to connect to a lattice API server,
// and an optional default system.
type Context struct {
	URL    string       `json:"url"`
	Auth   *AuthContext `json:"auth"`
	System v1.SystemID  `json:"system,omitempty"`
}

// AuthContext contains information about how to authenticate to a lattice API server.
type AuthContext struct {
	BearerToken *string `json:"bearerToken"`
}
