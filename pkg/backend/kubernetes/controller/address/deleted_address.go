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
		return err
	}

	systemID, err := kubeutil.SystemID(c.namespacePrefix, address.Namespace)
	if err != nil {
		return err
	}

	domain := kubeutil.InternalSubdomain(path.ToDomain(), systemID, c.latticeID)
	err = c.cloudProvider.DestroyDNSRecord(c.latticeID, domain)
	if err != nil {
		state := latticev1.AddressStateFailed
		failureInfo := &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error deleting DNS record: %v", err),
			Time:    metav1.Now(),
		}

		c.updateAddressStatus(address, state, failureInfo, address.Status.Ports)
		return err
	}

	_, err = c.removeFinalizer(address)
	return err
}
