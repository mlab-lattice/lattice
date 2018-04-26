package service

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncKubeService(service *latticev1.Service) error {
	name := kubeutil.GetKubeServiceNameForService(service.Name)
	_, err := c.kubeServiceLister.Services(service.Namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("error getting kube service for %v: %v", service.Description(c.namespacePrefix), err)
		}

		_, err = c.createNewKubeService(service)
		return err
	}

	return nil
}

func (c *Controller) createNewKubeService(service *latticev1.Service) (*corev1.Service, error) {
	name := kubeutil.GetKubeServiceNameForService(service.Name)

	// Create a headless service (https://kubernetes.io/docs/concepts/services-networking/service/#headless-services)
	// so the endpoints collection will be populated
	spec := kubeServiceSpec(service)
	kubeService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*controllerRef(service)},
		},
		Spec: spec,
	}

	result, err := c.kubeClient.CoreV1().Services(service.Namespace).Create(kubeService)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			result, err := c.kubeClient.CoreV1().Services(service.Namespace).Get(kubeService.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					err := fmt.Errorf(
						"could not create kube service %v for %v because it already exists, but it does not exist",
						kubeService.Name,
						service.Description(c.namespacePrefix),
					)
					return nil, err
				}

				err := fmt.Errorf(
					"error getting kube service %v for %v: %v",
					kubeService.Name, service.Description(c.namespacePrefix),
					err,
				)
				return nil, err
			}

			return result, nil
		}

		err := fmt.Errorf(
			"error creating kube service %v for %v: %v",
			kubeService.Name, service.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return result, nil
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
