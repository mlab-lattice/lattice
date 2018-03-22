package local

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	dnsPod        = "local-dns"
	dnsController = "local-dns-controller"
	dnsmasqNanny  = "local-dnsmasq-nanny"
	dnsService    = "local-dns-service"

	serviceAccountDNS = "local-dns"
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

func NewLatticeBootstrapper(latticeID types.LatticeID, options *LatticeBootstrapperOptions) *DefaultLocalLatticeBootstrapper {
	return &DefaultLocalLatticeBootstrapper{
		LatticeID: latticeID,
		ip:        options.IP,
		DNS:       options.DNS,
	}
}

func LatticeBootstrapperFlags() (command.Flags, *LatticeBootstrapperOptions) {
	options := &LatticeBootstrapperOptions{
		DNS: &OptionsDNS{},
	}
	flags := command.Flags{
		&command.StringFlag{
			Name:     "ip",
			Required: true,
			Target:   &options.IP,
		},
		&command.EmbeddedFlag{
			Name:     "dns-var",
			Required: true,
			Flags: command.Flags{
				&command.StringFlag{
					Name:     "dnsmasq-nanny-image",
					Required: true,
					Target:   &options.DNS.DnsmasqNannyImage,
				},
				&command.StringSliceFlag{
					Name:   "dnsmasq-nanny-args",
					Target: &options.DNS.DnsmasqNannyArgs,
				},
				&command.StringFlag{
					Name:     "controller-image",
					Required: true,
					Target:   &options.DNS.ControllerImage,
				},
				&command.StringSliceFlag{
					Name:   "controller-args",
					Target: &options.DNS.DnsmasqNannyArgs,
				},
			},
		},
	}
	return flags, options
}

type DefaultLocalLatticeBootstrapper struct {
	LatticeID types.LatticeID
	ip        string
	DNS       *OptionsDNS
}

func (cp *DefaultLocalLatticeBootstrapper) BootstrapLatticeResources(resources *bootstrapper.Resources) {
	cp.bootstrapLatticeDNS(resources)

	for _, daemonSet := range resources.DaemonSets {
		template := transformPodTemplateSpec(&daemonSet.Spec.Template)

		if daemonSet.Name == kubeconstants.MasterNodeComponentLatticeControllerManager {
			template.Spec.Containers[0].Args = append(
				template.Spec.Containers[0].Args,
				"--cloud-provider-var", fmt.Sprintf("ip=%v", cp.ip),
			)
		}

		daemonSet.Spec.Template = *template
	}
}

func (cp *DefaultLocalLatticeBootstrapper) bootstrapLatticeDNS(resources *bootstrapper.Resources) {
	namespace := kubeutil.InternalNamespace(cp.LatticeID)

	controllerArgs := []string{"--lattice-id", string(cp.LatticeID)}
	controllerArgs = append(controllerArgs, cp.DNS.ControllerArgs...)

	dnsmasqNannyArgs := []string{}
	dnsmasqNannyArgs = append(dnsmasqNannyArgs, cp.DNS.DnsmasqNannyArgs...)

	labels := map[string]string{
		"local.cloud-provider.lattice.mlab.com/dns": dnsmasqNanny,
	}

	daemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsPod,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   dnsPod,
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
									MountPath: DNSConfigDirectory,
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
									MountPath: DNSConfigDirectory,
								},
							},
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: serviceAccountDNS,
					Volumes: []corev1.Volume{
						{
							Name: "dns-config",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: DNSConfigDirectory,
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
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsService,
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

	clusterRole := &rbacv1.ClusterRole{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccountDNS,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice endpoints
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{"endpoints"},
				Verbs:     []string{rbacv1.VerbAll},
			},
		},
	}

	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)

	serviceAccount := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountDNS,
			Namespace: namespace,
		},
	}

	resources.ServiceAccounts = append(resources.ServiceAccounts, serviceAccount)

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccountDNS,
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
}
