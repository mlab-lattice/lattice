package rest

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication/authenticator"
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
	LegacyAPIAuthKey string
	Token            authenticator.Token
}
