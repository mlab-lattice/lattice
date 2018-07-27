package rest

const (
	RootPath = "/v1"

	SystemsPath      = RootPath + "/systems"
	SystemPathFormat = SystemsPath + "/%v"

	BuildsPathFormat    = SystemPathFormat + "/builds"
	BuildPathFormat     = BuildsPathFormat + "/%v"
	BuildLogsPathFormat = BuildPathFormat + "/logs"

	DeploysPathFormat = SystemPathFormat + "/deploys"
	DeployPathFormat  = DeploysPathFormat + "/%v"

	NodePoolsPathFormat = SystemPathFormat + "/node-pools"
	NodePoolPathFormat  = NodePoolsPathFormat + "/%v"

	SystemSecretsPathFormat = SystemPathFormat + "/secrets"
	SystemSecretPathFormat  = SystemSecretsPathFormat + "/%v"

	JobsPathFormat    = SystemPathFormat + "/jobs"
	JobPathFormat     = JobsPathFormat + "/%v"
	JobLogsPathFormat = JobPathFormat + "/logs"

	ServicesPathFormat    = SystemPathFormat + "/services"
	ServicePathFormat     = ServicesPathFormat + "/%v"
	ServiceLogsPathFormat = ServicePathFormat + "/logs"

	TeardownsPathFormat = SystemPathFormat + "/teardowns"
	TeardownPathFormat  = TeardownsPathFormat + "%v"

	VersionsPathFormat = SystemPathFormat + "/versions"
)
