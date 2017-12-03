package docker

type RegistryLoginProvider interface {
	GetLoginCredentials(registry string) (username, password string, err error)
}
