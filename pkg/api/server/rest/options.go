package rest

import (
	"io"
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
	BearerTokenFile  io.ReadCloser
}
