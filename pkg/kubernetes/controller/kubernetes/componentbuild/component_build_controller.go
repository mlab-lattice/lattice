package componentbuild

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/listers/lattice/v1"

	batchv1 "k8s.io/api/batch/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	batchinformers "k8s.io/client-go/informers/batch/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("ComponentBuild")

type Controller struct {
	syncHandler func(bKey string) error
	enqueue     func(cb *crv1.ComponentBuild)

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	configLister       latticelisters.ConfigLister
	configListerSynced cache.InformerSynced
	configSetChan      chan struct{}
	configSet          bool
	configLock         sync.RWMutex
	config             *crv1.ConfigSpec

	componentBuildLister       latticelisters.ComponentBuildLister
	componentBuildListerSynced cache.InformerSynced

	jobLister       batchlisters.JobLister
	jobListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	configInformer latticeinformers.ConfigInformer,
	componentBuildInformer latticeinformers.ComponentBuildInformer,
	jobInformer batchinformers.JobInformer,
) *Controller {
	cbc := &Controller{
		kubeClient:    kubeClient,
		latticeClient: latticeClient,
		configSetChan: make(chan struct{}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "component-build"),
	}

	cbc.syncHandler = cbc.syncComponentBuild
	cbc.enqueue = cbc.enqueueComponentBuild

	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    cbc.addConfig,
		UpdateFunc: cbc.updateConfig,
		// TODO: for now it is assumed that ComponentBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ComponentBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document CB GC ideas (need to write down last used date, lock properly, etc)
	})
	cbc.configLister = configInformer.Lister()
	cbc.configListerSynced = configInformer.Informer().HasSynced

	componentBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cbc.addComponentBuild,
		UpdateFunc: cbc.updateComponentBuild,
		// TODO: for now we'll assume that Config is never deleted
	})
	cbc.componentBuildLister = componentBuildInformer.Lister()
	cbc.componentBuildListerSynced = componentBuildInformer.Informer().HasSynced

	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cbc.addJob,
		UpdateFunc: cbc.updateJob,
		// We should probably never delete BuildComponent jobs, but just in case
		// we need to pull the plug on one we'll look out for it.
		DeleteFunc: cbc.deleteJob,
	})
	cbc.jobLister = jobInformer.Lister()
	cbc.jobListerSynced = jobInformer.Informer().HasSynced

	return cbc
}

func (cbc *Controller) addConfig(obj interface{}) {
	config := obj.(*crv1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	cbc.configLock.Lock()
	defer cbc.configLock.Unlock()
	cbc.config = &config.DeepCopy().Spec

	if !cbc.configSet {
		cbc.configSet = true
		close(cbc.configSetChan)
	}
}

func (cbc *Controller) updateConfig(old, cur interface{}) {
	oldConfig := old.(*crv1.Config)
	curConfig := cur.(*crv1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	cbc.configLock.Lock()
	defer cbc.configLock.Unlock()
	cbc.config = &curConfig.DeepCopy().Spec
}

func (cbc *Controller) addComponentBuild(obj interface{}) {
	cb := obj.(*crv1.ComponentBuild)
	glog.V(4).Infof("Adding ComponentBuild %s", cb.Name)
	cbc.enqueueComponentBuild(cb)
}

func (cbc *Controller) updateComponentBuild(old, cur interface{}) {
	oldCb := old.(*crv1.ComponentBuild)
	curCb := cur.(*crv1.ComponentBuild)
	glog.V(4).Infof("Updating ComponentBuild %s", oldCb.Name)
	cbc.enqueueComponentBuild(curCb)
}

// addJob enqueues the ComponentBuild that manages a Job when the Job is created.
func (cbc *Controller) addJob(obj interface{}) {
	j := obj.(*batchv1.Job)

	if j.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		cbc.deleteJob(j)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(j); controllerRef != nil {
		cb := cbc.resolveControllerRef(j.Namespace, controllerRef)

		// Not a ComponentBuild Job
		if cb == nil {
			return
		}

		glog.V(4).Infof("Job %s added.", j.Name)
		cbc.enqueueComponentBuild(cb)
		return
	}

	// Otherwise, it's an orphan. We don't care about these since ComponentBuild Jobs should
	// always have a ControllerRef, so therefore this Job is not a ComponentBuild Job.
}

// updateJob figures out what ComponentBuild manages a Job when the Job
// is updated and wake them up.
// Note that a ComponentBuild Job should only ever and should always be owned by a single ComponentBuild
// (the one referenced in its ControllerRef), so an updated job should
// have the same controller ref for both the old and current versions.
func (cbc *Controller) updateJob(old, cur interface{}) {
	glog.V(5).Info("Got Job putComponentBuildUpdate")
	oldJ := old.(*batchv1.Job)
	curJ := cur.(*batchv1.Job)
	if curJ.ResourceVersion == oldJ.ResourceVersion {
		// Periodic resync will send putComponentBuildUpdate events for all known jobs.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("Job ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curJ)
	oldControllerRef := metav1.GetControllerOf(oldJ)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// The ControllerRef was changed. If this is a ComponentBuild Job, this shouldn't happen.
		if b := cbc.resolveControllerRef(oldJ.Namespace, oldControllerRef); b != nil {
			// FIXME: send error event here, this should not happen
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		cb := cbc.resolveControllerRef(curJ.Namespace, curControllerRef)

		// Not a ComponentBuild Job
		if cb == nil {
			return
		}

		cbc.enqueueComponentBuild(cb)
		return
	}

	// Otherwise, it's an orphan. We don't care about these since ComponentBuild Jobs should
	// always have a ControllerRef, so therefore this Job is not a ComponentBuild Job.
}

// deleteJob enqueues the ComponentBuild that manages a Job when
// the Job is deleted. obj could be an *extensions.ComponentBuild, or
// a DeletionFinalStateUnknown marker item.
// Note that this should never really happen. ComponentBuild Jobs should not
// be getting deleted. But if they do, we should write a warn event
// and putComponentBuildUpdate the ComponentBuild.
func (cbc *Controller) deleteJob(obj interface{}) {
	j, ok := obj.(*batchv1.Job)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		j, ok = tombstone.Obj.(*batchv1.Job)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Job %#v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(j)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	cb := cbc.resolveControllerRef(j.Namespace, controllerRef)

	// Not a ComponentBuild job
	if cb == nil {
		return
	}

	glog.V(4).Infof("Job %s deleted.", j.Name)
	cbc.enqueueComponentBuild(cb)
}

func (cbc *Controller) enqueueComponentBuild(cBuild *crv1.ComponentBuild) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cBuild)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", cBuild, err))
		return
	}

	cbc.queue.Add(key)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (cbc *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.ComponentBuild {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}

	cb, err := cbc.componentBuildLister.ComponentBuilds(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	if cb.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return cb
}

func (cbc *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer cbc.queue.ShutDown()

	glog.Infof("Starting component-build controller")
	defer glog.Infof("Shutting down component-build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, cbc.configListerSynced, cbc.componentBuildListerSynced, cbc.jobListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Waiting for config to be set")

	// wait for config to be set
	<-cbc.configSetChan

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(cbc.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (cbc *Controller) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for cbc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (cbc *Controller) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := cbc.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer cbc.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := cbc.syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		cbc.queue.Forget(key)
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
	cbc.queue.AddRateLimited(key)

	return true
}

// syncComponentBuild will sync the ComponentBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (cbc *Controller) syncComponentBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing ComponentBuild %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing ComponentBuild %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	cb, err := cbc.componentBuildLister.ComponentBuilds(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("ComponentBuild %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	stateInfo, err := cbc.calculateState(cb)
	if err != nil {
		return err
	}

	glog.V(5).Infof("ComponentBuild %v state: %v", cb.Name, stateInfo.state)

	// Deep-copy otherwise we are mutating our cache.
	cBuildCopy := cb.DeepCopy()

	switch stateInfo.state {
	case cBuildStateJobNotCreated:
		return cbc.syncJoblessComponentBuild(cBuildCopy)
	case cBuildStateJobSucceeded:
		return cbc.syncSuccessfulComponentBuild(cBuildCopy, stateInfo.job)
	case cBuildStateJobFailed:
		return cbc.syncFailedComponentBuild(cBuildCopy)
	case cBuildStateJobRunning:
		return cbc.syncUnfinishedComponentBuild(cBuildCopy, stateInfo.job)
	default:
		panic("unreachable")
	}
}
