package system

import (
	"fmt"
	"reflect"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/listers/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("System")

type Controller struct {
	syncHandler   func(sysKey string) error
	enqueueSystem func(sysBuild *crv1.System)

	latticeClient latticeclientset.Interface

	systemLister       latticelisters.SystemLister
	systemListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	latticeClient latticeclientset.Interface,
	systemInformer latticeinformers.SystemInformer,
	serviceInformer latticeinformers.ServiceInformer,
) *Controller {
	sc := &Controller{
		latticeClient: latticeClient,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system"),
	}

	sc.enqueueSystem = sc.enqueue
	sc.syncHandler = sc.syncSystem

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

	return sc
}

func (sc *Controller) handleSystemAdd(obj interface{}) {
	sys := obj.(*crv1.System)
	glog.V(4).Infof("Adding System %s", sys.Name)
	sc.enqueueSystem(sys)
}

func (sc *Controller) handleSystemUpdate(old, cur interface{}) {
	oldSys := old.(*crv1.System)
	curSys := cur.(*crv1.System)
	glog.V(4).Infof("Updating System %s", oldSys.Name)
	sc.enqueueSystem(curSys)
}

// handleServiceAdd enqueues the System that manages a Service when the Service is created.
func (sc *Controller) handleServiceAdd(obj interface{}) {
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

// handleServiceAdd enqueues the System that manages a Service when the Service is update.
func (sc *Controller) handleServiceUpdate(old, cur interface{}) {
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

func (sc *Controller) handleServiceDelete(obj interface{}) {
	svc, ok := obj.(*crv1.Service)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		svc, ok = tombstone.Obj.(*crv1.Service)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}
	glog.V(4).Infof("Service %s deleted", svc.Name)
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

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (sc *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.System {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		// This shouldn't happen
		// FIXME: send error event
		return nil
	}

	sys, err := sc.systemLister.Systems(namespace).Get(controllerRef.Name)
	if err != nil {
		// This shouldn't happen.
		// FIXME: send error event
		return nil
	}

	if sys.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to. This shouldn't happen.
		// FIXME: send error event
		return nil
	}
	return sys
}

func (sc *Controller) enqueue(sys *crv1.System) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sys)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", sys, err))
		return
	}

	sc.queue.Add(key)
}

func (sc *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer sc.queue.ShutDown()

	glog.Infof("Starting system controller")
	defer glog.Infof("Shutting down system controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, sc.systemListerSynced, sc.serviceListerSynced) {
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

func (sc *Controller) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for sc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (sc *Controller) processNextWorkItem() bool {
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
func (sc *Controller) syncSystem(key string) error {
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
	sys, err := sc.systemLister.Systems(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("System %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	sysCopy := sys.DeepCopy()

	if sysCopy.DeletionTimestamp != nil {
		return sc.syncDeletedSystem(sysCopy)
	}

	err = sc.syncSystemServices(sysCopy)
	if err != nil {
		return err
	}

	if err := sc.syncSystemServiceStatuses(sysCopy); err != nil {
		return err
	}

	return sc.syncSystemStatus(sysCopy)
}

// Warning: syncDeletedSystem mutates sys. Do not pass in a pointer to a
// System from the shared cache.
func (sc *Controller) syncDeletedSystem(sys *crv1.System) error {
	deletedSvc := false
	// Delete all Services in our namespace
	svcs, err := sc.serviceLister.Services(sys.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	for _, svc := range svcs {
		glog.V(4).Infof("Found Service %q in Namespace %q, deleting", svc.Name, svc.Namespace)
		deletedSvc = true
		err := sc.deleteService(svc)
		if err != nil {
			return err
		}
	}

	if !deletedSvc {
		return sc.removeFinalizer(sys)
	}

	return nil
}

// Warning: syncSystemServices mutates sys. Do not pass in a pointer to a
// System from the shared cache.
func (sc *Controller) syncSystemServices(sys *crv1.System) error {
	validSvcNames := map[string]bool{}

	// Loop through the Services defined in the System's Spec, and create/update any that need it
	for path, svcInfo := range sys.Spec.Services {
		// If the Service doesn't exist already, create one.
		if svcInfo.ServiceName == nil {
			glog.V(5).Infof("Did not find a Service for %q, creating one", path)
			svc, err := sc.createService(sys, &svcInfo, path)
			if err != nil {
				return err
			}
			svcInfo.ServiceName = &(svc.Name)
			svcInfo.ServiceState = &(svc.Status.State)
			sys.Spec.Services[path] = svcInfo

			validSvcNames[svc.Name] = true
			continue
		}

		// A Service has already been created. Check if its definition is the same
		// definition. We'll assume that the rest of the spec is properly formed.
		svc, err := sc.getService(sys.Namespace, *svcInfo.ServiceName)
		if err != nil {
			return err
		}

		if svc == nil {
			// FIXME: send warn event
			// TODO: should we just create a new Service here?
			return fmt.Errorf(
				"service %v has ServiceName %v but Service does not exist",
				path,
				svcInfo.ServiceName,
			)
		}

		validSvcNames[svc.Name] = true

		// If the definitions are the same, assume we're good.
		if reflect.DeepEqual(svcInfo.Definition, svc.Spec.Definition) {
			continue
		}

		// Otherwise, get a new spec and update the service.
		newSpec, err := getNewServiceSpec(&svcInfo, path)
		if err != nil {
			return nil
		}
		svc.Spec = newSpec

		// Need to update that we're rolling out again.
		svc.Status.State = crv1.ServiceStateRollingOut

		_, err = sc.updateService(svc)
		if err != nil {
			return nil
		}
	}

	// Loop through all of the Services that exist in the System's namespace, and delete any
	// that are no longer a part of the System's Spec
	// TODO: should we wait until all other services are successfully rolled out before deleting these?
	// need to figure out what the rollout/automatic roll-back strategy is
	svcs, err := sc.serviceLister.Services(sys.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	for _, svc := range svcs {
		// Only care about Services in this System's Namespace
		if svc.Namespace != sys.Namespace {
			continue
		}

		if _, ok := validSvcNames[svc.Name]; !ok {
			glog.V(4).Infof("Found Service %q in Namespace %q that is no longer in the System Spec", svc.Name, svc.Namespace)
			err := sc.deleteService(svc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Warning: syncSystemServiceStatuses mutates sys. Do not pass in a pointer to a
// System from the shared cache.
func (sc *Controller) syncSystemServiceStatuses(sys *crv1.System) error {
	for path, svcInfo := range sys.Spec.Services {
		// Services should have been created already by syncSystemServices.
		if svcInfo.ServiceName == nil {
			// FIXME: send warn event
			return fmt.Errorf("expected Service %v to have ServiceName", path)
		}

		svcState, err := sc.getServiceState(sys.Namespace, *svcInfo.ServiceName)
		if err != nil {
			return err
		}

		if svcState == nil {
			// This shouldn't happen.
			// FIXME: send error event
			return fmt.Errorf("Service %v exists but does not have a State", path)
		}

		svcInfo.ServiceState = svcState
		sys.Spec.Services[path] = svcInfo
	}

	result, err := sc.updateSystem(sys)
	if err != nil {
		return err
	}

	*sys = *result
	return nil
}

// Warning: syncSystemStatus mutates sysBuild. Do not pass in a pointer to a
// SystemBuild from the shared cache.
// syncSystemStatus assumes that all sysBuild.Spec.Services have all had their
// ServiceBuilds created and ServiceBuildStates populated
func (sc *Controller) syncSystemStatus(sys *crv1.System) error {
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

	sys.Status.State = calculateSystemState(hasFailedSvcRollout, hasActiveSvcRollout)
	result, err := sc.updateSystem(sys)
	if err != nil {
		return err
	}

	*sys = *result
	return nil
}

func (sc *Controller) updateSystem(sys *crv1.System) (*crv1.System, error) {
	return sc.latticeClient.LatticeV1().Systems(sys.Namespace).Update(sys)
}

func calculateSystemState(hasFailedSvcRollout, hasActiveSvcRollout bool) crv1.SystemState {
	if hasFailedSvcRollout {
		return crv1.SystemStateRolloutFailed
	}

	if hasActiveSvcRollout {
		return crv1.SystemStateRollingOut
	}

	return crv1.SystemStateRolloutSucceeded
}
