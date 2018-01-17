package local

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/golang/glog"
)

const (
	workDirectoryVolumeHostPathPrefix = "/data/component-builder"

	DockerImageDNSController = "kubernetes-local-dns-controller"
	DockerImageDnsmasqNanny  = "gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.7"

	// This is the default IP for kube-dns
	localDNSServerIP = "10.96.0.53"

	DNSConfigDirectory = "/etc/lattice/local/dns"
	DNSHostsFile       = DNSConfigDirectory + "hosts"
	DnsmasqConfigFile  = DNSConfigDirectory + "dnsmasq.conf"
)

type Options struct {
	IP string
}

type CloudProvider interface {
	IP() string
}

func NewCloudProvider(options *Options) *DefaultLocalCloudProvider {
	return &DefaultLocalCloudProvider{
		ip: options.IP,
	}
}

type DefaultLocalCloudProvider struct {
	ip string
}

func (cp *DefaultLocalCloudProvider) TransformComponentBuildJobSpec(spec *batchv1.JobSpec) *batchv1.JobSpec {
	spec = spec.DeepCopy()
	spec.Template = *transformPodTemplateSpec(&spec.Template)

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
	spec.Template = *transformPodTemplateSpec(&spec.Template)

	found := false

	// This uses DNSNone and supplies the local dnsmasq server as the only nameserver. This is because it
	// is the only way to have names in the node to have priority, whilst still inheriting the clusters
	// dns config. It's hacky, but it's how DNSClusterFirst works aswell:
	// https://github.com/kubernetes/kubernetes/blob/v1.9.0/pkg/kubelet/network/dns/dns.go#L340-L360
	spec.Template.Spec.DNSPolicy = corev1.DNSNone

	for idx, nameserver := range spec.Template.Spec.DNSConfig.Nameservers {
		if nameserver == localDNSServerIP {
			// Nameserver already present, so no need to update
			found = true

			if idx != 0 {
				glog.Warningf("Local DNS server found, but not as the first nameserver. This will not be modified...")
			}

			break
		}
	}

	if !found {
		// Add the DNS server IP as the first nameserver.
		spec.Template.Spec.DNSConfig.Nameservers = append([]string{localDNSServerIP}, spec.Template.Spec.DNSConfig.Nameservers...)
	}

	glog.V(4).Infof("Updated nameservers: %v", spec.Template.Spec.DNSConfig.Nameservers)

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

func (cp *DefaultLocalCloudProvider) IP() string {
	return cp.ip
}

func transformPodTemplateSpec(template *corev1.PodTemplateSpec) *corev1.PodTemplateSpec {
	template = template.DeepCopy()
	template.Spec.Affinity = nil

	return template
}
