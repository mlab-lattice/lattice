package local

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	basebootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper/base"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

const (
	dnsController = "dns-controller"
	dnsmasqNanny  = "dnsmasq-nanny"

	dnsConfigDirectory = "/var/run/lattice"
	dnsHostsFile       = dnsConfigDirectory + "hosts"
	dnsmasqConfigFile  = dnsConfigDirectory + "dnsmasq.conf"
)

type LatticeBootstrapperOptions struct {
	IP  string
	DNS *OptionsDNS
}

type OptionsDNS struct {
	DnsmasqNannyImage string
	DnsmasqNannyArgs  []string
	ControllerImage   string
	ControllerArgs    []string
}

func NewLatticeBootstrapper(latticeID v1.LatticeID, namespacePrefix, internalDNSDomain string, options *LatticeBootstrapperOptions) *DefaultLocalLatticeBootstrapper {
	return &DefaultLocalLatticeBootstrapper{
		LatticeID:         latticeID,
		NamespacePrefix:   namespacePrefix,
		InternalDNSDomain: internalDNSDomain,
		IP:                options.IP,
		DNS:               options.DNS,
	}
}

func LatticeBootstrapperFlags() (cli.Flags, *LatticeBootstrapperOptions) {
	options := &LatticeBootstrapperOptions{
		DNS: &OptionsDNS{},
	}
	flags := cli.Flags{
		"ip": &flags.String{
			Required: true,
			Target:   &options.IP,
		},
		"dns-var": &flags.Embedded{
			Required: true,
			Flags: cli.Flags{
				"dnsmasq-nanny-image": &flags.String{
					Required: true,
					Target:   &options.DNS.DnsmasqNannyImage,
				},
				// the args for dnsmasq nanny contain commas, so use a
				// StringArray so these don't try to be parsed as separate
				// args
				"dnsmasq-nanny-args": &flags.StringArray{
					Target: &options.DNS.DnsmasqNannyArgs,
				},
				"controller-image": &flags.String{
					Required: true,
					Target:   &options.DNS.ControllerImage,
				},
				"controller-args": &flags.StringSlice{
					Target: &options.DNS.DnsmasqNannyArgs,
				},
			},
		},
	}
	return flags, options
}

type DefaultLocalLatticeBootstrapper struct {
	LatticeID         v1.LatticeID
	NamespacePrefix   string
	InternalDNSDomain string
	IP                string
	DNS               *OptionsDNS
}

func (cp *DefaultLocalLatticeBootstrapper) BootstrapLatticeResources(resources *bootstrapper.Resources) {
	resources.Config.Spec.CloudProvider.Local = &latticev1.ConfigCloudProviderLocal{}

	var serviceMeshVars []string
	for _, daemonSet := range resources.DaemonSets {
		template := removePodTemplateSpecAffinity(&daemonSet.Spec.Template)

		if daemonSet.Name == kubeconstants.ControlPlaneServiceLatticeControllerManager {
			serviceMeshArg := false
			for _, arg := range template.Spec.Containers[0].Args {
				if serviceMeshArg {
					serviceMeshVars = append(serviceMeshVars, arg)
					serviceMeshArg = false
				}

				if strings.HasPrefix(arg, "--service-mesh") {
					serviceMeshArg = true
					serviceMeshVars = append(serviceMeshVars, arg)
				}
			}

			template.Spec.Containers[0].Args = append(
				template.Spec.Containers[0].Args,
				"--cloud-provider-var", fmt.Sprintf("ip=%v", cp.IP),
			)
		}

		daemonSet.Spec.Template = *template
	}

	cp.bootstrapLatticeDNS(resources, serviceMeshVars)
}

func (cp *DefaultLocalLatticeBootstrapper) bootstrapLatticeDNS(resources *bootstrapper.Resources, serviceMeshVars []string) {
	namespace := kubeutil.InternalNamespace(cp.NamespacePrefix)

	serviceAccount := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsController,
			Namespace: namespace,
		},
	}

	resources.ServiceAccounts = append(resources.ServiceAccounts, serviceAccount)

	clusterRole := &rbacv1.ClusterRole{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: dnsController,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice config
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{"configs"},
				Verbs:     basebootstrapper.ReadVerbs,
			},
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{"addresses"},
				Verbs:     basebootstrapper.ReadVerbs,
			},
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{"services"},
				Verbs:     basebootstrapper.ReadVerbs,
			},
		},
	}

	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: dnsController,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      serviceAccount.Name,
				Namespace: serviceAccount.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		},
	}

	resources.ClusterRoleBindings = append(resources.ClusterRoleBindings, clusterRoleBinding)

	controllerArgs := []string{
		"--namespace-prefix", cp.NamespacePrefix,
		"--internal-dns-domain", cp.InternalDNSDomain,
		"--lattice-id", string(cp.LatticeID),
	}
	controllerArgs = append(controllerArgs, cp.DNS.ControllerArgs...)
	controllerArgs = append(controllerArgs, serviceMeshVars...)

	var dnsmasqNannyArgs []string
	dnsmasqNannyArgs = append(dnsmasqNannyArgs, cp.DNS.DnsmasqNannyArgs...)

	labels := map[string]string{
		"local.cloud-provider.lattice.mlab.com/dns-controller": dnsmasqNanny,
	}

	daemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsController,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   dnsController,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  dnsController,
							Image: cp.DNS.ControllerImage,
							Args:  controllerArgs,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dns-config",
									MountPath: dnsConfigDirectory,
								},
							},
						},
						{
							Name:  dnsmasqNanny,
							Image: cp.DNS.DnsmasqNannyImage,
							Args:  dnsmasqNannyArgs,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 53,
									Name:          "dns",
									Protocol:      "UDP",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dns-config",
									MountPath: dnsConfigDirectory,
								},
							},
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: serviceAccount.Name,
					Volumes: []corev1.Volume{
						{
							Name: "dns-config",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: dnsConfigDirectory,
								},
							},
						},
					},
				},
			},
		},
	}

	resources.DaemonSets = append(resources.DaemonSets, daemonSet)

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsController,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector:  labels,
			ClusterIP: localDNSServerIP,
			Type:      corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "dns-tcp",
					Port:       53,
					TargetPort: intstr.FromInt(53),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "dns-udp",
					Port:       53,
					TargetPort: intstr.FromInt(53),
					Protocol:   corev1.ProtocolUDP,
				},
			},
		},
	}

	resources.Services = append(resources.Services, service)
}
