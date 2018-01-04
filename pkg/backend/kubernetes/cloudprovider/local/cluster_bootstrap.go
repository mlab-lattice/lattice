package local

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func (cp *DefaultLocalCloudProvider) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	cp.bootstrapDNS(resources)

	for _, daemonSet := range resources.DaemonSets {
		template := cp.TransformPodTemplateSpec(&daemonSet.Spec.Template)
		daemonSet.Spec.Template = *template
	}
}

func (cp *DefaultLocalCloudProvider) bootstrapDNS(resources *clusterbootstrapper.ClusterResources) {

	namespace := kubeutil.InternalNamespace(cp.ClusterID)

	controllerArgs := []string{}
	controllerArgs = append(controllerArgs, cp.Options.DNSServer.DNSControllerArgs...)

	dnsmasqArgs := []string{}
	dnsmasqArgs = append(dnsmasqArgs, cp.Options.DNSServer.DNSServerArgs...)

	labels := map[string]string{
		"key": constants.MasterNodeDNSServer,
	}

	daemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MasterNodeDNSServer,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.MasterNodeDNSServer,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  constants.MasterNodeDNSSController,
							Image: cp.Options.DNSServer.DNSControllerIamge,
							Args:  controllerArgs,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dns-config",
									MountPath: constants.DNSSharedConfigDirectory,
								},
							},
						},
						{
							Name:  constants.MasterNodeDNSServer,
							Image: cp.Options.DNSServer.DNSServerImage,
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
									MountPath: constants.DNSSharedConfigDirectory,
								},
							},
						},
					},
					DNSPolicy: corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountLocalDNS,
					Volumes: []corev1.Volume{
						{
							Name: "dns-config",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: constants.DNSSharedConfigDirectory,
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
			Name:      constants.MasterNodeDNSService,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector:  labels,
			ClusterIP: constants.LocalDNSServerIP,
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
