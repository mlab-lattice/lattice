package provider

type Interface interface {
	ComponentBuildJobWorkingDirectoryVolumePathPrefix() string
	ServiceEnvoyConfigDirectoryVolumePathPrefix() string
}
