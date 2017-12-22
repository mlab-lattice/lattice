package local

import (
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func (cp *DefaultLocalCloudProvider) seedDNS(resources *clusterbootstrapper.ClusterResources) {

	// TODO :: Handle namespace
	namespace := kubeutil.InternalNamespace("lattice")

	controller_args := []string{"--provider", cp.Provider, "--cluster-id", string(cp.ClusterID)}
	controller_args = append(controller_args, cp.Options.LocalComponents.LocalDNSController.Args...)

	server_args := []string{}
	server_args = append(server_args, cp.Options.LocalComponents.LocalDNSServer.Args...)

	labels := map[string]string{
		"key" : constants.MasterNodeDNSServer,
	}

	localDNSDaemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:		constants.MasterNodeDNSServer,
			Namespace: 	namespace,
			Labels:		labels,
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
							Image: cp.Options.LocalComponents.LocalDNSController.Image,
							Args:  controller_args,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dns-config",
									MountPath: "/etc/dns-config/",
								},
							},
						},
						{
							Name:  constants.MasterNodeDNSServer,
							Image: cp.Options.LocalComponents.LocalDNSServer.Image,
							Args:  server_args,
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

	resources.DaemonSets = append(resources.DaemonSets, localDNSDaemonSet)

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

	resources.Services = append(resources.Services, localDNSService)
}
