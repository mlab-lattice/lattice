package build

import (
	"fmt"
	"time"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type Controller struct {
	syncHandler func(key string) error
	enqueue     func(build *latticev1.Build)

	namespacePrefix string

	latticeClient latticeclientset.Interface

	componentResolver resolver.Interface

	systemLister       latticelisters.SystemLister
	systemListerSynced cache.InformerSynced

	buildLister       latticelisters.BuildLister
	buildListerSynced cache.InformerSynced

	containerBuildLister       latticelisters.ContainerBuildLister
	containerBuildListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	namespacePrefix string,
	latticeClient latticeclientset.Interface,
	componentResolver resolver.Interface,
	systemInformer latticeinformers.SystemInformer,
	buildInformer latticeinformers.BuildInformer,
	containerBuildInformer latticeinformers.ContainerBuildInformer,
) *Controller {
	sbc := &Controller{
		namespacePrefix: namespacePrefix,

		latticeClient: latticeClient,

		componentResolver: componentResolver,

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "build"),
	}

	sbc.enqueue = sbc.enqueueBuild
	sbc.syncHandler = sbc.syncSystemBuild

	sbc.systemLister = systemInformer.Lister()
	sbc.systemListerSynced = systemInformer.Informer().HasSynced

	buildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.handleBuildAdd,
		UpdateFunc: sbc.handleBuildUpdate,
		DeleteFunc: sbc.handleBuildDelete,
	})
	sbc.buildLister = buildInformer.Lister()
	sbc.buildListerSynced = buildInformer.Informer().HasSynced

	containerBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.handleContainerBuildAdd,
		UpdateFunc: sbc.handleContainerBuildUpdate,
		// only orphaned service builds should be deleted
	})
	sbc.containerBuildLister = containerBuildInformer.Lister()
	sbc.containerBuildListerSynced = containerBuildInformer.Informer().HasSynced

	return sbc
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("starting build controller")
	defer glog.Infof("shutting down build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.systemListerSynced, c.buildListerSynced, c.containerBuildListerSynced) {
		return
	}

	glog.V(4).Info("caches synced")

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

func (c *Controller) enqueueBuild(sysb *latticev1.Build) {
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

// syncSystemBuild will sync the Build with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncSystemBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("started syncing build %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("finished syncing build %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	build, err := c.buildLister.Builds(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("build %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	// TODO: when adding delete build to the api, make sure to make it an orphaning delete
	// see pkg/backend/kubernetes/controller/servicebuild/service_build.go:deleteServiceBuild for reasoning
	if build.DeletionTimestamp != nil {
		return c.syncDeletedBuild(build)
	}

	build, err = c.addFinalizer(build)
	if err != nil {
		return err
	}

	if build.Status.State == latticev1.BuildStatePending {
		return c.syncPendingBuild(build)
	}

	stateInfo, err := c.calculateState(build)
	if err != nil {
		return err
	}

	glog.V(5).Infof("%v state: %v", build.Description(c.namespacePrefix), stateInfo.state)

	switch stateInfo.state {
	case stateHasFailedContainerBuilds:
		return c.syncFailedBuild(build, stateInfo)
	case stateHasOnlyRunningOrSucceededContainerBuilds:
		return c.syncRunningBuild(build, stateInfo)
	case stateNoFailuresNeedsNewContainerBuilds:
		return c.syncMissingContainerBuildsBuild(build, stateInfo)
	case stateAllContainerBuildsSucceeded:
		return c.syncSucceededBuild(build, stateInfo)
	default:
		return fmt.Errorf("%v in unrecognized state %v", build.Description(c.namespacePrefix), stateInfo.state)
	}
}
