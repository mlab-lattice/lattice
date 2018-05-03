package address

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncExternalNameAddress(address *latticev1.Address) error {
	if address.Spec.ExternalName == nil {
		return fmt.Errorf("cannot sync external name address with no external name")
	}

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

	domain := kubeutil.InternalAddressSubdomain(path.ToDomain(), systemID, c.latticeID)
	err = c.cloudProvider.EnsureDNSCNAMERecord(c.latticeID, domain, *address.Spec.ExternalName)
	if err != nil {
		return err
	}

	state := latticev1.AddressStateStable
	var failureInfo *latticev1.AddressStatusFailureInfo
	if err != nil {
		state = latticev1.AddressStateFailed
		failureInfo = &latticev1.AddressStatusFailureInfo{
			Message: fmt.Sprintf("error creating DNS CNAME record: %v", err),
			Time:    metav1.Now(),
		}
	}

	_, updateErr := c.updateAddressStatus(address, state, failureInfo, address.Status.Ports)
	if err != nil {
		return err
	}
	if updateErr != nil {
		return updateErr
	}

	return nil
}
