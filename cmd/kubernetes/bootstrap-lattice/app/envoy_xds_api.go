package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/constants"
	kubeconstants "github.com/mlab-lattice/system/pkg/kubernetes/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
)

func seedEnvoyXdsApi() {
	fmt.Println("Seeding envoy xds api...")

	// Create envoy-xds-api daemon set
	labels := map[string]string{
		"user-system.lattice.mlab.com/envoy-xds-api": "true",
	}
	ds := &appsv1beta2.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-xds-api",
			Namespace: string(constants.UserSystemNamespace),
			Labels:    labels,
		},
		Spec: appsv1beta2.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "envoy-xds-api",
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					// FIXME: add service-node toleration
					Containers: []corev1.Container{
						{
							Name:  "envoy-xds-api",
							Image: getContainerImageFQN(kubeconstants.DockerImageEnvoyXDSAPIRestPerNode),
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									HostPort:      8080,
									ContainerPort: 8080,
								},
							},
						},
					},
					// Use HostNetworking so that envoys can address it just using the hostIp.
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: kubeconstants.ServiceAccountEnvoyXDSAPI,
					Tolerations: []corev1.Toleration{
						{
							Key:      kubeconstants.TaintServiceNode,
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClient.AppsV1beta2().DaemonSets(ds.Namespace).Create(ds)
	})
}
