package rest

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication"
)

type ServerOptions struct {
	AuthOptions *ServerAuthOptions
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		AuthOptions: &ServerAuthOptions{},
	}
}

type ServerAuthOptions struct {
	LegacyApiAuthKey string
	Token            authentication.Token
}
