package provider

type Interface interface {
	ComponentBuildJobWorkingDirectoryVolumePathPrefix() string
	ServiceEnvoyConfigDirectoryVolumePath() string
}
