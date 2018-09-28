package command

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type Context struct {
	URL    string       `json:"lattice"`
	System v1.SystemID  `json:"system,omitempty"`
	Auth   *AuthContext `json:"auth"`
}

type AuthContext struct {
	LegacyApiKey *string `json:"legacyApiKey"`
	BearerToken  *string `json:"bearerToken"`
}
