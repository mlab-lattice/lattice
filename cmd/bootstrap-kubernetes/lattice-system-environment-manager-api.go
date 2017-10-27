package main

import (
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/client-go/kubernetes"
)

func seedLatticeSystemEnvironmentManagerAPI(kubeClientset *kubernetes.Clientset) {
	var dockerRegistry string
	if dev {
		dockerRegistry = localDevDockerRegistry
	} else {
		dockerRegistry = devDockerRegistry
	}

	latticeSystemEnvironmentManagerAPIDaemonSet := &extensionsv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lattice-system-environment-manager-api",
			Namespace: constants.NamespaceInternal,
		},
		Spec: extensionsv1beta1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lattice-system-environment-manager-api",
					Labels: map[string]string{
						"master.lattice.mlab.com/system-environment-manager-api": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "api",
							Image:   dockerRegistry + "/system-environment-manager-rest-api-kubernetes",
							Command: []string{"/app/cmd/rest-api-kubernetes/go_image.binary"},
							Args:    []string{"-port", "80"},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									HostPort:      80,
									ContainerPort: 80,
								},
							},
						},
					},
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountLatticeSystemEnvironmentManagerAPI,
					// Can tolerate the master-node taint in the local case when it's not applied harmlessly
					Tolerations: []corev1.Toleration{
						constants.TolerationMasterNode,
					},
				},
			},
		},
	}

	// FIXME: add NodeSelector for cloud providers
	//switch coretypes.Provider(providerName) {
	//case coreconstants.ProviderLocal:
	//
	//default:
	//	panic("unsupported providerName")
	//}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClientset.
			ExtensionsV1beta1().
			DaemonSets(string(constants.NamespaceInternal)).
			Create(latticeSystemEnvironmentManagerAPIDaemonSet)
	})
}
