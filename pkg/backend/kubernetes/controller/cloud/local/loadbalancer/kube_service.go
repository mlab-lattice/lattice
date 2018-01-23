package loadbalancer

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncLoadBalancerKubeService(loadBalancer *latticev1.LoadBalancer) (*corev1.Service, error) {
	name := kubeutil.GetKubeServiceNameForLoadBalancer(loadBalancer.Name)
	kubeService, err := c.kubeServiceLister.Services(loadBalancer.Namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		return c.createNewKubeService(loadBalancer)
	}

	return kubeService, nil
}

func (c *Controller) createNewKubeService(loadBalancer *latticev1.LoadBalancer) (*corev1.Service, error) {
	name := kubeutil.GetKubeServiceNameForLoadBalancer(loadBalancer.Name)

	spec, err := c.kubeServiceSpec(loadBalancer)
	if err != nil {
		return nil, err
	}

	kubeService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(loadBalancer, controllerKind)},
		},
		Spec: spec,
	}

	// TODO: handle Conflict/AlreadyExists due to slow cache
	return c.kubeClient.CoreV1().Services(loadBalancer.Namespace).Create(kubeService)
}

func (c *Controller) kubeServiceSpec(loadBalancer *latticev1.LoadBalancer) (corev1.ServiceSpec, error) {
	service, err := c.serviceLister.Services(loadBalancer.Namespace).Get(loadBalancer.Name)
	if err != nil {
		return corev1.ServiceSpec{}, err
	}

	var ports []corev1.ServicePort

	for component, componentPorts := range service.Spec.Ports {
		for _, componentPort := range componentPorts {
			if componentPort.Public {
				ports = append(ports, corev1.ServicePort{
					// FIXME: need a better naming scheme
					Name: fmt.Sprintf("%v-%v", component, componentPort.Name),
					Port: componentPort.Port,
				})
			}
		}
	}

	labels := deploymentLabels(service.Name)
	spec := corev1.ServiceSpec{
		Selector: labels,
		Type:     corev1.ServiceTypeNodePort,
		Ports:    ports,
	}
	return spec, nil
}

// FIXME: abstract this out into kubeutils or something
func deploymentLabels(serviceName string) map[string]string {
	return map[string]string{
		kubeconstants.LabelKeyServiceID: serviceName,
	}
}
