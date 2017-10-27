package main

import (
	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/client-go/kubernetes"
)

func seedEnvoyXdsApi(kubeClientset *kubernetes.Clientset) {
	var dockerRegistry string
	if dev {
		dockerRegistry = localDevDockerRegistry
	} else {
		dockerRegistry = devDockerRegistry
	}

	// Create envoy-xds-api daemon set
	envoyApiDaemonSet := &extensionsv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-xds-api",
			Namespace: string(coreconstants.UserSystemNamespace),
		},
		Spec: extensionsv1beta1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-xds-api",
					Labels: map[string]string{
						"envoy.lattice.mlab.com/xds-api": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "envoy-xds-api",
							Image:           dockerRegistry + "/envoy-xds-api-kubernetes-per-node-rest",
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
					ServiceAccountName: constants.ServiceAccountEnvoyXdsApi,
				},
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			ExtensionsV1beta1().
			DaemonSets(string(coreconstants.UserSystemNamespace)).
			Create(envoyApiDaemonSet)
	})
}
