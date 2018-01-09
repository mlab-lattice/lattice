package local

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (cp *DefaultLocalCloudProvider) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	cp.bootstrapClusterDNS(resources)

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

func (cp *DefaultLocalCloudProvider) bootstrapClusterDNS(resources *clusterbootstrapper.ClusterResources) {

	namespace := kubeutil.InternalNamespace(cp.ClusterID)

	clusterIDParam := "clusterID=" + string(cp.ClusterID)
	controllerArgs := []string{clusterIDParam}
	controllerArgs = append(controllerArgs, cp.DNS.DNSControllerArgs...)

	dnsmasqArgs := []string{}
	dnsmasqArgs = append(dnsmasqArgs, cp.DNS.DNSServerArgs...)

	labels := map[string]string{
		"key": MasterNodeDNSServer,
	}

	daemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      MasterNodeDNSServer,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   MasterNodeDNSServer,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  MasterNodeDNSSController,
							Image: cp.DNS.DNSControllerImage,
							Args:  controllerArgs,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dns-config",
									MountPath: DNSSharedConfigDirectory,
								},
							},
						},
						{
							Name:  MasterNodeDNSServer,
							Image: cp.DNS.DNSServerImage,
							Args:  dnsmasqArgs,
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
									MountPath: DNSSharedConfigDirectory,
								},
							},
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: ServiceAccountLocalDNS,
					Volumes: []corev1.Volume{
						{
							Name: "dns-config",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: DNSSharedConfigDirectory,
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
			Name:      MasterNodeDNSService,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector:  labels,
			ClusterIP: LocalDNSServerIP,
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
			Name: MasterNodeDNSService,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice endpoints
			{
				APIGroups: []string{crv1.GroupName},
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
			Name:      ServiceAccountLocalDNS,
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
			Name: MasterNodeDNSService,
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
