package main

import (
	"fmt"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
)

func seedEnvoyXdsApi(kubeClientset *kubernetes.Clientset) {
	fmt.Println("Seeding envoy xds api...")

	var dockerRegistry string
	if dev {
		dockerRegistry = localDevDockerRegistry
	} else {
		dockerRegistry = devDockerRegistry
	}

	// Create envoy-xds-api daemon set
	ds := &appsv1beta2.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-xds-api",
			Namespace: string(coreconstants.UserSystemNamespace),
		},
		Spec: appsv1beta2.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-xds-api",
					Labels: map[string]string{
						"envoy.lattice.mlab.com/xds-api": "true",
					},
				},
				Spec: corev1.PodSpec{
					// FIXME: add service-node toleration
					Containers: []corev1.Container{
						{
							Name:  "envoy-xds-api",
							Image: dockerRegistry + "/envoy-xds-api-kubernetes-per-node-rest",
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
		return kubeClientset.AppsV1beta2().DaemonSets(ds.Namespace).Create(ds)
	})
}
