package command

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type Context struct {
	Lattice string       `json:"lattice"`
	System  v1.SystemID  `json:"system"`
	Auth    *AuthContext `json:"auth"`
}

type AuthContext struct {
	BearerToken *string `json:"bearerToken"`
}
