package address

import (
	"fmt"

	"github.com/golang/glog"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncDeletedAddress(address *latticev1.Address) error {
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	path, err := address.PathLabel()
	if err != nil {
		return fmt.Errorf("error getting path label for %v: %v", address.Description(c.namespacePrefix), err)
	}

	systemID, err := kubeutil.SystemID(c.namespacePrefix, address.Namespace)
	if err != nil {
		return fmt.Errorf("error getting system id for %v: %v", address.Description(c.namespacePrefix), err)
	}

	message := "deleting DNS record"
	address, err = c.updateAddressStatus(
		address,
		latticev1.AddressStateDeleting,
		&message,
		nil,
		address.Status.Ports,
	)
	if err != nil {
		return err
	}

	domain := kubeutil.InternalAddressSubdomain(path.ToDomain(), systemID, c.latticeID)
	err = c.cloudProvider.DestroyDNSRecord(c.latticeID, domain)
	if err != nil {
		state := latticev1.AddressStateFailed
		failureInfo := &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error deleting DNS record: %v", err),
			Time:    metav1.Now(),
		}

		// swallow any error from updating the status and return the original error
		c.updateAddressStatus(address, state, &failureInfo.Message, failureInfo, address.Status.Ports)
		return fmt.Errorf("error deleting DNS record: %v", err)
	}

	// TODO <GEB>: add check to verify the IP found in the DNS A record is the same as the one being
	//             released

	annotations, err := c.serviceMesh.ReleaseServiceIP(address)
	if err != nil {
		return fmt.Errorf("error releasing service IP: %v", err)
	} else {
		glog.V(4).Infof("released service IP for: %v", address.Name)
		glog.V(4).Infof("service IP annotations: %v", annotations)
	}

	address, err = c.mergeAndUpdateAddressAnnotations(address, annotations)
	if err != nil {
		return fmt.Errorf("error updating address annotations to remove service IP: %v", err)
	}

	message = "deleting load balancer"
	address, err = c.updateAddressStatus(
		address,
		latticev1.AddressStateDeleting,
		&message,
		nil,
		address.Status.Ports,
	)
	if err != nil {
		return err
	}

	err = c.cloudProvider.DestroyServiceAddressLoadBalancer(c.latticeID, address)
	if err != nil {
		state := latticev1.AddressStateFailed
		failureInfo := &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error deleting load balancer: %v", err),
			Time:    metav1.Now(),
		}

		// swallow any error from updating the status and return the original error
		c.updateAddressStatus(address, state, &failureInfo.Message, failureInfo, address.Status.Ports)
		return fmt.Errorf("error deleting load balancer: %v", err)
	}

	_, err = c.removeFinalizer(address)
	return err
}
