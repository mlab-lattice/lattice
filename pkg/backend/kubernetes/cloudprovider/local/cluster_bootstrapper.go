package local

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
)

type ClusterBootstrapperOptions struct {
	IP string
}

func NewClusterBootstrapper(options *ClusterBootstrapperOptions) *DefaultLocalClusterBootstrapper {
	return &DefaultLocalClusterBootstrapper{
		ip: options.IP,
	}
}

type DefaultLocalClusterBootstrapper struct {
	ip string
}

func (cp *DefaultLocalClusterBootstrapper) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	for _, daemonSet := range resources.DaemonSets {
		template := transformPodTemplateSpec(&daemonSet.Spec.Template)

		if daemonSet.Name == kubeconstants.MasterNodeComponentLatticeControllerManager {
			template.Spec.Containers[0].Args = append(
				template.Spec.Containers[0].Args,
				"--cloud-provider-var", fmt.Sprintf("cluster-ip=%v", cp.ip),
			)
		}

		daemonSet.Spec.Template = *template
	}
}
