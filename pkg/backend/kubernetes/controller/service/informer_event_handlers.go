package service

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		glog.Errorf("error creating service mesh: %v", err)
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
		glog.Errorf("error creating service mesh: %v", err)
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

	cloudProvider, err := cloudprovider.NewCloudProvider(c.namespacePrefix, nil, nil, options)
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

func (c *Controller) handleServiceAdd(obj interface{}) {
	service := obj.(*latticev1.Service)
	glog.V(4).Infof("%v added", service.Description(c.namespacePrefix))
	c.enqueueService(service)
}

func (c *Controller) handleServiceUpdate(old, cur interface{}) {
	service := cur.(*latticev1.Service)
	glog.V(4).Infof("%v updated", service.Description(c.namespacePrefix))
	c.enqueueService(service)
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
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}
	glog.V(4).Infof("%v delete", service.Description(c.namespacePrefix))
	c.enqueueService(service)
}

func (c *Controller) handleNodePoolAdd(obj interface{}) {
	nodePool := obj.(*latticev1.NodePool)
	c.handleNodePoolEvent(nodePool, "added")
}

func (c *Controller) handleNodePoolUpdate(old, cur interface{}) {
	nodePool := cur.(*latticev1.NodePool)
	c.handleNodePoolEvent(nodePool, "added")
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
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Deployment %#v", obj))
			return
		}
	}

	c.handleNodePoolEvent(nodePool, "deleted")
}

func (c *Controller) handleNodePoolEvent(nodePool *latticev1.NodePool, verb string) {
	glog.V(4).Infof("%v %v", nodePool.Description(c.namespacePrefix), verb)

	services, err := c.nodePoolServices(nodePool)
	if err != nil {
		// FIXME: log/send warn event
		return
	}

	for _, service := range services {
		c.enqueueService(&service)
	}
}

// handleDeploymentAdd enqueues the Service that manages a Deployment when the Deployment is created.
func (c *Controller) handleDeploymentAdd(obj interface{}) {
	deployment := obj.(*appsv1.Deployment)

	if deployment.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleDeploymentDelete(deployment)
		return
	}

	c.handleDeploymentEvent(deployment, "added")
}

// handleDeploymentUpdate figures out what Service manages a Deployment when the Deployment
// is updated and enqueues it.
func (c *Controller) handleDeploymentUpdate(old, cur interface{}) {
	deployment := cur.(*appsv1.Deployment)
	c.handleDeploymentEvent(deployment, "updated")
}

// handleDeploymentDelete enqueues the Service that manages a Deployment when
// the Deployment is deleted.
func (c *Controller) handleDeploymentDelete(obj interface{}) {
	deployment, ok := obj.(*appsv1.Deployment)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		deployment, ok = tombstone.Obj.(*appsv1.Deployment)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Deployment %#v", obj))
			return
		}
	}

	c.handleDeploymentEvent(deployment, "deleted")
}

func (c *Controller) handleDeploymentEvent(deployment *appsv1.Deployment, verb string) {
	glog.V(4).Infof("deployment %v/%v %v", deployment.Namespace, deployment.Name, verb)

	// see if the deployment has a service as a controller owning reference
	if controllerRef := metav1.GetControllerOf(deployment); controllerRef != nil {
		service := c.resolveControllerRef(deployment.Namespace, controllerRef)

		// Not a Service Deployment.
		if service == nil {
			return
		}

		c.enqueueService(service)
		return
	}

	// Otherwise, it's an orphan. These shouldn't exist within a lattice controlled namespace.
	// TODO: maybe log/send warn event if there's an orphan deployment in a lattice controlled namespace
}

// handlePodDelete enqueues the Service that manages a Pod when
// the Pod is deleted.
func (c *Controller) handlePodDelete(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		pod, ok = tombstone.Obj.(*corev1.Pod)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Deployment %#v", obj))
			return
		}
	}

	serviceID, ok := pod.Labels[latticev1.ServiceIDLabelKey]
	if !ok {
		// not a service pod
		return
	}

	service, err := c.serviceLister.Services(pod.Namespace).Get(serviceID)
	if err != nil {
		if errors.IsNotFound(err) {
			// service doesn't exist anymore, so it doesn't care about this
			// pod's deletion
			return
		}

		runtime.HandleError(fmt.Errorf("error retrieving service %v/%v", pod.Namespace, serviceID))
		return
	}

	glog.V(4).Infof("pod %s/%s deleted", pod.Namespace, pod.Name)
	c.enqueueService(service)
}

func (c *Controller) handleKubeServiceAdd(obj interface{}) {
	kubeService := obj.(*corev1.Service)

	if kubeService.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleKubeServiceDelete(kubeService)
		return
	}

	c.handleKubeServiceEvent(kubeService, "added")
}

func (c *Controller) handleKubeServiceUpdate(old, cur interface{}) {
	kubeService := cur.(*corev1.Service)
	c.handleKubeServiceEvent(kubeService, "updated")
}

func (c *Controller) handleKubeServiceDelete(obj interface{}) {
	kubeService, ok := obj.(*corev1.Service)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		kubeService, ok = tombstone.Obj.(*corev1.Service)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	c.handleKubeServiceEvent(kubeService, "deleted")
}

func (c *Controller) handleKubeServiceEvent(kubeService *corev1.Service, verb string) {
	glog.V(4).Infof("kube service %v/%v %v", kubeService.Namespace, kubeService.Name, verb)

	// see if the kube service has a service as a controller owning reference
	if controllerRef := metav1.GetControllerOf(kubeService); controllerRef != nil {
		service := c.resolveControllerRef(kubeService.Namespace, controllerRef)

		// Not a service kube service
		if service == nil {
			return
		}

		c.enqueueService(service)
		return
	}

	// Otherwise, it's an orphan. These shouldn't exist within a lattice controlled namespace.
	// TODO: maybe log/send warn event if there's an orphan deployment in a lattice controlled namespace
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
	glog.V(4).Infof("%v %v", address.Description(c.namespacePrefix), verb)

	// see if the address has a service as a controller owning reference
	if controllerRef := metav1.GetControllerOf(address); controllerRef != nil {
		service := c.resolveControllerRef(address.Namespace, controllerRef)

		// Not a service address
		if service == nil {
			return
		}

		c.enqueueService(service)
		return
	}

	// Otherwise, it's an orphan. These shouldn't exist within a lattice controlled namespace.
	// TODO: maybe log/send warn event if there's an orphan address in a lattice controlled namespace
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (c *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *latticev1.Service {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != latticev1.ServiceKind.Kind {
		return nil
	}

	service, err := c.serviceLister.Services(namespace).Get(controllerRef.Name)
	if err != nil {
		// FIXME(kevinrosendahl): send error?
		return nil
	}

	if service.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return service
}
