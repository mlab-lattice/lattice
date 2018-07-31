package address

import (
	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/labels"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) initLeaseManager() error {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	needsLease := make([]*latticev1.Address, 0)
	addresses, err := c.addressLister.List(labels.Everything())
	if err != nil {
		return err
	}
	for _, address := range addresses {
		// NOTE: shouldn't be a race between HasServiceIP here and ServiceIP for addresses still
		//       that still have not been assigned an IP since this method is called before the
		//       controller starts processing events and we are synchronized with other users of
		//       the serviceMesh via the configLock
		ip, err := c.serviceMesh.HasServiceIP(address)
		if err != nil {
			return err
		}
		if ip == "" {
			// NOTE: we need to populate the lease manager with all currently active leases
			//       before assigning new leases
			glog.V(4).Infof("Found service address (%s) with no assigned IP", address.Name)
			needsLease = append(needsLease, address)
			continue
		}
		service, err := c.service(address.Namespace, *address.Spec.Service)
		if err != nil {
			return err
		}
		// ServiceIP will respect the IP currently assigned to an Address as specified in the
		// annotation
		ip, annotations, err := c.serviceMesh.ServiceIP(service, address)
		if err != nil {
			return err
		}
		_, err = c.mergeAndUpdateAddressAnnotations(address, annotations)
		if err != nil {
			return err
		}
		glog.V(4).Infof("Added existing service (%s) IP %s to lease manager", service.Name, ip)
	}
	for _, address := range needsLease {
		service, err := c.service(address.Namespace, *address.Spec.Service)
		if err != nil {
			return err
		}
		ip, annotations, err := c.serviceMesh.ServiceIP(service, address)
		if err != nil {
			return err
		}
		_, err = c.mergeAndUpdateAddressAnnotations(address, annotations)
		if err != nil {
			return err
		}
		glog.V(4).Infof("Added new service (%s) IP %s to lease manager", service.Name, ip)
	}

	return nil
}
