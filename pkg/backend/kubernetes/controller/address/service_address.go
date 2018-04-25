package address

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) syncServiceAddress(address *latticev1.Address) error {
	if address.Spec.Service == nil {
		return fmt.Errorf("cannot sync service address with no service path")
	}

	c.configLock.RLock()
	defer c.configLock.RUnlock()

	service, err := c.service(address.Namespace, *address.Spec.Service)
	if err != nil {
		return fmt.Errorf("error finding service for %v: %v", address.Description(c.namespacePrefix), err)
	}

	if service == nil {
		if address.Status.State == latticev1.AddressStateStable {
			_, err = c.updateAddressStatus(address, latticev1.AddressStateUpdating, nil, address.Status.Ports)
			return fmt.Errorf("error updating %v to updating: %v", address.Description(c.namespacePrefix), err)
		}

		return nil
	}

	ip, err := c.serviceMesh.ServiceIP(service)
	if err != nil {
		return fmt.Errorf(
			"error getting %v %v ip from service mesh: %v",
			address.Description(c.namespacePrefix),
			service.Description(c.namespacePrefix),
			err,
		)
	}

	path, err := address.PathLabel()
	if err != nil {
		return fmt.Errorf(
			"error getting path label for %v: %v",
			service.Description(c.namespacePrefix),
			err,
		)
	}

	systemID, err := kubeutil.SystemID(c.namespacePrefix, address.Namespace)
	if err != nil {
		return err
	}

	domain := kubeutil.InternalSubdomain(path.ToDomain(), systemID, c.latticeID)
	err = c.cloudProvider.EnsureDNSARecord(c.latticeID, domain, ip)
	if err != nil {
		state := latticev1.AddressStateFailed
		failureInfo := &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error creating DNS A record: %v", err),
			Time:    metav1.Now(),
		}

		c.updateAddressStatus(address, state, failureInfo, address.Status.Ports)
		return fmt.Errorf("error creating service address DNS A record: %v", err)
	}

	if !serviceNeedsAddressLoadBalancer(service) {
		c.updateAddressStatus(address, latticev1.AddressStateStable, nil, nil)
		return nil
	}

	err = c.cloudProvider.EnsureServiceAddressLoadBalancer(c.latticeID, address, service)
	if err != nil {
		state := latticev1.AddressStateFailed
		failureInfo := &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error creating load balancer: %v", err),
			Time:    metav1.Now(),
		}

		_, err := c.updateAddressStatus(address, state, failureInfo, address.Status.Ports)
		if err != nil {
			return fmt.Errorf("error updating %v status: %v", address.Description(c.namespacePrefix), err)
		}

		return nil
	}

	// Add any annotations needed by the cloud provider.
	// Copy annotations so cloud provider doesn't mutate the cache
	annotations := make(map[string]string)
	for k, v := range address.Annotations {
		annotations[k] = v
	}

	err = c.cloudProvider.ServiceAddressLoadBalancerAddAnnotations(c.latticeID, address, service, annotations)
	if err != nil {
		return fmt.Errorf("cloud provider could not get annotations for %v: %v", address.Description(c.namespacePrefix), err)
	}

	address, err = c.updateAddressAnnotations(address, annotations)
	if err != nil {
		return fmt.Errorf("could not update %v annotations: %v", address.Description(c.namespacePrefix), err)
	}

	ports, err := c.cloudProvider.ServiceAddressLoadBalancerPorts(c.latticeID, address, service)
	if err != nil {
		return fmt.Errorf(
			"error getting %v %v load balancer ports: %v",
			address.Description(c.namespacePrefix),
			service.Description(c.namespacePrefix),
			err,
		)
	}

	_, err = c.updateAddressStatus(address, latticev1.AddressStateStable, nil, ports)
	return fmt.Errorf(
		"error updating %v status: %v",
		address.Description(c.namespacePrefix),
		err,
	)
}

func (c *Controller) service(namespace string, path tree.NodePath) (*latticev1.Service, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServicePathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	services, err := c.serviceLister.Services(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	if len(services) > 1 {
		return nil, fmt.Errorf("found multiple services for path %v in namespace %v", path.String(), namespace)
	}

	return services[0], nil
}

func serviceNeedsAddressLoadBalancer(service *latticev1.Service) bool {
	for _, componentPorts := range service.Spec.Ports {
		for _, componentPort := range componentPorts {
			if componentPort.Public {
				return true
			}
		}
	}

	return false
}
