package rest

const (
	RootPath = "/v1"

	SystemsPath      = RootPath + "/systems"
	SystemPathFormat = SystemsPath + "/%v"

	BuildsPathFormat = SystemPathFormat + "/builds"
	BuildPathFormat  = BuildsPathFormat + "/%v"

	DeploysPathFormat = SystemPathFormat + "/deploys"
	DeployPathFormat  = DeploysPathFormat + "/%v"

	SecretsPathFormat = SystemPathFormat + "/secrets"
	SecretPathFormat  = SecretsPathFormat + "/%v"

	ServicesPathFormat = SystemPathFormat + "/servicesS"
	ServicePathFormat  = ServicesPathFormat + "/%v"

	TeardownsPathFormat = SystemPathFormat + "/teardowns"
	TeardownPathFormat  = TeardownsPathFormat + "%v"

	VersionsPathFormat = SystemPathFormat + "/versions"
)
