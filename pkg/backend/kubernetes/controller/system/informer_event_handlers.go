package system

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	err = c.newServiceMesh()
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
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

	cloudProvider, err := cloudprovider.NewCloudProvider(
		c.namespacePrefix,
		c.kubeClient,
		c.kubeInformerFactory,
		c.latticeInformerFactory,
		options,
	)
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

func (c *Controller) handleSystemAdd(obj interface{}) {
	system := obj.(*latticev1.System)

	if system.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleSystemDelete(system)
		return
	}

	c.handleSystemEvent(system, "added")
}

func (c *Controller) handleSystemUpdate(old, cur interface{}) {
	system := cur.(*latticev1.System)
	c.enqueue(system)
}

func (c *Controller) handleSystemDelete(obj interface{}) {
	system, ok := obj.(*latticev1.System)

	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		system, ok = tombstone.Obj.(*latticev1.System)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a system %#v", obj))
			return
		}
	}

	c.handleSystemEvent(system, "deleted")
}

func (c *Controller) handleSystemEvent(system *latticev1.System, verb string) {
	glog.V(4).Infof("%v %v", system.Description(), verb)
	c.enqueue(system)
}

// handleServiceAdd enqueues the System that manages a Service when the Service is created.
func (c *Controller) handleServiceAdd(obj interface{}) {
	service := obj.(*latticev1.Service)

	if service.DeletionTimestamp != nil {
		c.handleServiceDelete(service)
		return
	}

	c.handleServiceEvent(service, "added")
}

// handleServiceAdd enqueues the System that manages a Service when the Service is update.
func (c *Controller) handleServiceUpdate(old, cur interface{}) {
	service := cur.(*latticev1.Service)
	c.handleServiceEvent(service, "updated")
}

func (c *Controller) handleServiceDelete(obj interface{}) {
	service, ok := obj.(*latticev1.Service)
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

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(service); controllerRef != nil {
		system := c.resolveControllerRef(service.Namespace, controllerRef)

		// not a system. This shouldn't happen.
		if system == nil {
			// FIXME: send error event
			return
		}

		c.enqueue(system)
		return
	}

	// It's an orphan. This shouldn't happen.
	// FIXME: send error event
}

// handleNamespaceAdd enqueues the System that manages a Service when the Service is created.
func (c *Controller) handleNamespaceAdd(obj interface{}) {
	namespace := obj.(*corev1.Namespace)

	if namespace.DeletionTimestamp != nil {
		c.handleNamespaceDelete(obj)
		return
	}

	c.handleNamespaceEvent(namespace, "added")
}

// handleServiceAdd enqueues the System that manages a Service when the Service is update.
func (c *Controller) handleNamespaceUpdate(old, cur interface{}) {
	namespace := cur.(*corev1.Namespace)
	c.handleNamespaceEvent(namespace, "updated")
}

func (c *Controller) handleNamespaceDelete(obj interface{}) {
	namespace, ok := obj.(*corev1.Namespace)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		namespace, ok = tombstone.Obj.(*corev1.Namespace)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a namespace %#v", obj))
			return
		}
	}

	c.handleNamespaceEvent(namespace, "deleted")
}

func (c *Controller) handleNamespaceEvent(namespace *corev1.Namespace, verb string) {
	glog.V(4).Infof("namespace %v %v", namespace.Name, verb)

	system := c.resolveNamespaceSystem(namespace.Name)
	if system != nil {
		c.enqueue(system)
	}
}

func (c *Controller) resolveNamespaceSystem(namespace string) *latticev1.System {
	systemID, err := kubeutil.SystemID(c.namespacePrefix, namespace)
	if err != nil {
		// namespace did not conform to the system namespace convention,
		// so it is not a system namespace and thus we don't care about it
		return nil
	}

	system, err := c.systemLister.Systems(kubeutil.InternalNamespace(c.namespacePrefix)).Get(string(systemID))
	if err != nil {
		// FIXME: probably want to send a warning if this wasn't a does not exist
		return nil
	}

	return system
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (c *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *latticev1.System {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != latticev1.SystemKind.Kind {
		// This shouldn't happen
		// FIXME: send error event
		return nil
	}

	internalNamespace := kubeutil.InternalNamespace(c.namespacePrefix)
	system, err := c.systemLister.Systems(internalNamespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	if system.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to. This shouldn't happen.
		// FIXME: send error event
		return nil
	}
	return system
}
