package main

import (
	"time"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/client-go/kubernetes"
)

func seedEnvoyXdsApi(kubeClientset *kubernetes.Clientset) {
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
							Name:  "envoy-xds-api",
							Image: "bazel/cmd/kubernetes-per-node-rest:go_image",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									HostPort:      8080,
									ContainerPort: 8080,
								},
							},
						},
					},
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountEnvoyXdsApi,
				},
			},
		},
	}

	err := wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := kubeClientset.ExtensionsV1beta1().DaemonSets(string(coreconstants.UserSystemNamespace)).Create(envoyApiDaemonSet)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}
}
