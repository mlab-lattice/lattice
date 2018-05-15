package local

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/util/cli"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
)

const (
	workDirectoryVolumeHostPathPrefix = "/data/component-builder"

	dockerImageDNSController = "kubernetes-local-dns-controller"
	dockerImageDnsmasqNanny  = "gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.7"

	// This is the default IP for kube-dns
	localDNSServerIP = "10.96.0.53"
)

type Options struct {
	IP string
}

func NewOptions(staticOptions *Options, dynamicConfig *latticev1.ConfigCloudProviderLocal) (*Options, error) {
	options := &Options{
		IP: staticOptions.IP,
	}
	return options, nil
}

func NewCloudProvider(
	namespacePrefix string,
	kubeClient kubeclientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	options *Options,
) (*DefaultLocalCloudProvider, error) {
	cp := &DefaultLocalCloudProvider{
		namespacePrefix: namespacePrefix,
		ip:              options.IP,

		kubeClient:        kubeClient,
		kubeNodeLister:    kubeInformerFactory.Core().V1().Nodes().Lister(),
		kubeServiceLister: kubeInformerFactory.Core().V1().Services().Lister(),
	}

	// wait for secondary caches to fill
	if !cache.WaitForCacheSync(
		nil,
		kubeInformerFactory.Core().V1().Nodes().Informer().HasSynced,
		kubeInformerFactory.Core().V1().Services().Informer().HasSynced,
	) {
		return nil, fmt.Errorf("failed to sync caches for aws cloud provider")
	}

	return cp, nil
}

func Flags() (cli.Flags, *Options) {
	options := &Options{}
	flags := cli.Flags{
		&cli.StringFlag{
			Name:     "ip",
			Required: true,
			Target:   &options.IP,
		},
	}
	return flags, options
}

type DefaultLocalCloudProvider struct {
	namespacePrefix string
	ip              string

	kubeClient        kubeclientset.Interface
	kubeNodeLister    corelisters.NodeLister
	kubeServiceLister corelisters.ServiceLister
}

func (cp *DefaultLocalCloudProvider) BootstrapSystemResources(resources *bootstrapper.SystemResources) {
	for _, daemonSet := range resources.DaemonSets {
		template := transformPodTemplateSpec(&daemonSet.Spec.Template)
		daemonSet.Spec.Template = *template
	}
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

func (cp *DefaultLocalCloudProvider) TransformServiceDeploymentSpec(
	service *latticev1.Service,
	spec *appsv1.DeploymentSpec,
) *appsv1.DeploymentSpec {
	spec = spec.DeepCopy()
	spec.Template = *transformPodTemplateSpec(&spec.Template)

	// This uses DNSNone and supplies the local dnsmasq server as the only nameserver. This is because it
	// is the only way to have names in the node to have priority, whilst still inheriting the clusters
	// dns config. It's hacky, but it's how DNSClusterFirst works aswell:
	// https://github.com/kubernetes/kubernetes/blob/v1.9.0/pkg/kubelet/network/dns/dns.go#L340-L360
	spec.Template.Spec.DNSPolicy = corev1.DNSNone

	found := false
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
	service *latticev1.Service,
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
