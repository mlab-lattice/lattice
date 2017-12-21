package local

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"

)

func (b *DefaultBootstrapper) seedDNS() ([]interface{}, error) {
	if !b.Options.DryRun {
		fmt.Println("Seeding local DNS server")
	}

	// TODO :: Handle namespace
	namespace := kubeutil.InternalNamespace("lattice")

	controller_args := []string{"--provider", b.Provider, "--cluster-id", string(b.ClusterID)}
	controller_args = append(controller_args, b.Options.LocalComponents.LocalDNSController.Args...)

	server_args := []string{}
	server_args = append(server_args, b.Options.LocalComponents.LocalDNSServer.Args...)

	labels := map[string]string{
		"key" : constants.MasterNodeDNSServer,
	}

	localDNSDaemonSet := &appsv1beta2.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1beta2.GroupName + "/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:		constants.MasterNodeDNSServer,
			Namespace: 	namespace,
			Labels:		labels,
		},
		Spec: appsv1beta2.DaemonSetSpec{
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
							Image: b.Options.LocalComponents.LocalDNSController.Image,
							Args:  controller_args,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dns-config",
									MountPath: "/etc/dns-config/",
								},
							},
						},
						{
							Name:	constants.MasterNodeDNSServer,
							Image:	b.Options.LocalComponents.LocalDNSServer.Image,
							Args:	server_args,
							// TODO :: Ports
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 53,
									Name: "dns",
									Protocol: "UDP",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dns-config",
									MountPath: "/etc/dns-config/",
								},
							},
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					// TODO :: This is default until I know what SA, if any, to use for the DNS.
					ServiceAccountName: constants.ServiceAccountLatticeControllerManager,
					Tolerations: []corev1.Toleration{
						constants.TolerationMasterNode,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &constants.NodeAffinityMasterNode,
					},
					Volumes: []corev1.Volume{
						{
							Name: "dns-config",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/etc/dns-config/",
								},
							},
						},
					},
				},
			},
		},
	}

	if b.Options.DryRun {
		return []interface{}{localDNSDaemonSet}, nil
	}

	localDNSDaemonSet, err := b.KubeClient.AppsV1beta2().DaemonSets(namespace).Create(localDNSDaemonSet)

	localDNSService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:		constants.MasterNodeDNSService,
			Namespace: 	namespace,
			Labels:		labels,
		},
		Spec:corev1.ServiceSpec{
			Selector:labels,
			ClusterIP: constants.LocalDNSServerIP,
			Type:corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:"dns-tcp",
					Port:53,
					TargetPort:intstr.FromInt(53),
					Protocol:corev1.ProtocolTCP,
				},
				{
					Name:"dns-udp",
					Port:53,
					TargetPort:intstr.FromInt(53),
					Protocol:corev1.ProtocolUDP,
				},
			},
		},
	}

	localDNSService, err = b.KubeClient.CoreV1().Services(namespace).Create(localDNSService)

	return []interface{}{localDNSDaemonSet, localDNSService}, err
}
