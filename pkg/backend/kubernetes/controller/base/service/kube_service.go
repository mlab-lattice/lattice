package service

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncServiceKubeService(service *crv1.Service) (*corev1.Service, error) {
	name := kubeutil.GetKubeServiceNameForService(service)
	kubeService, err := c.kubeServiceLister.Services(service.Namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		return c.createNewKubeService(service)
	}

	return kubeService, nil
}

func (c *Controller) createNewKubeService(service *crv1.Service) (*corev1.Service, error) {
	name := kubeutil.GetKubeServiceNameForService(service)

	// Create a headless service (https://kubernetes.io/docs/concepts/services-networking/service/#headless-services)
	// so the endpoints collection will be populated
	spec := kubeServiceSpec(service)
	kubeService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(service, controllerKind)},
		},
		Spec: spec,
	}

	// TODO: handle Conflic/AlreadyExists due to slow cache
	return c.kubeClient.CoreV1().Services(service.Namespace).Create(kubeService)
}

func kubeServiceSpec(service *crv1.Service) corev1.ServiceSpec {
	labels := deploymentLabels(service)
	return corev1.ServiceSpec{
		Selector:  labels,
		ClusterIP: corev1.ClusterIPNone,
		Type:      corev1.ServiceTypeClusterIP,
		Ports: []corev1.ServicePort{
			{
				// Temporarily put this meaningless value here.
				// Kubernetes claims to support headless services with
				// no ports but actually does not.
				Port: 12345,
			},
		},
	}
}
