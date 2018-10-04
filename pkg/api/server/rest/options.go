package rest

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/authentication/authenticator"
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
	Token             authenticator.Token
	OIDCIssuerURL     string
	OIDCClientID      string
	OIDCUsernameClaim string
}
