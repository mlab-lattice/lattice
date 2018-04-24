package address

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncDeletedAddress(address *latticev1.Address) error {
	if address.Spec.ExternalName != nil {
		return c.syncDeletedServiceAddress(address)
	}

	if address.Spec.Service != nil {
		return c.syncDeletedServiceAddress(address)
	}

	return nil
}

func (c *Controller) syncDeletedExternalNameAddress(address *latticev1.Address) error {
	if address.Spec.ExternalName == nil {
		return fmt.Errorf("cannot sync external name address with no external name")
	}

	c.configLock.RLock()
	defer c.configLock.RUnlock()

	path, err := address.PathLabel()
	if err != nil {
		return err
	}

	systemID, err := kubernetes.SystemID(c.namespacePrefix, address.Namespace)
	if err != nil {
		return err
	}

	domain := fmt.Sprintf("%v.local.%v", path.ToDomain(), systemID)
	err = c.cloudProvider.DestroyDNSCNAMERecord(c.latticeID, domain)
	if err != nil {
		state := latticev1.AddressStateFailed
		failureInfo := &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error deleting DNS A record: %v", err),
			Time:    metav1.Now(),
		}

		c.updateAddressStatus(address, state, failureInfo, address.Status.Ports)
		return err
	}

	return nil
}

func (c *Controller) syncDeletedServiceAddress(address *latticev1.Address) error {
	if address.Spec.Service == nil {
		return fmt.Errorf("cannot delete service address with no service path")
	}

	c.configLock.RLock()
	defer c.configLock.RUnlock()

	path, err := address.PathLabel()
	if err != nil {
		return err
	}

	systemID, err := kubernetes.SystemID(c.namespacePrefix, address.Namespace)
	if err != nil {
		return err
	}

	// FIXME: get proper dns name
	domain := fmt.Sprintf("%v.local.%v", path.ToDomain(), systemID)
	err = c.cloudProvider.DestroyDNSARecord(c.latticeID, domain)
	if err != nil {
		state := latticev1.AddressStateFailed
		failureInfo := &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error deleting DNS A record: %v", err),
			Time:    metav1.Now(),
		}

		c.updateAddressStatus(address, state, failureInfo, address.Status.Ports)
		return err
	}

	err = c.cloudProvider.DestroyServiceAddressLoadBalancer(c.latticeID, address)
	if err != nil {
		state := latticev1.AddressStateFailed
		failureInfo := &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error deleting load balancer: %v", err),
			Time:    metav1.Now(),
		}

		c.updateAddressStatus(address, state, failureInfo, address.Status.Ports)
		return err
	}

	return nil
}
