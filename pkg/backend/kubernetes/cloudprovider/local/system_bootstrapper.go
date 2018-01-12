package local

import (
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
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
