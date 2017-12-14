package systembuild

import (
	"fmt"
	"reflect"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("SystemBuild")

type Controller struct {
	syncHandler        func(bKey string) error
	enqueueSystemBuild func(sysBuild *crv1.SystemBuild)

	latticeClient latticeclientset.Interface

	systemBuildLister       latticelisters.SystemBuildLister
	systemBuildListerSynced cache.InformerSynced

	serviceBuildLister       latticelisters.ServiceBuildLister
	serviceBuildListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	latticeClient latticeclientset.Interface,
	systemBuildInformer latticeinformers.SystemBuildInformer,
	serviceBuildInformer latticeinformers.ServiceBuildInformer,

) *Controller {
	sbc := &Controller{
		latticeClient: latticeClient,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-build"),
	}

	sbc.enqueueSystemBuild = sbc.enqueue
	sbc.syncHandler = sbc.syncSystemBuild

	systemBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addSystemBuild,
		UpdateFunc: sbc.updateSystemBuild,
		// TODO: for now it is assumed that SystemBuilds are not deleted.
		// in the future we'll probably want to add a GC process for SystemBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document SysB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.systemBuildLister = systemBuildInformer.Lister()
	sbc.systemBuildListerSynced = systemBuildInformer.Informer().HasSynced

	serviceBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addServiceBuild,
		UpdateFunc: sbc.updateServiceBuild,
		// TODO: for now it is assumed that ServiceBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ServiceBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document SvcB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.serviceBuildLister = serviceBuildInformer.Lister()
	sbc.serviceBuildListerSynced = serviceBuildInformer.Informer().HasSynced

	return sbc
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting system-build controller")
	defer glog.Infof("Shutting down system-build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.systemBuildListerSynced, c.serviceBuildListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced.")

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

func (c *Controller) addSystemBuild(obj interface{}) {
	build := obj.(*crv1.SystemBuild)
	glog.V(4).Infof("Adding SystemBuild %s", build.Name)
	c.enqueueSystemBuild(build)
}

func (c *Controller) updateSystemBuild(old, cur interface{}) {
	oldBuild := old.(*crv1.SystemBuild)
	curBuild := cur.(*crv1.SystemBuild)
	glog.V(4).Infof("Updating SystemBuild %s", oldBuild.Name)
	c.enqueueSystemBuild(curBuild)
}

// addServiceBuild enqueues the System that manages a Service when the Service is created.
func (c *Controller) addServiceBuild(obj interface{}) {
	serviceBuild := obj.(*crv1.ServiceBuild)

	if serviceBuild.DeletionTimestamp != nil {
		// We assume for now that ServiceBuilds do not get deleted.
		// FIXME: send error event
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(serviceBuild); controllerRef != nil {
		build := c.resolveControllerRef(serviceBuild.Namespace, controllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if build == nil {
			// FIXME: send error event
			return
		}

		glog.V(4).Infof("ServiceBuild %s added.", serviceBuild.Name)
		c.enqueueSystemBuild(build)
		return
	}

	// It's an orphan. This shouldn't happen.
	// FIXME: send error event
}

// updateServiceBuild figures out what SystemBuild manages a Service when the
// Service is updated and enqueues them.
func (c *Controller) updateServiceBuild(old, cur interface{}) {
	glog.V(5).Info("Got ServiceBuild update")
	oldServiceBuild := old.(*crv1.ServiceBuild)
	curServiceBuild := cur.(*crv1.ServiceBuild)
	if curServiceBuild.ResourceVersion == oldServiceBuild.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("ServiceBuild ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curServiceBuild)
	oldControllerRef := metav1.GetControllerOf(oldServiceBuild)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// This shouldn't happen
		// FIXME: send error event
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		build := c.resolveControllerRef(curServiceBuild.Namespace, curControllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if build == nil {
			// FIXME: send error event
			return
		}

		c.enqueueSystemBuild(build)
		return
	}

	// Otherwise, it's an orphan. This should not happen.
	// FIXME: send error event
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (c *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.SystemBuild {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		// This shouldn't happen
		// FIXME: send error event
		return nil
	}

	build, err := c.systemBuildLister.SystemBuilds(namespace).Get(controllerRef.Name)
	if err != nil {
		// This shouldn't happen.
		// FIXME: send error event
		return nil
	}

	if build.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to. This shouldn't happen.
		// FIXME: send error event
		return nil
	}
	return build
}

func (c *Controller) enqueue(sysb *crv1.SystemBuild) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sysb)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", sysb, err))
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

// syncSystemBuild will sync the SystemBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncSystemBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing SystemBuild %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing SystemBuild %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	build, err := c.systemBuildLister.SystemBuilds(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("SystemBuild %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	stateInfo, err := c.calculateState(build)
	if err != nil {
		return err
	}

	glog.V(5).Infof("SystemBuild %v state: %v", key, stateInfo.state)

	switch stateInfo.state {
	case stateHasFailedServiceBuilds:
		return c.syncFailedSystemBuild(build, stateInfo)
	case stateHasOnlyRunningOrSucceededServiceBuilds:
		return c.syncRunningSystemBuild(build, stateInfo)
	case stateNoFailuresNeedsNewServiceBuilds:
		return c.syncMissingServiceBuildsSystemBuild(build, stateInfo)
	case stateAllServiceBuildsSucceeded:
		return c.syncSucceededSystemBuild(build, stateInfo)
	default:
		panic("unreachable")
	}
}
