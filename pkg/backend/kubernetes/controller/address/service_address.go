package address

import (
	"fmt"

	"github.com/golang/glog"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) syncServiceAddress(address *latticev1.Address) error {
	if address.Spec.Service == nil {
		return fmt.Errorf("cannot sync service address with no service path (%v)", address.Description(c.namespacePrefix))
	}

	c.configLock.RLock()
	defer c.configLock.RUnlock()

	service, err := c.service(address.Namespace, *address.Spec.Service)
	if err != nil {
		return fmt.Errorf("error finding service for %v: %v", address.Description(c.namespacePrefix), err)
	}

	if service == nil {
		if address.Status.State == latticev1.AddressStateStable {
			message := fmt.Sprintf("waiting for service %v", *address.Spec.Service)
			_, err = c.updateAddressStatus(
				address,
				latticev1.AddressStateUpdating,
				&message,
				nil,
				address.Status.Ports,
			)
			return err
		}

		return nil
	}

	ip, annotations, err := c.serviceMesh.WorkloadIP(address, service.Spec.Definition.Ports)
	if err != nil {
		return fmt.Errorf(
			"error getting %v %v IP from service mesh: %v",
			address.Description(c.namespacePrefix),
			service.Description(c.namespacePrefix),
			err,
		)
	}

	address_, err := c.mergeAndUpdateAddressAnnotations(address, annotations)
	if err != nil {
		address_ = address.DeepCopy()
		for k, v := range annotations {
			address_.Annotations[k] = v
		}
		_, err = c.serviceMesh.ReleaseWorkloadIP(address_)
		if err != nil {
			glog.Errorf(
				"Got an error trying to release a service IP lease for %s after failed update: %v",
				service.Name, err)
		}
		return fmt.Errorf(
			"error updating %v address annotations: %v",
			address.Description(c.namespacePrefix),
			err,
		)
	} else {
		address = address_
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
		return fmt.Errorf("error getting system id for %v: %v", address.Description(c.namespacePrefix), err)
	}

	domain := kubeutil.InternalAddressSubdomain(path.ToDomain(), systemID, c.latticeID)
	needsUpdate, err := c.cloudProvider.DNSARecordNeedsUpdate(c.latticeID, domain, ip)
	if err != nil {
		return fmt.Errorf(
			"error checking if DNS A record(s) for %v needs update: %v",
			address.Description(c.namespacePrefix),
			err,
		)
	}

	if needsUpdate {
		message := "updating internal DNS record(s)"
		address, err = c.updateAddressStatus(
			address,
			latticev1.AddressStateUpdating,
			&message,
			nil,
			address.Status.Ports,
		)
		if err != nil {
			return err
		}

		err = c.cloudProvider.EnsureDNSARecord(c.latticeID, domain, ip)
		if err != nil {
			state := latticev1.AddressStateFailed
			failureInfo := &latticev1.AddressStatusFailureInfo{
				Message: fmt.Sprintf("error creating DNS A record: %v", err),
				Time:    metav1.Now(),
			}

			// swallow any errors from updating the status and return the original error
			c.updateAddressStatus(address, state, &failureInfo.Message, failureInfo, address.Status.Ports)
			return fmt.Errorf("error creating service address DNS A record for %v: %v", address.Description(c.namespacePrefix), err)
		}
	}

	serviceMeshPorts, err := c.serviceMesh.ServiceMeshPorts(service.Annotations)
	if err != nil {
		return fmt.Errorf("error getting service mesh ports for %v: %v", service.Description(c.namespacePrefix), err)
	}

	needsUpdate, err = c.cloudProvider.ServiceAddressLoadBalancerNeedsUpdate(c.latticeID, address, service, serviceMeshPorts)
	if err != nil {
		return fmt.Errorf(
			"error checking if load balancer for %v needs update: %v",
			address.Description(c.namespacePrefix),
			err,
		)
	}

	if needsUpdate {
		message := "updating load balancer"
		address, err = c.updateAddressStatus(
			address,
			latticev1.AddressStateUpdating,
			&message,
			nil,
			address.Status.Ports,
		)
		if err != nil {
			return err
		}

		err = c.cloudProvider.EnsureServiceAddressLoadBalancer(c.latticeID, address, service, serviceMeshPorts)
		if err != nil {
			state := latticev1.AddressStateFailed
			failureInfo := &latticev1.AddressStatusFailureInfo{
				Message: fmt.Sprintf("error creating load balancer: %v", err),
				Time:    metav1.Now(),
			}

			// swallow any errors from updating the status and return the original error
			c.updateAddressStatus(address, state, &failureInfo.Message, failureInfo, address.Status.Ports)
			return fmt.Errorf("error creating load balancer for %v: %v", address.Description(c.namespacePrefix), err)
		}

		// Add any annotations needed by the cloud provider.
		// Copy annotations so cloud provider doesn't mutate the cache
		annotations := make(map[string]string)
		for k, v := range address.Annotations {
			annotations[k] = v
		}

		err = c.cloudProvider.ServiceAddressLoadBalancerAddAnnotations(c.latticeID, address, service, serviceMeshPorts, annotations)
		if err != nil {
			return fmt.Errorf("cloud provider could not get annotations for %v: %v", address.Description(c.namespacePrefix), err)
		}

		address, err = c.updateAddressAnnotations(address, annotations)
		if err != nil {
			return err
		}
	}

	ports := address.Status.Ports
	if service.NeedsAddressLoadBalancer() {
		ports, err = c.cloudProvider.ServiceAddressLoadBalancerPorts(c.latticeID, address, service, serviceMeshPorts)
		if err != nil {
			return fmt.Errorf(
				"error getting %v %v load balancer ports: %v",
				address.Description(c.namespacePrefix),
				service.Description(c.namespacePrefix),
				err,
			)
		}
	}

	_, err = c.updateAddressStatus(address, latticev1.AddressStateStable, nil, nil, ports)
	return err
}

func (c *Controller) service(namespace string, path tree.Path) (*latticev1.Service, error) {
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
