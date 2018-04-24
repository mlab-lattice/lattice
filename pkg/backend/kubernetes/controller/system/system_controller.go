package system

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	coreinformers "k8s.io/client-go/informers/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
)

type Controller struct {
	syncHandler   func(sysKey string) error
	enqueueSystem func(sysBuild *latticev1.System)

	namespacePrefix string
	latticeID       v1.LatticeID

	// NOTE: you must get a read lock on the configLock for the duration
	//       of your use of the cloudProvider
	staticCloudProviderOptions *cloudprovider.Options
	cloudProvider              cloudprovider.Interface

	// NOTE: you must get a read lock on the configLock for the duration
	//       of your use of the serviceMesh
	staticServiceMeshOptions *servicemesh.Options
	serviceMesh              servicemesh.Interface

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	configLister       latticelisters.ConfigLister
	configListerSynced cache.InformerSynced
	configSetChan      chan struct{}
	configSet          bool
	configLock         sync.RWMutex
	config             latticev1.ConfigSpec

	systemLister       latticelisters.SystemLister
	systemListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	namespaceLister       corelisters.NamespaceLister
	namespaceListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	namespacePrefix string,
	latticeID v1.LatticeID,
	cloudProviderOptions *cloudprovider.Options,
	serviceMeshOptions *servicemesh.Options,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	configInformer latticeinformers.ConfigInformer,
	systemInformer latticeinformers.SystemInformer,
	serviceInformer latticeinformers.ServiceInformer,
	namespaceInformer coreinformers.NamespaceInformer,
) *Controller {
	sc := &Controller{
		namespacePrefix: namespacePrefix,
		latticeID:       latticeID,

		staticCloudProviderOptions: cloudProviderOptions,
		staticServiceMeshOptions:   serviceMeshOptions,

		kubeClient:    kubeClient,
		latticeClient: latticeClient,
		configSetChan: make(chan struct{}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system"),
	}

	sc.enqueueSystem = sc.enqueue
	sc.syncHandler = sc.syncSystem

	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    sc.handleConfigAdd,
		UpdateFunc: sc.handleConfigUpdate,
		// TODO(kevinrosendahl): for now it is assumed that ComponentBuilds are not deleted.
	})
	sc.configLister = configInformer.Lister()
	sc.configListerSynced = configInformer.Informer().HasSynced

	systemInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleSystemAdd,
		UpdateFunc: sc.handleSystemUpdate,
	})
	sc.systemLister = systemInformer.Lister()
	sc.systemListerSynced = systemInformer.Informer().HasSynced

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleServiceAdd,
		UpdateFunc: sc.handleServiceUpdate,
		DeleteFunc: sc.handleServiceDelete,
	})
	sc.serviceLister = serviceInformer.Lister()
	sc.serviceListerSynced = serviceInformer.Informer().HasSynced

	namespaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleNamespaceAdd,
		UpdateFunc: sc.handleNamespaceUpdate,
		DeleteFunc: sc.handleNamespaceDelete,
	})
	sc.namespaceLister = namespaceInformer.Lister()
	sc.namespaceListerSynced = namespaceInformer.Informer().HasSynced

	return sc
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting system controller")
	defer glog.Infof("Shutting down system controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.configListerSynced, c.systemListerSynced, c.serviceListerSynced, c.namespaceListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Waiting for config to be set")

	// wait for config to be set
	<-c.configSetChan

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

func (c *Controller) handleSystemAdd(obj interface{}) {
	system := obj.(*latticev1.System)
	glog.V(4).Infof("Adding System %s", system.Name)
	c.enqueueSystem(system)
}

func (c *Controller) handleSystemUpdate(old, cur interface{}) {
	oldSystem := old.(*latticev1.System)
	curSystem := cur.(*latticev1.System)
	glog.V(4).Infof("Updating System %s", oldSystem.Name)
	c.enqueueSystem(curSystem)
}

// handleServiceAdd enqueues the System that manages a Service when the Service is created.
func (c *Controller) handleServiceAdd(obj interface{}) {
	service := obj.(*latticev1.Service)

	if service.DeletionTimestamp != nil {
		// We assume for now that ServiceBuilds do not get deleted.
		// FIXME: send error event
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(service); controllerRef != nil {
		system := c.resolveControllerRef(service.Namespace, controllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if system == nil {
			// FIXME: send error event
			return
		}

		glog.V(4).Infof("Service %s added.", service.Name)
		c.enqueueSystem(system)
		return
	}

	// It's an orphan. This shouldn't happen.
	// FIXME: send error event
}

// handleServiceAdd enqueues the System that manages a Service when the Service is update.
func (c *Controller) handleServiceUpdate(old, cur interface{}) {
	glog.V(5).Info("Got Service update")
	oldService := old.(*latticev1.Service)
	curService := cur.(*latticev1.Service)
	if curService.ResourceVersion == oldService.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("Service ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curService)
	oldControllerRef := metav1.GetControllerOf(oldService)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// This shouldn't happen
		// FIXME: send error event
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		system := c.resolveControllerRef(curService.Namespace, curControllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if system == nil {
			// FIXME: send error event
			return
		}

		c.enqueueSystem(system)
		return
	}

	// Otherwise, it's an orphan. This should not happen.
	// FIXME: send error event
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
	glog.V(4).Infof("Service %s deleted", service.Name)
	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(service); controllerRef != nil {
		system := c.resolveControllerRef(service.Namespace, controllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if system == nil {
			// FIXME: send error event
			return
		}

		glog.V(4).Infof("Service %s added.", service.Name)
		c.enqueueSystem(system)
		return
	}

	// It's an orphan. This shouldn't happen.
	// FIXME: send error event
}

// handleNamespaceAdd enqueues the System that manages a Service when the Service is created.
func (c *Controller) handleNamespaceAdd(obj interface{}) {
	ns := obj.(*corev1.Namespace)

	if ns.DeletionTimestamp != nil {
		c.handleNamespaceDelete(obj)
		return
	}

	system := c.resolveNamespaceSystem(ns.Name)
	if system != nil {
		c.enqueueSystem(system)
	}
}

// handleServiceAdd enqueues the System that manages a Service when the Service is update.
func (c *Controller) handleNamespaceUpdate(old, cur interface{}) {
	glog.V(5).Info("Got Namespace update")
	c.handleNamespaceAdd(cur)
}

func (c *Controller) handleNamespaceDelete(obj interface{}) {
	ns := obj.(*corev1.Namespace)
	system := c.resolveNamespaceSystem(ns.Name)
	if system != nil {
		c.enqueueSystem(system)
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

func (c *Controller) enqueue(sys *latticev1.System) {
	if sys.Namespace != kubeutil.InternalNamespace(c.namespacePrefix) {
		glog.V(4).Infof("System %v/%v is not a part of this lattice, ignoring", sys.Namespace, sys.Name)
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sys)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", sys, err))
		return
	}

	c.queue.Add(key)
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

// syncSystem will sync the SystemBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncSystem(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing System %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing System %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	system, err := c.systemLister.Systems(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("System %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	if system.DeletionTimestamp != nil {
		return c.syncDeletingSystem(system)
	}

	if system.Status.State == latticev1.SystemStatePending {
		return c.syncPendingSystem(system)
	}

	return c.syncLiveSystem(system)
}
