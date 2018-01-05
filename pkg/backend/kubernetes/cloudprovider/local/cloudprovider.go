package local

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"

	"github.com/golang/glog"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/system/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	workDirectoryVolumeHostPathPrefix = "/data/component-builder"
)

type Options struct {
	IP  string
	DNS *LocalDNSControllerOptions
}

type CloudProvider interface {
	IP() string
}

func NewLocalCloudProvider(clusterID types.ClusterID, options *Options) *DefaultLocalCloudProvider {
	cp := &DefaultLocalCloudProvider{
		ClusterID: clusterID,
		ip:        options.IP,
		Options:   &crv1.ConfigCloudProviderLocal{},
	}

	if options.DNS != nil {
		cp.Options.DNSServer = &crv1.ConfigCloudProviderLocalDNS{
			DNSControllerIamge: options.DNS.DNSControllerImage,
			DNSControllerArgs:  options.DNS.DNSControllerArgs,
			DNSServerImage:     options.DNS.DNSServerImage,
			DNSServerArgs:      options.DNS.DNSServerArgs,
		}
	}

	return cp
}

type DefaultLocalCloudProvider struct {
	ClusterID types.ClusterID
	Options   *crv1.ConfigCloudProviderLocal
	ip        string
}

type LocalDNSControllerOptions struct {
	DNSServerImage     string
	DNSServerArgs      []string
	DNSControllerImage string
	DNSControllerArgs  []string
}

func (cp *DefaultLocalCloudProvider) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	cp.bootstrapDNS(resources)

	for _, daemonSet := range resources.DaemonSets {
		template := cp.transformPodTemplateSpec(&daemonSet.Spec.Template)

		if daemonSet.Name == kubeconstants.MasterNodeComponentLatticeControllerManager {
			template.Spec.Containers[0].Args = append(
				template.Spec.Containers[0].Args,
				"--cloud-provider-var", fmt.Sprintf("cluster-ip=%v", cp.ip),
			)
		}

		daemonSet.Spec.Template = *template
	}
}

func (cp *DefaultLocalCloudProvider) BootstrapSystemResources(resources *systembootstrapper.SystemResources) {
	for _, daemonSet := range resources.DaemonSets {
		template := cp.transformPodTemplateSpec(&daemonSet.Spec.Template)
		daemonSet.Spec.Template = *template
	}
}

func (cp *DefaultLocalCloudProvider) TransformComponentBuildJobSpec(spec *batchv1.JobSpec) *batchv1.JobSpec {
	spec = spec.DeepCopy()
	spec.Template = *cp.transformPodTemplateSpec(&spec.Template)

	return spec
}

func (cp *DefaultLocalCloudProvider) ComponentBuildWorkDirectoryVolumeSource(jobName string) corev1.VolumeSource {
	return corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: workDirectoryVolumeHostPathPrefix + "/" + jobName,
		},
	}
}

func (cp *DefaultLocalCloudProvider) TransformServiceDeploymentSpec(service *crv1.Service, spec *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	spec = spec.DeepCopy()
	spec.Template = *cp.transformPodTemplateSpec(&spec.Template)
	spec.Template.Spec.DNSConfig.Nameservers = []string{constants.LocalDNSServerIP}

	found := false

	for idx, nameserver := range spec.Template.Spec.DNSConfig.Nameservers {
		if nameserver == constants.LocalDNSServerIP {
			// Nameserver already present, so no need to update
			found = true

			if idx != 0 {
				glog.Warningf("Local DNS server found, but not as the first nameserver. This will not be modified...")
			}
		}
	}

	if !found {
		// Add the DNS server IP as the first nameserver.
		spec.Template.Spec.DNSConfig.Nameservers = append([]string{constants.LocalDNSServerIP}, spec.Template.Spec.DNSConfig.Nameservers...)
	}

	glog.V(4).Infof("Updated nameservers: %v", spec.Template.Spec.DNSConfig.Nameservers)

	return spec
}

func (cp *DefaultLocalCloudProvider) transformPodTemplateSpec(template *corev1.PodTemplateSpec) *corev1.PodTemplateSpec {
	template = template.DeepCopy()
	template.Spec.Affinity = nil

	return template
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

func (cp *DefaultLocalCloudProvider) IP() string {
	return cp.ip
}
