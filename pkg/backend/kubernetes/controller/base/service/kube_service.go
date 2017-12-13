package service

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncServiceKubeService(service *crv1.Service) (*corev1.Service, error) {
	name := kubeutil.GetKubeServiceNameForService(service)
	return c.kubeClient.CoreV1().Services(service.Namespace).Get(name, metav1.GetOptions{})
}
