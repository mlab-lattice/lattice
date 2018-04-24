package servicebuild

import (
	"fmt"
	"time"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/deckarep/golang-set"
	"github.com/golang/glog"
)

type Controller struct {
	syncHandler func(svcBuildKey string) error
	enqueue     func(svcBuild *latticev1.ServiceBuild)

	namespacePrefix string

	latticeClient latticeclientset.Interface

	serviceBuildLister       latticelisters.ServiceBuildLister
	serviceBuildListerSynced cache.InformerSynced

	componentBuildLister       latticelisters.ComponentBuildLister
	componentBuildListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	namespacePrefix string,
	latticeClient latticeclientset.Interface,
	serviceBuildInformer latticeinformers.ServiceBuildInformer,
	componentBuildInformer latticeinformers.ComponentBuildInformer,
) *Controller {
	sbc := &Controller{
		namespacePrefix: namespacePrefix,

		latticeClient: latticeClient,

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-build"),
	}

	sbc.syncHandler = sbc.syncServiceBuild
	sbc.enqueue = sbc.enqueueServiceBuild

	serviceBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.handleServiceBuildAdd,
		UpdateFunc: sbc.handleServiceBuildUpdate,
		// TODO: for now it is assumed that ServiceBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ServiceBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document SvcB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.serviceBuildLister = serviceBuildInformer.Lister()
	sbc.serviceBuildListerSynced = serviceBuildInformer.Informer().HasSynced

	componentBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.handleComponentBuildAdd,
		UpdateFunc: sbc.handleComponentBuildUpdate,
		// TODO: for now it is assumed that ComponentBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ComponentBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document CB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.componentBuildLister = componentBuildInformer.Lister()
	sbc.componentBuildListerSynced = componentBuildInformer.Informer().HasSynced

	return sbc
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting service-build controller")
	defer glog.Infof("Shutting down service-build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.serviceBuildListerSynced, c.componentBuildListerSynced) {
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

func (c *Controller) handleServiceBuildAdd(obj interface{}) {
	build := obj.(*latticev1.ServiceBuild)
	glog.V(4).Infof("Adding %s", build.Description(c.namespacePrefix))
	c.enqueueServiceBuild(build)
}

func (c *Controller) handleServiceBuildUpdate(old, cur interface{}) {
	build := cur.(*latticev1.ServiceBuild)
	glog.V(4).Infof("Updating %s", build.Description(c.namespacePrefix))
	c.enqueueServiceBuild(build)
}

// handleComponentBuildAdd enqueues any ServiceBuilds which may be interested in it when
// a ComponentBuild is added.
func (c *Controller) handleComponentBuildAdd(obj interface{}) {
	componentBuild := obj.(*latticev1.ComponentBuild)

	if componentBuild.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		// component builds should only be deleted if there are no service builds
		// pointing at it, so no need to enqueue anything
		return
	}

	glog.V(4).Infof("%s added", componentBuild.Description(c.namespacePrefix))
	builds, err := c.owningServiceBuilds(componentBuild)
	if err != nil {
		// FIXME: send error event?
		return
	}
	for _, build := range builds {
		c.enqueueServiceBuild(&build)
	}
}

// handleComponentBuildUpdate enqueues any ServiceBuilds which may be interested in it when
// a ComponentBuild is updated.
func (c *Controller) handleComponentBuildUpdate(old, cur interface{}) {
	componentBuild := cur.(*latticev1.ComponentBuild)

	glog.V(4).Infof("%s updated", componentBuild.Description(c.namespacePrefix))
	builds, err := c.owningServiceBuilds(componentBuild)
	if err != nil {
		// FIXME: send error event?
		return
	}

	for _, build := range builds {
		c.enqueueServiceBuild(&build)
	}
}

func (c *Controller) owningServiceBuilds(componentBuild *latticev1.ComponentBuild) ([]latticev1.ServiceBuild, error) {
	owningBuilds := mapset.NewSet()
	for _, owner := range componentBuild.OwnerReferences {
		// Not a lattice.mlab.com owner (probably shouldn't happen)
		if owner.APIVersion != latticev1.SchemeGroupVersion.String() {
			continue
		}

		// Not a service build owner (probably shouldn't happen)
		if owner.Kind != latticev1.ServiceBuildKind.Kind {
			continue
		}

		owningBuilds.Add(owner.UID)
	}

	builds, err := c.serviceBuildLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var matchingBuilds []latticev1.ServiceBuild
	for _, build := range builds {
		if owningBuilds.Contains(build.UID) {
			matchingBuilds = append(matchingBuilds, *build)
		}
	}

	return matchingBuilds, nil
}

func (c *Controller) enqueueServiceBuild(build *latticev1.ServiceBuild) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(build)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", build, err))
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

// syncServiceBuild will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncServiceBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing ServiceBuild %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing ServiceBuild %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	build, err := c.serviceBuildLister.ServiceBuilds(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("ServiceBuild %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	stateInfo, err := c.calculateState(build)
	if err != nil {
		return err
	}

	glog.V(5).Infof("ServiceBuild %v state: %v", key, stateInfo.state)

	switch stateInfo.state {
	case stateHasFailedComponentBuilds:
		return c.syncFailedServiceBuild(build, stateInfo)
	case stateHasOnlyRunningOrSucceededComponentBuilds:
		return c.syncRunningServiceBuild(build, stateInfo)
	case stateNoFailuresNeedsNewComponentBuilds:
		return c.syncMissingComponentBuildsServiceBuild(build, stateInfo)
	case stateAllComponentBuildsSucceeded:
		return c.syncSucceededServiceBuild(build, stateInfo)
	default:
		return fmt.Errorf("ServiceBuild %v in unexpected state %v", key, stateInfo.state)
	}
}
