package systembuild

import (
	"fmt"
	"reflect"
	"time"

	latticeclientset "github.com/mlab-lattice/system/pkg/kubernetes/customresource/client"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("SystemBuild")

type SystemBuildController struct {
	syncHandler        func(bKey string) error
	enqueueSystemBuild func(sysBuild *crv1.SystemBuild)

	latticeClient latticeclientset.Interface

	systemBuildStore       cache.Store
	systemBuildStoreSynced cache.InformerSynced

	serviceBuildStore       cache.Store
	serviceBuildStoreSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewSystemBuildController(
	latticeClient latticeclientset.Interface,
	systemBuildInformer cache.SharedInformer,
	serviceBuildInformer cache.SharedInformer,
) *SystemBuildController {
	sbc := &SystemBuildController{
		latticeClient: latticeClient,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-build"),
	}

	sbc.enqueueSystemBuild = sbc.enqueue
	sbc.syncHandler = sbc.syncSystemBuild

	systemBuildInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addSystemBuild,
		UpdateFunc: sbc.updateSystemBuild,
		// TODO: for now it is assumed that SystemBuilds are not deleted.
		// in the future we'll probably want to add a GC process for SystemBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document SysB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.systemBuildStore = systemBuildInformer.GetStore()
	sbc.systemBuildStoreSynced = systemBuildInformer.HasSynced

	serviceBuildInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addServiceBuild,
		UpdateFunc: sbc.updateServiceBuild,
		// TODO: for now it is assumed that ServiceBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ServiceBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document SvcB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.serviceBuildStore = serviceBuildInformer.GetStore()
	sbc.serviceBuildStoreSynced = serviceBuildInformer.HasSynced

	return sbc
}

func (sbc *SystemBuildController) addSystemBuild(obj interface{}) {
	sysb := obj.(*crv1.SystemBuild)
	glog.V(4).Infof("Adding SystemBuild %s", sysb.Name)
	sbc.enqueueSystemBuild(sysb)
}

func (sbc *SystemBuildController) updateSystemBuild(old, cur interface{}) {
	oldSysb := old.(*crv1.SystemBuild)
	curSysb := cur.(*crv1.SystemBuild)
	glog.V(4).Infof("Updating SystemBuild %s", oldSysb.Name)
	sbc.enqueueSystemBuild(curSysb)
}

// addServiceBuild enqueues the System that manages a Service when the Service is created.
func (sbc *SystemBuildController) addServiceBuild(obj interface{}) {
	svcb := obj.(*crv1.ServiceBuild)

	if svcb.DeletionTimestamp != nil {
		// We assume for now that ServiceBuilds do not get deleted.
		// FIXME: send error event
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(svcb); controllerRef != nil {
		sysb := sbc.resolveControllerRef(svcb.Namespace, controllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if sysb == nil {
			// FIXME: send error event
			return
		}

		glog.V(4).Infof("ServiceBuild %s added.", svcb.Name)
		sbc.enqueueSystemBuild(sysb)
		return
	}

	// It's an orphan. This shouldn't happen.
	// FIXME: send error event
}

// updateServiceBuild figures out what SystemBuild manages a Service when the
// Service is updated and enqueues them.
func (sbc *SystemBuildController) updateServiceBuild(old, cur interface{}) {
	glog.V(5).Info("Got ServiceBuild update")
	oldSvcb := old.(*crv1.ServiceBuild)
	curSvcb := cur.(*crv1.ServiceBuild)
	if curSvcb.ResourceVersion == oldSvcb.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("ServiceBuild ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curSvcb)
	oldControllerRef := metav1.GetControllerOf(oldSvcb)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// This shouldn't happen
		// FIXME: send error event
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		sysb := sbc.resolveControllerRef(curSvcb.Namespace, curControllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if sysb == nil {
			// FIXME: send error event
			return
		}

		sbc.enqueueSystemBuild(sysb)
		return
	}

	// Otherwise, it's an orphan. This should not happen.
	// FIXME: send error event
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (sbc *SystemBuildController) resolveControllerRef(ns string, controllerRef *metav1.OwnerReference) *crv1.SystemBuild {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		// This shouldn't happen
		// FIXME: send error event
		return nil
	}

	sysbKey := ns + "/" + controllerRef.Name
	sysbObj, exists, err := sbc.systemBuildStore.GetByKey(sysbKey)
	if err != nil || !exists {
		// This shouldn't happen.
		// FIXME: send error event
		return nil
	}

	sysb := sysbObj.(*crv1.SystemBuild)

	if sysb.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to. This shouldn't happen.
		// FIXME: send error event
		return nil
	}
	return sysb
}

func (sbc *SystemBuildController) enqueue(sysb *crv1.SystemBuild) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sysb)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", sysb, err))
		return
	}

	sbc.queue.Add(key)
}

func (sbc *SystemBuildController) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer sbc.queue.ShutDown()

	glog.Infof("Starting system-build controller")
	defer glog.Infof("Shutting down system-build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, sbc.systemBuildStoreSynced, sbc.serviceBuildStoreSynced) {
		return
	}

	glog.V(4).Info("Caches synced.")

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(sbc.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (sbc *SystemBuildController) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for sbc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (sbc *SystemBuildController) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := sbc.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer sbc.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := sbc.syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		sbc.queue.Forget(key)
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
	sbc.queue.AddRateLimited(key)

	return true
}

// syncSystemBuild will sync the SystemBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (sbc *SystemBuildController) syncSystemBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing SystemBuild %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing SystemBuild %q (%v)", key, time.Now().Sub(startTime))
	}()

	sysbObj, exists, err := sbc.systemBuildStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("SystemBuild %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	sysb := sysbObj.(*crv1.SystemBuild)

	stateInfo, err := sbc.calculateState(sysb)
	if err != nil {
		return err
	}

	glog.V(5).Infof("SystemBuild %v state: %v", sysb.Name, stateInfo.state)

	sysbCopy := sysb.DeepCopy()

	err = sbc.syncServiceBuildStates(sysbCopy, stateInfo)
	if err != nil {
		return nil
	}

	switch stateInfo.state {
	case sysBuildStateHasFailedCBuilds:
		return sbc.syncFailedSystemBuild(sysbCopy, stateInfo.failedSvcbs)
	case sysBuildStateHasOnlyRunningOrSucceededCBuilds:
		return sbc.syncRunningSystemBuild(sysbCopy, stateInfo.activeSvcbs)
	case sysBuildStateNoFailuresNeedsNewCBuilds:
		return sbc.syncMissingServiceBuildsSystemBuild(sysbCopy, stateInfo.needsNewSvcb)
	case sysBuildStateAllCBuildsSucceeded:
		return sbc.syncSucceededSystemBuild(sysbCopy)
	default:
		panic("unreachable")
	}
}
