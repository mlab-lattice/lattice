package address

import (
	"fmt"

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

	domain := kubeutil.InternalAddressSubdomain(path.ToDomain(), systemID, c.latticeID)
	err = c.cloudProvider.DestroyDNSRecord(c.latticeID, domain)
	if err != nil {
		state := latticev1.AddressStateFailed
		failureInfo := &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error deleting DNS record: %v", err),
			Time:    metav1.Now(),
		}

		// swallow any error from updating the status and return the original error
		c.updateAddressStatus(address, state, failureInfo, address.Status.Ports)
		return fmt.Errorf("error deleting DNS record: %v", err)
	}

	_, err = c.removeFinalizer(address)
	return err
}
