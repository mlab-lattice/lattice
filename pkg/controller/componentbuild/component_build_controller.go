package componentbuild

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	batchv1 "k8s.io/api/batch/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	batchinformers "k8s.io/client-go/informers/batch/v1"
	clientset "k8s.io/client-go/kubernetes"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("ComponentBuild")

type ComponentBuildController struct {
	provider string

	syncHandler func(bKey string) error
	enqueue     func(cb *crv1.ComponentBuild)

	latticeResourceRestClient rest.Interface
	kubeClient                clientset.Interface

	configStore       cache.Store
	configStoreSynced cache.InformerSynced
	configSetChan     chan struct{}
	configSet         bool
	configLock        sync.RWMutex
	config            crv1.ComponentBuildConfig

	componentBuildStore       cache.Store
	componentBuildStoreSynced cache.InformerSynced

	jobLister       batchlisters.JobLister
	jobListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewComponentBuildController(
	provider string,
	kubeClient clientset.Interface,
	latticeResourceRestClient rest.Interface,
	configInformer cache.SharedInformer,
	componentBuildInformer cache.SharedInformer,
	jobInformer batchinformers.JobInformer,
) *ComponentBuildController {
	cbc := &ComponentBuildController{
		provider:                  provider,
		latticeResourceRestClient: latticeResourceRestClient,
		kubeClient:                kubeClient,
		queue:                     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "component-build"),
	}

	cbc.syncHandler = cbc.syncComponentBuild
	cbc.enqueue = cbc.enqueueComponentBuild

	configInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    cbc.addConfig,
		UpdateFunc: cbc.updateConfig,
		// TODO: for now it is assumed that ComponentBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ComponentBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document CB GC ideas (need to write down last used date, lock properly, etc)
	})
	cbc.configStore = configInformer.GetStore()
	cbc.configStoreSynced = configInformer.HasSynced

	cbc.configSetChan = make(chan struct{})

	componentBuildInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cbc.addComponentBuild,
		UpdateFunc: cbc.updateComponentBuild,
		// TODO: for now we'll assume that Config is never deleted
	})
	cbc.componentBuildStore = componentBuildInformer.GetStore()
	cbc.componentBuildStoreSynced = componentBuildInformer.HasSynced

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

func (cbc *ComponentBuildController) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer cbc.queue.ShutDown()

	glog.Infof("Starting component-build controller")
	defer glog.Infof("Shutting down component-build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, cbc.configStoreSynced, cbc.componentBuildStoreSynced, cbc.jobListerSynced) {
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

func (cbc *ComponentBuildController) addConfig(obj interface{}) {
	config := obj.(*crv1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	cbc.configLock.Lock()
	defer cbc.configLock.Unlock()
	cbc.config = config.DeepCopy().Spec.ComponentBuild

	if !cbc.configSet {
		cbc.configSet = true
		close(cbc.configSetChan)
	}
}

func (cbc *ComponentBuildController) updateConfig(old, cur interface{}) {
	oldConfig := old.(*crv1.Config)
	curConfig := cur.(*crv1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	cbc.configLock.Lock()
	defer cbc.configLock.Unlock()
	cbc.config = curConfig.DeepCopy().Spec.ComponentBuild
}

func (cbc *ComponentBuildController) addComponentBuild(obj interface{}) {
	cBuild := obj.(*crv1.ComponentBuild)
	glog.V(4).Infof("Adding ComponentBuild %s", cBuild.Name)
	cbc.enqueueComponentBuild(cBuild)
}

func (cbc *ComponentBuildController) updateComponentBuild(old, cur interface{}) {
	oldCBuild := old.(*crv1.ComponentBuild)
	curCBuild := cur.(*crv1.ComponentBuild)
	glog.V(4).Infof("Updating ComponentBuild %s", oldCBuild.Name)
	cbc.enqueueComponentBuild(curCBuild)
}

// addJob enqueues the ComponentBuild that manages a Job when the Job is created.
func (cbc *ComponentBuildController) addJob(obj interface{}) {
	job := obj.(*batchv1.Job)

	if job.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		cbc.deleteJob(job)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(job); controllerRef != nil {
		b := cbc.resolveControllerRef(job.Namespace, controllerRef)

		// Not a ComponentBuild Job
		if b == nil {
			return
		}

		glog.V(4).Infof("Job %s added.", job.Name)
		cbc.enqueueComponentBuild(b)
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
func (cbc *ComponentBuildController) updateJob(old, cur interface{}) {
	glog.V(5).Info("Got Job update")
	oldJob := old.(*batchv1.Job)
	curJob := cur.(*batchv1.Job)
	if curJob.ResourceVersion == oldJob.ResourceVersion {
		// Periodic resync will send update events for all known jobs.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("Job ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curJob)
	oldControllerRef := metav1.GetControllerOf(oldJob)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// The ControllerRef was changed. If this is a ComponentBuild Job, this shouldn't happen.
		if b := cbc.resolveControllerRef(oldJob.Namespace, oldControllerRef); b != nil {
			// FIXME: send error event here, this should not happen
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		b := cbc.resolveControllerRef(curJob.Namespace, curControllerRef)

		// Not a ComponentBuild Job
		if b == nil {
			return
		}

		cbc.enqueueComponentBuild(b)
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
// and update the ComponentBuild.
func (cbc *ComponentBuildController) deleteJob(obj interface{}) {
	job, ok := obj.(*batchv1.Job)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		job, ok = tombstone.Obj.(*batchv1.Job)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a Job %#v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(job)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	b := cbc.resolveControllerRef(job.Namespace, controllerRef)

	// Not a ComponentBuild job
	if b == nil {
		return
	}

	glog.V(4).Infof("Job %s deleted.", job.Name)
	cbc.enqueueComponentBuild(b)
}

func (cbc *ComponentBuildController) enqueueComponentBuild(cBuild *crv1.ComponentBuild) {
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
func (cbc *ComponentBuildController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.ComponentBuild {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}

	cBuildKey := fmt.Sprintf("%v/%v", namespace, controllerRef.Name)
	bi, exists, err := cbc.componentBuildStore.GetByKey(cBuildKey)
	if err != nil || !exists {
		return nil
	}

	b := bi.(*crv1.ComponentBuild)

	if b.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return b
}

// getJobForBuild uses ControllerRefManager to retrieve the Job for a ComponentBuild
func (cbc *ComponentBuildController) getJobForBuild(cBuild *crv1.ComponentBuild) (*batchv1.Job, error) {
	// List all Jobs to find in the ComponentBuild's namespace to find the Job the ComponentBuild manages.
	jobList, err := cbc.jobLister.Jobs(cBuild.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	matchingJobs := []*batchv1.Job{}
	cBuildControllerRef := metav1.NewControllerRef(cBuild, controllerKind)

	for _, job := range jobList {
		jobControllerRef := metav1.GetControllerOf(job)

		if reflect.DeepEqual(cBuildControllerRef, jobControllerRef) {
			matchingJobs = append(matchingJobs, job)
		}
	}

	if len(matchingJobs) == 0 {
		return nil, nil
	}

	if len(matchingJobs) > 1 {
		return nil, fmt.Errorf("ComponentBuild %v has multiple Jobs", cBuild.Name)
	}

	return matchingJobs[0], nil
}

func (cbc *ComponentBuildController) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for cbc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (cbc *ComponentBuildController) processNextWorkItem() bool {
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
func (cbc *ComponentBuildController) syncComponentBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing ComponentBuild %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing ComponentBuild %q (%v)", key, time.Now().Sub(startTime))
	}()

	bi, exists, err := cbc.componentBuildStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("ComponentBuild %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	cBuild := bi.(*crv1.ComponentBuild)

	job, err := cbc.getJobForBuild(cBuild)
	if err != nil {
		return err
	}

	// Deep-copy otherwise we are mutating our cache.
	cBuildCopy := cBuild.DeepCopy()

	if job == nil {
		return cbc.createComponentBuildJob(cBuildCopy)
	}

	return cbc.syncComponentBuildStatus(cBuildCopy, job)
}

// Warning: createComponentBuildJob mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) createComponentBuildJob(build *crv1.ComponentBuild) error {
	job := cbc.getBuildJob(build)
	jobResp, err := cbc.kubeClient.BatchV1().Jobs(build.Namespace).Create(job)
	if err != nil {
		// FIXME: send warn event
		return err
	}

	glog.V(4).Infof("Created Job %s", jobResp.Name)
	// FIXME: send normal event
	return cbc.syncComponentBuildStatus(build, jobResp)
}

// Warning; syncComponentBuildStatus mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) syncComponentBuildStatus(build *crv1.ComponentBuild, job *batchv1.Job) error {
	// FIXME: add docker image fqn to build spec
	newStatus := calculateComponentBuildStatus(job)

	if reflect.DeepEqual(build.Status, newStatus) {
		return nil
	}

	build.Status = newStatus

	err := cbc.latticeResourceRestClient.Put().
		Namespace(build.Namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(build.Name).
		Body(build).
		Do().
		Into(nil)

	return err
}

func calculateComponentBuildStatus(job *batchv1.Job) crv1.ComponentBuildStatus {
	finished, succeeded := jobStatus(job)
	if finished {
		if succeeded {
			return crv1.ComponentBuildStatus{
				State: crv1.ComponentBuildStateSucceeded,
			}
		}

		return crv1.ComponentBuildStatus{
			State: crv1.ComponentBuildStateFailed,
		}
	}

	// The Job Pods have been able to be scheduled, so the ComponentBuild is "running" even though
	// a Job Pod may not currently be active.
	if job.Status.Active > 0 || job.Status.Failed > 0 || job.Status.Succeeded > 0 {
		return crv1.ComponentBuildStatus{
			State: crv1.ComponentBuildStateRunning,
		}
	}

	// No Job Pods have been scheduled yet, so the ComponentBuild is still "queued".
	return crv1.ComponentBuildStatus{
		State: crv1.ComponentBuildStateQueued,
	}
}
