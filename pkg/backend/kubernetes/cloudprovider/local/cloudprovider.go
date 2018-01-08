package local

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/golang/glog"
)

const (
	workDirectoryVolumeHostPathPrefix = "/data/component-builder"

	MasterNodeDNSSController = "local-dns-controller"
	MasterNodeDNSServer      = "local-dnsmasq-server"
	MasterNodeDNSService     = "local-dns-service"

	ServiceAccountLocalDNS = "local-dns"

	DockerImageLocalDNSController = "kubernetes-local-dns"
	DockerImageLocalDNSServer     = "gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.7"

	LocalDNSServerIP = "10.96.0.53"

	DNSSharedConfigDirectory = "/etc/dns-config/"
	DNSHostsFile             = "hosts"
	DnsmasqConfigFile        = "dnsmasq.conf"

	LocaldnsFinalizer = "localdns-finalizer"
)

type Options struct {
	IP  string
	DNS *DNSOptions
}

type CloudProvider interface {
	IP() string
}

func NewLocalCloudProvider(clusterID types.ClusterID, options *Options) *DefaultLocalCloudProvider {
	cp := &DefaultLocalCloudProvider{
		ClusterID: clusterID,
		ip:        options.IP,
		DNS:       &DNSOptions{},
	}

	if options.DNS != nil {
		cp.DNS = &DNSOptions{
			DNSControllerImage: options.DNS.DNSControllerImage,
			DNSControllerArgs:  options.DNS.DNSControllerArgs,
			DNSServerImage:     options.DNS.DNSServerImage,
			DNSServerArgs:      options.DNS.DNSServerArgs,
		}
	}

	return cp
}

type DefaultLocalCloudProvider struct {
	ClusterID types.ClusterID
	ip        string
	DNS       *DNSOptions
}

type DNSOptions struct {
	DNSServerImage     string
	DNSServerArgs      []string
	DNSControllerImage string
	DNSControllerArgs  []string
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
	spec.Template.Spec.DNSConfig.Nameservers = []string{LocalDNSServerIP}

	found := false

	for idx, nameserver := range spec.Template.Spec.DNSConfig.Nameservers {
		if nameserver == LocalDNSServerIP {
			// Nameserver already present, so no need to update
			found = true

			if idx != 0 {
				glog.Warningf("Local DNS server found, but not as the first nameserver. This will not be modified...")
			}
		}
	}

	if !found {
		// Add the DNS server IP as the first nameserver.
		spec.Template.Spec.DNSConfig.Nameservers = append([]string{LocalDNSServerIP}, spec.Template.Spec.DNSConfig.Nameservers...)
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
