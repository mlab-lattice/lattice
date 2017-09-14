package local

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (lp *Provider) ComponentBuildJobWorkingDirectoryVolumePathPrefix() string {
	return "/data/component-builder"
}
