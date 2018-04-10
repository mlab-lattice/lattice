package local

import (
	systembootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type SystemBootstrapperOptions struct {
}

func NewSystemBootstrapper(options *SystemBootstrapperOptions) *DefaultLocalSystemBootstrapper {
	return &DefaultLocalSystemBootstrapper{}
}

type DefaultLocalSystemBootstrapper struct {
}

func (cp *DefaultLocalSystemBootstrapper) BootstrapSystemResources(resources *systembootstrapper.SystemResources) {
	for _, daemonSet := range resources.DaemonSets {
		template := transformPodTemplateSpec(&daemonSet.Spec.Template)
		daemonSet.Spec.Template = *template
	}
}

func ParseSystemBootstrapperFlags(vars []string) *SystemBootstrapperOptions {
	return &SystemBootstrapperOptions{}
}

func SystemBootstrapperFlags() (cli.Flags, *SystemBootstrapperOptions) {
	return nil, &SystemBootstrapperOptions{}
}
