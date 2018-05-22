package rest

const (
	RootPath = "/v1"

	SystemsPath      = RootPath + "/systems"
	SystemPathFormat = SystemsPath + "/%v"

	BuildsPathFormat    = SystemPathFormat + "/builds"
	BuildPathFormat     = BuildsPathFormat + "/%v"
	BuildLogsPathFormat = BuildPathFormat + "/logs"
	BuildLogPathFormat  = BuildLogsPathFormat + "/%v"

	DeploysPathFormat = SystemPathFormat + "/deploys"
	DeployPathFormat  = DeploysPathFormat + "/%v"

	NodePoolsPathFormat = SystemPathFormat + "/node-pools"
	NodePoolPathFormat  = NodePoolsPathFormat + "/%v"

	SystemSecretsPathFormat = SystemPathFormat + "/secrets"
	SystemSecretPathFormat  = SystemSecretsPathFormat + "/%v"

	ServicesPathFormat = SystemPathFormat + "/services"
	ServicePathFormat  = ServicesPathFormat + "/%v"

	TeardownsPathFormat = SystemPathFormat + "/teardowns"
	TeardownPathFormat  = TeardownsPathFormat + "%v"

	VersionsPathFormat = SystemPathFormat + "/versions"
)
