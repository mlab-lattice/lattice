package serviceaddress

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("ServiceAddress")

type Controller struct {
	syncHandler           func(bKey string) error
	enqueueServiceAddress func(cb *crv1.ServiceAddress)

	serviceMesh servicemesh.Interface

	latticeClient latticeclientset.Interface

	configLister       latticelisters.ConfigLister
	configListerSynced cache.InformerSynced
	configSetChan      chan struct{}
	configSet          bool
	configLock         sync.RWMutex
	config             crv1.ConfigSpec

	serviceAddressLister       latticelisters.ServiceAddressLister
	serviceAddressListerSynced cache.InformerSynced

	endpointLister       latticelisters.EndpointLister
	endpointListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	latticeClient latticeclientset.Interface,
	configInformer latticeinformers.ConfigInformer,
	serviceAddressInformer latticeinformers.ServiceAddressInformer,
	endpointInformer latticeinformers.EndpointInformer,
) *Controller {
	sc := &Controller{
		latticeClient: latticeClient,
		configSetChan: make(chan struct{}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service"),
	}

	sc.syncHandler = sc.syncService
	sc.enqueueServiceAddress = sc.enqueue

	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    sc.handleConfigAdd,
		UpdateFunc: sc.handleConfigUpdate,
	})
	sc.configLister = configInformer.Lister()
	sc.configListerSynced = configInformer.Informer().HasSynced

	serviceAddressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleServiceAddressAdd,
		UpdateFunc: sc.handleServiceAddressUpdate,
		DeleteFunc: sc.handleServiceAddressDelete,
	})
	sc.serviceAddressLister = serviceAddressInformer.Lister()
	sc.serviceAddressListerSynced = serviceAddressInformer.Informer().HasSynced

	endpointInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleEndpointAdd,
		UpdateFunc: sc.handleEndpointUpdate,
		DeleteFunc: sc.handleEndpointDelete,
	})
	sc.endpointLister = endpointInformer.Lister()
	sc.endpointListerSynced = endpointInformer.Informer().HasSynced

	return sc
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting service controller")
	defer glog.Infof("Shutting down service controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.configListerSynced, c.serviceAddressListerSynced, c.endpointListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Waiting for config to be set")

	// wait for config to be set
	<-c.configSetChan

	glog.V(4).Info("Config set")

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (c *Controller) handleConfigAdd(obj interface{}) {
	config := obj.(*crv1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = config.DeepCopy().Spec

	serviceMesh, err := servicemesh.NewServiceMesh(&c.config.ServiceMesh)
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}

	c.serviceMesh = serviceMesh

	if !c.configSet {
		c.configSet = true
		close(c.configSetChan)
	}
}

func (c *Controller) handleConfigUpdate(old, cur interface{}) {
	oldConfig := old.(*crv1.Config)
	curConfig := cur.(*crv1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = curConfig.DeepCopy().Spec

	serviceMesh, err := servicemesh.NewServiceMesh(&c.config.ServiceMesh)
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}

	c.serviceMesh = serviceMesh
}

func (c *Controller) handleServiceAddressAdd(obj interface{}) {
	address := obj.(*crv1.ServiceAddress)
	glog.V(4).Infof("ServiceAddress %v/%v added", address.Namespace, address.Name)

	if address.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleServiceAddressDelete(address)
		return
	}

	c.enqueueServiceAddress(address)
}

func (c *Controller) handleServiceAddressUpdate(old, cur interface{}) {
	oldAddress := old.(*crv1.ServiceAddress)
	curAddress := cur.(*crv1.ServiceAddress)
	glog.V(5).Info("Got ServiceAddress %v/%v update", curAddress.Namespace, curAddress.Name)
	if curAddress.ResourceVersion == oldAddress.ResourceVersion {
		// Periodic resync will send update events for all known Services.
		// Two different versions of the same Service will always have different RVs.
		glog.V(5).Info("kube Service %v/%v ResourceVersions are the same", curAddress.Namespace, curAddress.Name)
		return
	}

	c.enqueueServiceAddress(curAddress)
}

func (c *Controller) handleServiceAddressDelete(obj interface{}) {
	address, ok := obj.(*crv1.ServiceAddress)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		address, ok = tombstone.Obj.(*crv1.ServiceAddress)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	c.enqueueServiceAddress(address)
}

func (c *Controller) handleEndpointAdd(obj interface{}) {
	endpoint := obj.(*crv1.Endpoint)

	if endpoint.DeletionTimestamp != nil {
		c.handleEndpointDelete(endpoint)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(endpoint); controllerRef != nil {
		address := c.resolveControllerRef(endpoint.Namespace, controllerRef)

		// Not a ServiceAddress. This shouldn't happen.
		if address == nil {
			// FIXME: send error event
			return
		}

		glog.V(4).Infof("Service %s added.", endpoint.Name)
		c.enqueueServiceAddress(address)
		return
	}

	// It's an orphan. This shouldn't happen.
	// FIXME: send error event
}

func (c *Controller) handleEndpointUpdate(old, cur interface{}) {
	glog.V(5).Info("Got Deployment update")
	oldEndpoint := old.(*crv1.Endpoint)
	curEndpoint := cur.(*crv1.Endpoint)
	if curEndpoint.ResourceVersion == oldEndpoint.ResourceVersion {
		// Periodic resync will send update events for all known Deployments.
		// Two different versions of the same Deployment will always have different RVs.
		glog.V(5).Info("Deployment ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curEndpoint)
	oldControllerRef := metav1.GetControllerOf(oldEndpoint)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// The ControllerRef was changed. If this is a Service Deployment, this shouldn't happen.
		if address := c.resolveControllerRef(oldEndpoint.Namespace, oldControllerRef); address != nil {
			// FIXME(kevinrosendahl): send error event here, this should not happen
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		address := c.resolveControllerRef(curEndpoint.Namespace, curControllerRef)

		// Not a Service Deployment
		if address == nil {
			return
		}

		c.enqueueServiceAddress(address)
		return
	}

	// Otherwise, it's an orphan. These should never exist. All deployments should be run by some
	// controller.
	// FIXME(kevinrosendahl): send warn event
}

func (c *Controller) handleEndpointDelete(obj interface{}) {
	endpoint, ok := obj.(*crv1.Endpoint)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		endpoint, ok = tombstone.Obj.(*crv1.Endpoint)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Deployment %#v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(endpoint)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	address := c.resolveControllerRef(endpoint.Namespace, controllerRef)

	// Not a Service Deployment
	if address == nil {
		return
	}

	glog.V(4).Infof("Endpoint %v/%v deleted.", endpoint.Namespace, endpoint.Name)
	c.enqueueServiceAddress(address)
}

func (c *Controller) enqueue(svc *crv1.ServiceAddress) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(svc)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", svc, err))
		return
	}

	c.queue.Add(key)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (c *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.ServiceAddress {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}

	address, err := c.serviceAddressLister.ServiceAddresses(namespace).Get(controllerRef.Name)
	if err != nil {
		// FIXME(kevinrosendahl): send error?
		return nil
	}

	if address.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return address
}

func (c *Controller) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (c *Controller) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer c.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := c.syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		c.queue.Forget(key)
		return true
	}

	// there was a failure so be sure to report it.  This method allows for
	// pluggable error handling which can be used for things like
	// cluster-monitoring
	runtime.HandleError(fmt.Errorf("%v failed with : %v", key, err))

	// since we failed, we should requeue the item to work on later.  This
	// method will add a backoff to avoid hotlooping on particular items
	// (they're probably still not going to work right away) and overall
	// controller protection (everything I've done is broken, this controller
	// needs to calm down or it can starve other useful work) cases.
	c.queue.AddRateLimited(key)

	return true
}

// syncService will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncService(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing Service %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing Service %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	address, err := c.serviceAddressLister.ServiceAddresses(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("Service %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	endpoint, err := c.syncEndpoint(address)
	if err != nil {
		return err
	}

	_, err = c.syncServiceAddressStatus(address, endpoint)
	return err
}
