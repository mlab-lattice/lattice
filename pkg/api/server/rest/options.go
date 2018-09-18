package rest

type ServerOptions struct {
	AuthOptions *ServerAuthOptions
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		AuthOptions: &ServerAuthOptions{},
	}
}

type AuthenticationType string

const (
	AuthTypeLegacy AuthenticationType = "legacy"
)

type ServerAuthOptions struct {
	AuthType         AuthenticationType
	LegacyApiAuthKey string
}
