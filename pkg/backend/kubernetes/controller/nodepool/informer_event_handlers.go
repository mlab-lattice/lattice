package nodepool

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
)

func (c *Controller) handleConfigAdd(obj interface{}) {
	config := obj.(*latticev1.Config)
	err := c.handleConfigEvent(config, "added")
	if err != nil {
		return
	}

	c.configLock.Lock()
	defer c.configLock.Unlock()
	if !c.configSet {
		c.configSet = true
		close(c.configSetChan)
	}
}

func (c *Controller) handleConfigUpdate(old, cur interface{}) {
	config := cur.(*latticev1.Config)
	c.handleConfigEvent(config, "updated")
}

func (c *Controller) handleConfigEvent(config *latticev1.Config, verb string) error {
	glog.V(4).Infof("config %v/%v %v", config.Namespace, config.Name, verb)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = config.DeepCopy().Spec

	err := c.newCloudProvider()
	if err != nil {
		glog.Errorf("error creating cloud provider: %v", err)
		// FIXME: what to do here?
		return err
	}

	return nil
}

func (c *Controller) newCloudProvider() error {
	options, err := cloudprovider.OverlayConfigOptions(c.staticCloudProviderOptions, &c.config.CloudProvider)
	if err != nil {
		return err
	}

	cloudProvider, err := cloudprovider.NewCloudProvider(c.namespacePrefix, nil, nil, options)
	if err != nil {
		return err
	}

	c.cloudProvider = cloudProvider
	return nil
}

func (c *Controller) handleNodePoolAdd(obj interface{}) {
	nodePool := obj.(*latticev1.NodePool)

	if nodePool.DeletionTimestamp != nil {
		// on a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion
		c.handleNodePoolDelete(nodePool)
		return
	}

	c.handleNodePoolEvent(nodePool, "added")
}

func (c *Controller) handleNodePoolUpdate(old, cur interface{}) {
	nodePool := cur.(*latticev1.NodePool)
	c.handleNodePoolEvent(nodePool, "updated")
}

func (c *Controller) handleNodePoolDelete(obj interface{}) {
	nodePool, ok := obj.(*latticev1.NodePool)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		nodePool, ok = tombstone.Obj.(*latticev1.NodePool)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a node pool %#v", obj))
			return
		}
	}

	c.handleNodePoolEvent(nodePool, "deleted")
}

func (c *Controller) handleNodePoolEvent(nodePool *latticev1.NodePool, verb string) {
	glog.V(4).Infof("%v %v", nodePool.Description(c.namespacePrefix), verb)
	c.enqueue(nodePool)
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
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a service %#v", obj))
			return
		}
	}

	c.handleServiceEvent(service, "deleted")
}

func (c *Controller) handleServiceEvent(service *latticev1.Service, verb string) {
	glog.V(4).Infof("%v %v", service.Description(c.namespacePrefix), verb)

	// FIXME: for now, just enqueue every node pool when services are updated, in the future
	//        should think about which node pools actually need to be synced
	nodePools, err := c.nodePoolLister.NodePools(service.Namespace).List(labels.Everything())
	if err != nil {
		// FIXME: send warning
		return
	}

	for _, nodePool := range nodePools {
		c.enqueue(nodePool)
	}
}
