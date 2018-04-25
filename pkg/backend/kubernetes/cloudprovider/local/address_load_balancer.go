package local

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

func (cp *DefaultLocalCloudProvider) EnsureServiceAddressLoadBalancer(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
) error {
	if !needsServiceAddressLoadBalancer(service) {
		return nil
	}

	// FIXME: move this out of kubeutil
	kubeServiceName := kubeutil.GetKubeServiceNameForLoadBalancer(address.Name)

	// Try to find the kube service in the cache
	kubeService, err := cp.kubeServiceLister.Services(address.Namespace).Get(kubeServiceName)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		// If it wasn't found, try to create it.
		spec, err := cp.kubeServiceSpec(address, service)
		if err != nil {
			return err
		}

		kubeService = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:            kubeServiceName,
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(address, latticev1.AddressKind)},
			},
			Spec: spec,
		}

		// If it wasn't found in our cache but we couldn't create it due to it already existing, we lost a race and
		// should retrieve it from the API.
		kubeService, err = cp.kubeClient.CoreV1().Services(address.Namespace).Create(kubeService)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return err
			}

			kubeService, err = cp.kubeClient.CoreV1().Services(address.Namespace).Get(kubeServiceName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					return err
				}

				return fmt.Errorf(
					"could not create kube service %v for %v because it already exists, but could not find it",
					kubeServiceName,
					address.Description(cp.namespacePrefix),
				)
			}
		}
	}

	spec, err := cp.kubeServiceSpec(address, service)
	if err != nil {
		return err
	}

	// If the kube service's spec isn't up to date, update it.
	if reflect.DeepEqual(spec, kubeService.Spec) {
		return nil
	}

	// Copy so we don't mutate the shared cache
	kubeService = kubeService.DeepCopy()
	kubeService.Spec = spec

	_, err = cp.kubeClient.CoreV1().Services(address.Namespace).Update(kubeService)
	return err
}

func (cp *DefaultLocalCloudProvider) DestroyServiceAddressLoadBalancer(
	latticeID v1.LatticeID,
	address *latticev1.Address,
) error {
	kubeServiceName := kubeutil.GetKubeServiceNameForLoadBalancer(address.Name)

	err := cp.kubeClient.CoreV1().Services(address.Namespace).Delete(kubeServiceName, nil)
	if err != nil {
		// if the kube service is already deleted then it's not an error
		if !errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (cp *DefaultLocalCloudProvider) ServiceAddressLoadBalancerAddAnnotations(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	annotations map[string]string,
) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) ServiceAddressLoadBalancerPorts(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
) (map[int32]string, error) {
	kubeServiceName := kubeutil.GetKubeServiceNameForLoadBalancer(address.Name)
	kubeService, err := cp.kubeServiceLister.Services(address.Namespace).Get(kubeServiceName)
	if err != nil {
		return nil, err
	}

	ports := make(map[int32]string)
	for _, port := range kubeService.Spec.Ports {
		ports[port.Port] = fmt.Sprintf("%v:%v", cp.IP(), port.NodePort)
	}

	return ports, nil
}

func (cp *DefaultLocalCloudProvider) kubeServiceSpec(address *latticev1.Address, service *latticev1.Service) (corev1.ServiceSpec, error) {
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

	labels := map[string]string{
		latticev1.ServiceIDLabelKey: service.Name,
	}
	spec := corev1.ServiceSpec{
		Selector: labels,
		Type:     corev1.ServiceTypeNodePort,
		Ports:    ports,
	}
	return spec, nil
}

func needsServiceAddressLoadBalancer(service *latticev1.Service) bool {
	for _, componentPorts := range service.Spec.Ports {
		for _, componentPort := range componentPorts {
			if componentPort.Public {
				return true
			}
		}
	}

	return false
}
