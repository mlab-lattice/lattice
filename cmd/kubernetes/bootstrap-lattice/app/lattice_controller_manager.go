package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

func seedLatticeControllerManager() {
	fmt.Println("Seeding lattice-controller-manager...")

	// TODO: for now we'll make a DaemonSet that runs on all the master nodes (aka all nodes in local)
	//		 and rely on the fact that we don't support multiple master nodes on AWS yet. Eventually we'll
	//		 either have to figure out how to have multiple lattice-controller-managers running (e.g. use leaderelect
	//		 in client-go) or find the best way to ensure there's at most one version of something running (maybe
	//		 StatefulSets?).
	latticeControllerManagerDaemonSet := &extensionsv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MasterNodeComponentLatticeControllerManager,
			Namespace: constants.NamespaceLatticeInternal,
		},
		Spec: extensionsv1beta1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.MasterNodeComponentLatticeControllerManager,
					Labels: map[string]string{
						constants.MasterNodeLabelComponent: constants.MasterNodeComponentLatticeControllerManager,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  constants.MasterNodeComponentLatticeControllerManager,
							Image: getContainerImageFQN(constants.DockerImageLatticeControllerManager),
							Args: []string{
								"-v", "5",
								"-logtostderr",
								"-provider", provider,
							},
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountLatticeControllerManager,
					// Can tolerate the master-node taint in the local case when it's not applied harmlessly
					Tolerations: []corev1.Toleration{
						constants.TolerationMasterNode,
					},
				},
			},
		},
	}

	// FIXME: add NodeSelector for cloud providers
	//switch coretypes.Provider(provider) {
	//case coreconstants.ProviderLocal:
	//
	//default:
	//	panic("unsupported provider")
	//}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClient.
			ExtensionsV1beta1().
			DaemonSets(string(constants.NamespaceLatticeInternal)).
			Create(latticeControllerManagerDaemonSet)
	})
}
