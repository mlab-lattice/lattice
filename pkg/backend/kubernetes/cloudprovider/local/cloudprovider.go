package local

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"

	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/golang/glog"
    "github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
    "github.com/mlab-lattice/system/pkg/types"
)

func NewLocalCloudProvider(clusterID types.ClusterID, providerName string, options cloudprovider.CloudProviderOptions) *DefaultLocalCloudProvider {
	return &DefaultLocalCloudProvider{
	    Options:    options,
        ClusterID:  clusterID,
        Provider:   providerName,
    }
}

type DefaultLocalCloudProvider struct {
    Options     cloudprovider.CloudProviderOptions
    ClusterID   types.ClusterID
    Provider    string
}

func (cp *DefaultLocalCloudProvider) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
    cp.seedDNS(resources)

	for _, daemonSet := range resources.DaemonSets {
		template := cp.TransformPodTemplateSpec(&daemonSet.Spec.Template)
		daemonSet.Spec.Template = *template
	}
}

func (cp *DefaultLocalCloudProvider) BootstrapSystemResources(resources *systembootstrapper.SystemResources) {
	for _, daemonSet := range resources.DaemonSets {
		template := cp.TransformPodTemplateSpec(&daemonSet.Spec.Template)
		daemonSet.Spec.Template = *template
	}
}

func (cp *DefaultLocalCloudProvider) TransformPodTemplateSpec(template *corev1.PodTemplateSpec) *corev1.PodTemplateSpec {
	template = template.DeepCopy()
	template.Spec.Affinity = nil

	return template
}

func (cp *DefaultLocalCloudProvider) TransformComponentBuildJobSpec(spec *batchv1.JobSpec) *batchv1.JobSpec {
	spec = spec.DeepCopy()

	spec.Template.Spec.Affinity = nil
	return spec
}

func (cp *DefaultLocalCloudProvider) TransformServiceDeploymentSpec(service *crv1.Service, spec *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	spec = spec.DeepCopy()
	spec.Template.Spec.Affinity = nil

	ndotsValue := "15"

	DNSConfig := corev1.PodDNSConfig{
	    Nameservers: []string{constants.LocalDNSServerIP},
	    Options: []corev1.PodDNSConfigOption{
	        {
                Name: "ndots",
                Value: &ndotsValue,
            },
        },
    }

	if spec.Template.Spec.DNSConfig == nil {
		spec.Template.Spec.DNSConfig = &DNSConfig
	} else {
		found := false

		for k,v := range spec.Template.Spec.DNSConfig.Nameservers {
			if v == constants.LocalDNSServerIP {
				// Nameserver already present, so no need to update
				found = true

				if k != 0 {
					glog.Warningf("Local DNS server found, but not as the first nameserver... ")
				}
			}
		}

		if !found {
			// Add the DNS server IP as the first nameserver.
			spec.Template.Spec.DNSConfig.Nameservers = append([]string{constants.LocalDNSServerIP}, spec.Template.Spec.DNSConfig.Nameservers...)
		}

		glog.V(4).Infof("Updated nameservers: %v", spec.Template.Spec.DNSConfig.Nameservers)
	}

    spec.Template.Spec.DNSPolicy = corev1.DNSNone

	return spec
}

func (cp *DefaultLocalCloudProvider) IsDeploymentSpecUpdated(
	service *crv1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string, *appsv1.DeploymentSpec) {
	// make a copy of the desired spec, and set the affinity to be the affinity
	// in untransformed
	spec := desired.DeepCopy()
	spec.Template.Spec.Affinity = untransformed.Template.Spec.Affinity

	return true, "", spec
}
