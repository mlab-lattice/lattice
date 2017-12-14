package local

import (

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
)

func (b *DefaultBootstrapper) seedDNS() ([]interface{}, error) {

	// Create some constants to refer to my image TBD where they will actually place
	my_const := "value"
	sa_name_const := "value"

	// TODO :: Handle namespace
	namespace := kubeutil.GetFullNamespace("lattice", constants.NamespaceLatticeInternal)
	args := []string{}
	args = append(args, b.Options.LocalComponents.LocalDNS.Args...)
	labels := map[string]string{
		"key" : my_const,
	}

	// Create a daemon set for my image
	localDNSDaemonSet := &appsv1beta2.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1beta2.GroupName + "/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:		my_const,
			Namespace: 	namespace,
			Labels:		labels,
		},
		Spec: appsv1beta2.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   my_const,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  my_const,
							Image: b.Options.LocalComponents.LocalDNS.Image,
							Args:  args,
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: sa_name_const,
					Tolerations: []corev1.Toleration{
						constants.TolerationMasterNode,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &constants.NodeAffinityMasterNode,
					},
				},
			},
		},
	}

	if b.Options.DryRun {
		return []interface{}{localDNSDaemonSet}, nil
	}

	localDNSDaemonSet, err := b.KubeClient.AppsV1beta2().DaemonSets(namespace).Create(localDNSDaemonSet)
	return []interface{}{localDNSDaemonSet}, err
}
