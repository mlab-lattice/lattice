package build

import (
	"fmt"
	"reflect"
	"time"

	"github.com/golang/glog"

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

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("Build")

type BuildController struct {
	syncHandler  func(bKey string) error
	enqueueBuild func(build *crv1.Build)

	latticeResourceRestClient rest.Interface
	kubeClient                clientset.Interface

	buildStore       cache.Store
	buildStoreSynced cache.InformerSynced

	jobLister       batchlisters.JobLister
	jobListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewBuildController(
	latticeResourceRestClient rest.Interface,
	bInformer cache.SharedInformer,
	jInformer batchinformers.JobInformer,
	kubeClient clientset.Interface,
) *BuildController {
	bc := &BuildController{
		latticeResourceRestClient: latticeResourceRestClient,
		kubeClient:                kubeClient,
		queue:                     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "build"),
	}

	bc.syncHandler = bc.syncBuild
	bc.enqueueBuild = bc.enqueue

	bInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    bc.addBuild,
		UpdateFunc: bc.updateBuild,
		// This will enter the sync loop and no-op, because the deployment has been deleted from the store.
		DeleteFunc: bc.deleteBuild,
	})
	bc.buildStore = bInformer.GetStore()
	bc.buildStoreSynced = bInformer.HasSynced

	jInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    bc.addJob,
		UpdateFunc: bc.updateJob,
		DeleteFunc: bc.deleteJob,
	})
	bc.jobLister = jInformer.Lister()
	bc.jobListerSynced = jInformer.Informer().HasSynced

	return bc
}

func (bc *BuildController) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer bc.queue.ShutDown()

	glog.Infof("Starting build controller")
	defer glog.Infof("Shutting down build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, bc.buildStoreSynced, bc.jobListerSynced) {
		return
	}

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(bc.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (bc *BuildController) addBuild(obj interface{}) {
	b := obj.(*crv1.Build)
	glog.V(4).Infof("Adding Build %s", b.Name)
	bc.enqueueBuild(b)
}

func (bc *BuildController) updateBuild(old, cur interface{}) {
	oldB := old.(*crv1.Build)
	curB := cur.(*crv1.Build)
	glog.V(4).Infof("Updating Build %s", oldB.Name)
	bc.enqueueBuild(curB)
}

func (bc *BuildController) deleteBuild(obj interface{}) {
	d, ok := obj.(*crv1.Build)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*crv1.Build)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a Build %#v", obj))
			return
		}
	}
	glog.V(4).Infof("Deleting Build %s", d.Name)
	bc.enqueueBuild(d)
}

// addJob enqueues the Build that manages a Job when the Job is created.
func (bc *BuildController) addJob(obj interface{}) {
	job := obj.(*batchv1.Job)

	if job.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		bc.deleteJob(job)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(job); controllerRef != nil {
		b := bc.resolveControllerRef(job.Namespace, controllerRef)

		// Not a Build Job
		if b == nil {
			return
		}

		glog.V(4).Infof("Job %s added.", job.Name)
		bc.enqueueBuild(b)
		return
	}

	// Otherwise, it's an orphan. We don't care about these since Build Jobs should
	// always have a ControllerRef, so therefore this Job is not a Build Job.
}

// updateJob figures out what Build manages a Job when the Job
// is updated and wake them up.
// Note that a Build Job should only ever and should always be owned by a single Build
// (the one referenced in its ControllerRef), so an updated job should
// have the same controller ref for both the old and current versions.
func (bc *BuildController) updateJob(old, cur interface{}) {
	glog.V(5).Info("Got update job")
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
		// The ControllerRef was changed. If this is a Build Job, this shouldn't happen.
		if b := bc.resolveControllerRef(oldJob.Namespace, oldControllerRef); b != nil {
			// TODO: send warn event here, this should not happen
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		b := bc.resolveControllerRef(curJob.Namespace, curControllerRef)

		// Not a Build Job
		if b == nil {
			return
		}

		bc.enqueueBuild(b)
		return
	}

	// Otherwise, it's an orphan. We don't care about these since Build Jobs should
	// always have a ControllerRef, so therefore this Job is not a Build Job.
}

// deleteJob enqueues the Build that manages a Job when
// the Job is deleted. obj could be an *extensions.Build, or
// a DeletionFinalStateUnknown marker item.
// Note that this should never really happen. Build Jobs should not
// be getting deleted. But if they do, we should write a warn event
// and update the Build.
func (bc *BuildController) deleteJob(obj interface{}) {
	job, ok := obj.(*batchv1.Job)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale. If the ReplicaSet
	// changed labels the new deployment will not be woken up till the periodic resync.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		job, ok = tombstone.Obj.(*batchv1.Job)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a ReplicaSet %#v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(job)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	b := bc.resolveControllerRef(job.Namespace, controllerRef)

	// Not a Build job
	if b == nil {
		return
	}

	glog.V(4).Infof("Job %s deleted.", job.Name)
	bc.enqueueBuild(b)
}

func (bc *BuildController) enqueue(build *crv1.Build) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(build)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", build, err))
		return
	}

	bc.queue.Add(key)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (bc *BuildController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.Build {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}

	buildKey := fmt.Sprintf("%v/%v", namespace, controllerRef.Name)
	bi, exists, err := bc.buildStore.GetByKey(buildKey)
	if err != nil || !exists {
		return nil
	}

	b := bi.(*crv1.Build)

	if b.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return b
}

// getJobForBuild uses ControllerRefManager to retrieve the Job for a Build
func (bc *BuildController) getJobForBuild(b *crv1.Build) (*batchv1.Job, error) {
	// List all Jobs to find in the Build's namespace to find the Job the Build manages.
	jobList, err := bc.jobLister.Jobs(b.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	matchingJobs := []*batchv1.Job{}
	buildControllerRef := metav1.NewControllerRef(b, controllerKind)

	for _, job := range jobList {
		jobControllerRef := metav1.GetControllerOf(job)

		if reflect.DeepEqual(buildControllerRef, jobControllerRef) {
			matchingJobs = append(matchingJobs, job)
		}
	}

	if len(matchingJobs) == 0 {
		return nil, nil
	}

	if len(matchingJobs) > 1 {
		return nil, fmt.Errorf("Build %v has multiple Jobs", b.Name)
	}

	return matchingJobs[0], nil
}

func (bc *BuildController) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for bc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (bc *BuildController) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := bc.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer bc.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := bc.syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		bc.queue.Forget(key)
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
	bc.queue.AddRateLimited(key)

	return true
}

// syncDeployment will sync the deployment with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (bc *BuildController) syncBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing build %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing build %q (%v)", key, time.Now().Sub(startTime))
	}()

	bi, exists, err := bc.buildStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("Build %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	build := bi.(*crv1.Build)

	job, err := bc.getJobForBuild(build)
	if err != nil {
		return err
	}

	// Deep-copy otherwise we are mutating our cache.
	b := build.DeepCopy()

	if job == nil {
		return bc.createBuildJob(b)
	}

	return bc.syncBuildStatus(b, job)
}

func (bc *BuildController) createBuildJob(build *crv1.Build) error {
	job := getBuildJob(build)
	jobResp, err := bc.kubeClient.BatchV1().Jobs(build.Namespace).Create(job)
	if err != nil {
		// TODO: send warn event
		return err
	}

	glog.V(4).Infof("Created Job %s", jobResp.Name)
	// TODO: send normal event
	return bc.syncBuildStatus(build, jobResp)
}

func (bc *BuildController) syncBuildStatus(build *crv1.Build, job *batchv1.Job) error {
	newStatus := calculateBuildStatus(job)

	if reflect.DeepEqual(build.Status, newStatus) {
		return nil
	}

	build.Status = newStatus

	err := bc.latticeResourceRestClient.Put().
		Namespace(build.Namespace).
		Resource(crv1.BuildResourcePlural).
		Name(build.Name).
		Body(build).
		Do().
		Into(nil)

	return err
}

func calculateBuildStatus(job *batchv1.Job) crv1.BuildStatus {
	finished, succeeded := jobStatus(job)
	if finished {
		if succeeded {
			return crv1.BuildStatus{
				State: crv1.BuildStateSucceeded,
			}
		}

		return crv1.BuildStatus{
			State: crv1.BuildStateFailed,
		}
	}

	// The Job Pods have been able to be scheduled, so the Build is "running" even though
	// a Job Pod may not currently be active.
	if job.Status.Active > 0 || job.Status.Failed > 0 || job.Status.Succeeded > 0 {
		return crv1.BuildStatus{
			State: crv1.BuildStateRunning,
		}
	}

	// No Job Pods have been scheduled yet, so the Build is still "queued".
	return crv1.BuildStatus{
		State: crv1.BuildStateQueued,
	}
}
