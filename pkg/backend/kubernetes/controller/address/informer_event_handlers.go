package address

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
)

func (c *Controller) handleConfigAdd(obj interface{}) {
	config := obj.(*latticev1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = config.DeepCopy().Spec

	err := c.newCloudProvider()
	if err != nil {
		glog.Errorf("error creating cloud provider: %v", err)
		// FIXME: what to do here?
		return
	}

	err = c.newServiceMesh()
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}

	if !c.configSet {
		c.configSet = true
		close(c.configSetChan)
	}
}

func (c *Controller) handleConfigUpdate(old, cur interface{}) {
	oldConfig := old.(*latticev1.Config)
	curConfig := cur.(*latticev1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = curConfig.DeepCopy().Spec

	err := c.newCloudProvider()
	if err != nil {
		glog.Errorf("error creating cloud provider: %v", err)
		// FIXME: what to do here?
		return
	}

	err = c.newServiceMesh()
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}
}

func (c *Controller) newCloudProvider() error {
	options, err := cloudprovider.OverlayConfigOptions(c.staticCloudProviderOptions, &c.config.CloudProvider)
	if err != nil {
		return err
	}

	cloudProvider, err := cloudprovider.NewCloudProvider(c.namespacePrefix, c.kubeClient, c.kubeServiceLister, options)
	if err != nil {
		return err
	}

	c.cloudProvider = cloudProvider
	return nil
}

func (c *Controller) newServiceMesh() error {
	options, err := servicemesh.OverlayConfigOptions(c.staticServiceMeshOptions, &c.config.ServiceMesh)
	if err != nil {
		return err
	}

	serviceMesh, err := servicemesh.NewServiceMesh(options)
	if err != nil {
		return err
	}

	c.serviceMesh = serviceMesh
	return nil
}

func (c *Controller) handleAddressAdd(obj interface{}) {
	address := obj.(*latticev1.Address)

	if address.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleAddressDelete(address)
		return
	}

	c.handleAddressEvent(address, "added")
}

func (c *Controller) handleAddressUpdate(old, cur interface{}) {
	address := cur.(*latticev1.Address)
	c.handleAddressEvent(address, "updated")
}

func (c *Controller) handleAddressDelete(obj interface{}) {
	address, ok := obj.(*latticev1.Address)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		address, ok = tombstone.Obj.(*latticev1.Address)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	c.handleAddressEvent(address, "deleted")
}

func (c *Controller) handleAddressEvent(address *latticev1.Address, verb string) {
	glog.V(5).Info("%v %v", address.Description(c.namespacePrefix), verb)
	c.enqueueAddress(address)
}

func (c *Controller) handleServiceAdd(obj interface{}) {
	service := obj.(*latticev1.Service)

	if service.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleServiceDelete(service)
		return
	}

	c.handleServiceEvent(service, "added")
}

func (c *Controller) handleServiceUpdate(old, cur interface{}) {
	service := cur.(*latticev1.Service)
	c.handleServiceEvent(service, "updated")
}

func (c *Controller) handleServiceDelete(obj interface{}) {
	service, ok := obj.(*latticev1.Service)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		service, ok = tombstone.Obj.(*latticev1.Service)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	c.handleServiceEvent(service, "deleted")
}

func (c *Controller) handleServiceEvent(service *latticev1.Service, verb string) {
	glog.V(5).Info("%v %v", service.Description(c.namespacePrefix), verb)

	path, err := service.PathLabel()
	if err != nil {
		// FIXME: send warn event
		return
	}

	addresses, err := c.addressLister.Addresses(service.Namespace).List(labels.Everything())
	if err != nil {
		// FIXME: send warn event
		return
	}

	for _, address := range addresses {
		if address.Spec.Service != nil && path == *address.Spec.Service {
			c.enqueueAddress(address)
		}
	}
}
