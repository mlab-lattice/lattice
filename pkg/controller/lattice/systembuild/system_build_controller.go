package systembuild

import (
	"fmt"
	"reflect"
	"time"

	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"

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

var controllerKind = crv1.SchemeGroupVersion.WithKind("SystemBuild")

type SystemBuildController struct {
	syncHandler        func(bKey string) error
	enqueueSystemBuild func(sysBuild *crv1.SystemBuild)

	latticeResourceRestClient rest.Interface

	systemBuildStore       cache.Store
	systemBuildStoreSynced cache.InformerSynced

	serviceBuildStore       cache.Store
	serviceBuildStoreSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewSystemBuildController(
	latticeResourceRestClient rest.Interface,
	systemBuildInformer cache.SharedInformer,
	serviceBuildInformer cache.SharedInformer,
) *SystemBuildController {
	sbc := &SystemBuildController{
		latticeResourceRestClient: latticeResourceRestClient,
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-build"),
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
	sysBuild := obj.(*crv1.SystemBuild)
	glog.V(4).Infof("Adding SystemBuild %s", sysBuild.Name)
	sbc.enqueueSystemBuild(sysBuild)
}

func (sbc *SystemBuildController) updateSystemBuild(old, cur interface{}) {
	oldSysBuild := old.(*crv1.SystemBuild)
	curSysBuild := cur.(*crv1.SystemBuild)
	glog.V(4).Infof("Updating SystemBuild %s", oldSysBuild.Name)
	sbc.enqueueSystemBuild(curSysBuild)
}

// addServiceBuild enqueues the System that manages a Service when the Service is created.
func (sbc *SystemBuildController) addServiceBuild(obj interface{}) {
	svcBuild := obj.(*crv1.ServiceBuild)

	if svcBuild.DeletionTimestamp != nil {
		// We assume for now that ServiceBuilds do not get deleted.
		// FIXME: send error event
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(svcBuild); controllerRef != nil {
		sysBuild := sbc.resolveControllerRef(svcBuild.Namespace, controllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if sysBuild == nil {
			// FIXME: send error event
			return
		}

		glog.V(4).Infof("ServiceBuild %s added.", svcBuild.Name)
		sbc.enqueueSystemBuild(sysBuild)
		return
	}

	// It's an orphan. This shouldn't happen.
	// FIXME: send error event
}

// updateServiceBuild figures out what SystemBuild manages a Service when the
// Service is updated and enqueues them.
func (sbc *SystemBuildController) updateServiceBuild(old, cur interface{}) {
	glog.V(5).Info("Got ServiceBuild update")
	oldSvcBuild := old.(*crv1.ServiceBuild)
	curSvcBuild := cur.(*crv1.ServiceBuild)
	if curSvcBuild.ResourceVersion == oldSvcBuild.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("ServiceBuild ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curSvcBuild)
	oldControllerRef := metav1.GetControllerOf(oldSvcBuild)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// This shouldn't happen
		// FIXME: send error event
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		sysBuild := sbc.resolveControllerRef(curSvcBuild.Namespace, curControllerRef)

		// Not a SystemBuild. This shouldn't happen.
		if sysBuild == nil {
			// FIXME: send error event
			return
		}

		sbc.enqueueSystemBuild(sysBuild)
		return
	}

	// Otherwise, it's an orphan. This should not happen.
	// FIXME: send error event
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (sbc *SystemBuildController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.SystemBuild {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		// This shouldn't happen
		// FIXME: send error event
		return nil
	}

	sysBuildKey := namespace + "/" + controllerRef.Name
	sysBuildObj, exists, err := sbc.systemBuildStore.GetByKey(sysBuildKey)
	if err != nil || !exists {
		// This shouldn't happen.
		// FIXME: send error event
		return nil
	}

	sysBuild := sysBuildObj.(*crv1.SystemBuild)

	if sysBuild.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to. This shouldn't happen.
		// FIXME: send error event
		return nil
	}
	return sysBuild
}

func (sbc *SystemBuildController) enqueue(sysBuild *crv1.SystemBuild) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sysBuild)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", sysBuild, err))
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

	sysBuildObj, exists, err := sbc.systemBuildStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("SystemBuild %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	sysBuild := sysBuildObj.(*crv1.SystemBuild)
	sysBuildCopy := sysBuild.DeepCopy()

	if err := sbc.syncSystemBuildServiceStatuses(sysBuildCopy); err != nil {
		return err
	}

	return sbc.syncSystemBuildStatus(sysBuildCopy)
}

// Warning: syncSystemBuildServiceStatuses mutates sysBuild. Do not pass in a pointer to a
// SystemBuild from the shared cache.
func (sbc *SystemBuildController) syncSystemBuildServiceStatuses(sysBuild *crv1.SystemBuild) error {
	for path, svc := range sysBuild.Spec.Services {
		// Check if we've already created a Service. If so just grab its status.
		if svc.ServiceBuildName != nil {
			svcBuildState := sbc.getServiceBuildState(sysBuild.Namespace, *svc.ServiceBuildName)
			if svcBuildState == nil {
				// This shouldn't happen.
				// FIXME: send error event
				failedState := crv1.ServiceBuildStateFailed
				svcBuildState = &failedState
				//sysBuild.Spec.Services[idx].ServiceBuildState = &failedState
			}

			svc.ServiceBuildState = svcBuildState
			sysBuild.Spec.Services[path] = svc
			continue
		}

		// Otherwise we'll have to create a new Service.
		svcBuild, err := sbc.createServiceBuild(sysBuild, &svc.Definition)
		if err != nil {
			return err
		}

		svc.ServiceBuildName = &(svcBuild.Name)
		svc.ServiceBuildState = &(svcBuild.Status.State)
		sysBuild.Spec.Services[path] = svc
	}

	response := &crv1.SystemBuild{}
	err := sbc.latticeResourceRestClient.Put().
		Namespace(sysBuild.Namespace).
		Resource(crv1.SystemBuildResourcePlural).
		Name(sysBuild.Name).
		Body(sysBuild).
		Do().
		Into(response)

	*sysBuild = *response
	return err
}

func (sbc *SystemBuildController) createServiceBuild(
	sysBuild *crv1.SystemBuild,
	svcDefinitionBlock *systemdefinition.Service,
) (*crv1.ServiceBuild, error) {
	svcBuild := getNewServiceBuildFromDefinition(sysBuild, svcDefinitionBlock)

	result := &crv1.ServiceBuild{}
	err := sbc.latticeResourceRestClient.Post().
		Namespace(sysBuild.Namespace).
		Resource(crv1.ServiceBuildResourcePlural).
		Body(svcBuild).
		Do().
		Into(result)
	return result, err
}

// Warning: syncSystemBuildStatus mutates sysBuild. Do not pass in a pointer to a
// SystemBuild from the shared cache.
// syncSystemBuildStatus assumes that all sysBuild.Spec.Services have all had their
// ServiceBuilds created and ServiceBuildStates populated
func (sbc *SystemBuildController) syncSystemBuildStatus(sysBuild *crv1.SystemBuild) error {
	hasFailedSvcBuild := false
	hasActiveSvcBuild := false

	for path, svc := range sysBuild.Spec.Services {
		if svc.ServiceBuildState == nil {
			return fmt.Errorf("ServiceBuild %v had no ServiceBuildState in syncSystemBuildStatus", path)
		}

		// If there's a failed build, no need to look any further, our SystemBuild has failed.
		if *svc.ServiceBuildState == crv1.ServiceBuildStateFailed {
			hasFailedSvcBuild = true
			break
		}

		if *svc.ServiceBuildState != crv1.ServiceBuildStateSucceeded {
			hasActiveSvcBuild = true
		}
	}

	newStatus := calculateSystemBuildStatus(hasFailedSvcBuild, hasActiveSvcBuild)

	if reflect.DeepEqual(sysBuild.Status, newStatus) {
		return nil
	}

	sysBuild.Status = newStatus

	err := sbc.latticeResourceRestClient.Put().
		Namespace(sysBuild.Namespace).
		Resource(crv1.SystemBuildResourcePlural).
		Name(sysBuild.Name).
		Body(sysBuild).
		Do().
		Into(nil)

	return err
}

func calculateSystemBuildStatus(hasFailedSvcBuild, hasActiveSvcBuild bool) crv1.SystemBuildStatus {
	if hasFailedSvcBuild {
		return crv1.SystemBuildStatus{
			State: crv1.SystemBuildStateFailed,
		}
	}

	if hasActiveSvcBuild {
		return crv1.SystemBuildStatus{
			State: crv1.SystemBuildStateRunning,
		}
	}

	return crv1.SystemBuildStatus{
		State: crv1.SystemBuildStateSucceeded,
	}
}
