package rest

const (
	RootPath = "/v1"

	SystemsPath      = RootPath + "/systems"
	SystemPathFormat = SystemsPath + "/%v"

	BuildsPathFormat = SystemPathFormat + "/builds"
	BuildPathFormat  = BuildsPathFormat + "/%v"

	DeploysPathFormat = SystemPathFormat + "/deploys"
	DeployPathFormat  = DeploysPathFormat + "/%v"

	SystemSecretsPathFormat = SystemPathFormat + "/secrets"
	SystemSecretPathFormat  = SystemSecretsPathFormat + "/%v"

	ServicesPathFormat = SystemPathFormat + "/services"
	ServicePathFormat  = ServicesPathFormat + "/%v"

	TeardownsPathFormat = SystemPathFormat + "/teardowns"
	TeardownPathFormat  = TeardownsPathFormat + "%v"

	VersionsPathFormat = SystemPathFormat + "/versions"
)
