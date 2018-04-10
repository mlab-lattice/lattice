package service

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncServiceKubeService(service *latticev1.Service) (*corev1.Service, error) {
	name := kubeutil.GetKubeServiceNameForService(service.Name)
	kubeService, err := c.kubeServiceLister.Services(service.Namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		return c.createNewKubeService(service)
	}

	return kubeService, nil
}

func (c *Controller) createNewKubeService(service *latticev1.Service) (*corev1.Service, error) {
	name := kubeutil.GetKubeServiceNameForService(service.Name)

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

	// TODO: handle Conflict/AlreadyExists due to slow cache
	return c.kubeClient.CoreV1().Services(service.Namespace).Create(kubeService)
}

func kubeServiceSpec(service *latticev1.Service) corev1.ServiceSpec {
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
				// TODO: pending https://github.com/kubernetes/kubernetes/issues/55158
				Port: 12345,
			},
		},
	}
}
