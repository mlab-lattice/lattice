package system

import (
	"fmt"
	"reflect"
	"time"

	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("System")

type SystemController struct {
	syncHandler   func(sysKey string) error
	enqueueSystem func(sysBuild *crv1.System)

	latticeResourceRestClient rest.Interface

	systemStore       cache.Store
	systemStoreSynced cache.InformerSynced

	serviceStore       cache.Store
	serviceStoreSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewSystemController(
	latticeResourceRestClient rest.Interface,
	systemInformer cache.SharedInformer,
	serviceInformer cache.SharedInformer,
) *SystemController {
	sc := &SystemController{
		latticeResourceRestClient: latticeResourceRestClient,
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system"),
	}

	sc.enqueueSystem = sc.enqueue
	sc.syncHandler = sc.syncSystem

	systemInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.addSystem,
		UpdateFunc: sc.updateSystem,
	})
	sc.systemStore = systemInformer.GetStore()
	sc.systemStoreSynced = systemInformer.HasSynced

	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.addService,
		UpdateFunc: sc.updateService,
		// TODO: for now it is assumed that ServiceBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ServiceBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document SvcB GC ideas (need to write down last used date, lock properly, etc)
	})
	sc.serviceStore = serviceInformer.GetStore()
	sc.serviceStoreSynced = serviceInformer.HasSynced

	return sc
}

func (sc *SystemController) addSystem(obj interface{}) {
	sys := obj.(*crv1.System)
	glog.V(4).Infof("Adding System %s", sys.Name)
	sc.enqueueSystem(sys)
}

func (sc *SystemController) updateSystem(old, cur interface{}) {
	oldSys := old.(*crv1.System)
	curSys := cur.(*crv1.System)
	glog.V(4).Infof("Updating System %s", oldSys.Name)
	sc.enqueueSystem(curSys)
}

// addService enqueues the System that manages a Service when the Service is created.
func (sc *SystemController) addService(obj interface{}) {
	svc := obj.(*crv1.Service)

	if svc.DeletionTimestamp != nil {
		// We assume for now that ServiceBuilds do not get deleted.
		// FIXME: send error event
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(svc); controllerRef != nil {
		sys := sc.resolveControllerRef(svc.Namespace, controllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if sys == nil {
			// FIXME: send error event
			return
		}

		glog.V(4).Infof("Service %s added.", svc.Name)
		sc.enqueueSystem(sys)
		return
	}

	// It's an orphan. This shouldn't happen.
	// FIXME: send error event
}

// addService enqueues the System that manages a Service when the Service is update.
func (sc *SystemController) updateService(old, cur interface{}) {
	glog.V(5).Info("Got Service update")
	oldSvc := old.(*crv1.Service)
	curSvc := cur.(*crv1.Service)
	if curSvc.ResourceVersion == oldSvc.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("Service ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curSvc)
	oldControllerRef := metav1.GetControllerOf(oldSvc)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// This shouldn't happen
		// FIXME: send error event
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		sys := sc.resolveControllerRef(curSvc.Namespace, curControllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if sys == nil {
			// FIXME: send error event
			return
		}

		sc.enqueueSystem(sys)
		return
	}

	// Otherwise, it's an orphan. This should not happen.
	// FIXME: send error event
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (sc *SystemController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.System {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		// This shouldn't happen
		// FIXME: send error event
		return nil
	}

	sysKey := namespace + "/" + controllerRef.Name
	sysObj, exists, err := sc.systemStore.GetByKey(sysKey)
	if err != nil || !exists {
		// This shouldn't happen.
		// FIXME: send error event
		return nil
	}

	sys := sysObj.(*crv1.System)

	if sys.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to. This shouldn't happen.
		// FIXME: send error event
		return nil
	}
	return sys
}

func (sc *SystemController) enqueue(sys *crv1.System) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sys)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", sys, err))
		return
	}

	sc.queue.Add(key)
}

func (sc *SystemController) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer sc.queue.ShutDown()

	glog.Infof("Starting system controller")
	defer glog.Infof("Shutting down system controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, sc.systemStoreSynced, sc.serviceStoreSynced) {
		return
	}

	glog.V(4).Info("Caches synced.")

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(sc.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (sc *SystemController) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for sc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (sc *SystemController) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := sc.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer sc.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := sc.syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		sc.queue.Forget(key)
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
	sc.queue.AddRateLimited(key)

	return true
}

// syncSystem will sync the SystemBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (sc *SystemController) syncSystem(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing System %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing System %q (%v)", key, time.Now().Sub(startTime))
	}()

	sysObj, exists, err := sc.systemStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("System %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	sys := sysObj.(*crv1.System)
	sysCopy := sys.DeepCopy()

	if err := sc.syncSystemServiceStatuses(sysCopy); err != nil {
		return err
	}

	return sc.syncSystemStatus(sysCopy)
}

// Warning: syncSystemServiceStatuses mutates sysBuild. Do not pass in a pointer to a
// System from the shared cache.
func (sc *SystemController) syncSystemServiceStatuses(sys *crv1.System) error {
	for path, svcInfo := range sys.Spec.Services {
		// Check if we've already created a Service. If so just grab its status.
		if svcInfo.ServiceName != nil {
			svcState := sc.getServiceState(sys.Namespace, *svcInfo.ServiceName)
			if svcState == nil {
				// This shouldn't happen.
				// FIXME: send error event
				failedState := crv1.ServiceStateRolloutFailed
				svcState = &failedState
			}
			svcInfo.ServiceState = svcState
			sys.Spec.Services[path] = svcInfo
			continue
		}

		// Otherwise we'll have to create a new Service.
		svc, err := sc.createService(sys, &svcInfo.Definition, path, svcInfo.BuildName)
		if err != nil {
			return err
		}
		svcInfo.ServiceName = &(svc.Name)
		svcInfo.ServiceState = &(svc.Status.State)
		sys.Spec.Services[path] = svcInfo
	}

	response := &crv1.System{}
	err := sc.latticeResourceRestClient.Put().
		Namespace(sys.Namespace).
		Resource(crv1.SystemResourcePlural).
		Name(sys.Name).
		Body(sys).
		Do().
		Into(response)

	sys = response
	return err
}

func (sc *SystemController) createService(
	sys *crv1.System,
	svcDefinitionBlock *systemdefinition.Service,
	svcPath systemtree.NodePath,
	svcBuildName string,
) (*crv1.Service, error) {
	svc := getNewServiceFromDefinition(sys, svcDefinitionBlock, svcPath, svcBuildName)

	result := &crv1.Service{}
	err := sc.latticeResourceRestClient.Post().
		Namespace(sys.Namespace).
		Resource(crv1.ServiceResourcePlural).
		Body(svc).
		Do().
		Into(result)
	return result, err
}

// Warning: syncSystemStatus mutates sysBuild. Do not pass in a pointer to a
// SystemBuild from the shared cache.
// syncSystemStatus assumes that all sysBuild.Spec.Services have all had their
// ServiceBuilds created and ServiceBuildStates populated
func (sc *SystemController) syncSystemStatus(sys *crv1.System) error {
	hasFailedSvcRollout := false
	hasActiveSvcRollout := false

	for path, svc := range sys.Spec.Services {
		if svc.ServiceState == nil {
			return fmt.Errorf("Service %v had no ServiceBuildState in syncSystemStatus", path)
		}

		// If there's a failed rollout, no need to look any further, our System has failed.
		if *svc.ServiceState == crv1.ServiceStateRolloutFailed {
			hasFailedSvcRollout = true
			break
		}

		if *svc.ServiceState != crv1.ServiceStateRolloutSucceeded {
			hasActiveSvcRollout = true
		}
	}

	newStatus := calculateSystemStatus(hasFailedSvcRollout, hasActiveSvcRollout)

	if reflect.DeepEqual(sys.Status, newStatus) {
		return nil
	}

	sys.Status = newStatus

	err := sc.latticeResourceRestClient.Put().
		Namespace(sys.Namespace).
		Resource(crv1.SystemResourcePlural).
		Name(sys.Name).
		Body(sys).
		Do().
		Into(nil)

	return err
}

func calculateSystemStatus(hasFailedSvcRollout, hasActiveSvcRollout bool) crv1.SystemStatus {
	if hasFailedSvcRollout {
		return crv1.SystemStatus{
			State: crv1.SystemStateRolloutFailed,
		}
	}

	if hasActiveSvcRollout {
		return crv1.SystemStatus{
			State: crv1.SystemStateRollingOut,
		}
	}

	return crv1.SystemStatus{
		State: crv1.SystemStateRolloutSucceeded,
	}
}
