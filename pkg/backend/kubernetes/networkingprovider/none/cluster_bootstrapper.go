package none

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
)

type ClusterBootstrapperOptions struct {
}

func NewClusterBootstrapper(options *ClusterBootstrapperOptions) *DefaultClusterBootstrapper {
	return &DefaultClusterBootstrapper{}
}

type DefaultClusterBootstrapper struct {
}

func (np *DefaultClusterBootstrapper) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	for _, daemonSet := range resources.DaemonSets {
		if daemonSet.Name == kubeconstants.MasterNodeComponentLatticeControllerManager {
			daemonSet.Spec.Template.Spec.Containers[0].Args = append(
				daemonSet.Spec.Template.Spec.Containers[0].Args,
				"--networking-provider", None,
			)
		}
	}
}
