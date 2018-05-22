package componentbuild

import (
	"fmt"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type Controller struct {
	syncHandler func(key string) error
	enqueue     func(build *latticev1.ComponentBuild)

	namespacePrefix string

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	kubeInformerFactory    kubeinformers.SharedInformerFactory
	latticeInformerFactory latticeinformers.SharedInformerFactory

	// NOTE: you must get a read lock on the configLock for the duration
	//       of your use of the cloudProvider
	staticCloudProviderOptions *cloudprovider.Options
	cloudProvider              cloudprovider.Interface

	configLister       latticelisters.ConfigLister
	configListerSynced cache.InformerSynced
	configSetChan      chan struct{}
	configSet          bool
	configLock         sync.RWMutex
	config             latticev1.ConfigSpec

	componentBuildLister       latticelisters.ComponentBuildLister
	componentBuildListerSynced cache.InformerSynced

	jobLister       batchlisters.JobLister
	jobListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	namespacePrefix string,
	cloudProviderOptions *cloudprovider.Options,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	latticeInformerFactory latticeinformers.SharedInformerFactory,
) *Controller {
	c := &Controller{
		namespacePrefix: namespacePrefix,

		kubeClient:    kubeClient,
		latticeClient: latticeClient,

		kubeInformerFactory:    kubeInformerFactory,
		latticeInformerFactory: latticeInformerFactory,

		staticCloudProviderOptions: cloudProviderOptions,

		configSetChan: make(chan struct{}),

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "componentbuild"),
	}

	c.syncHandler = c.syncComponentBuild
	c.enqueue = c.enqueueComponentBuild

	configInformer := latticeInformerFactory.Lattice().V1().Configs()
	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    c.handleConfigAdd,
		UpdateFunc: c.handleConfigUpdate,
	})
	c.configLister = configInformer.Lister()
	c.configListerSynced = configInformer.Informer().HasSynced

	componentBuildInformer := latticeInformerFactory.Lattice().V1().ComponentBuilds()
	componentBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleComponentBuildAdd,
		UpdateFunc: c.handleComponentBuildUpdate,
		// nothing to be done for deleted component builds
	})
	c.componentBuildLister = componentBuildInformer.Lister()
	c.componentBuildListerSynced = componentBuildInformer.Informer().HasSynced

	jobInformer := kubeInformerFactory.Batch().V1().Jobs()
	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleJobAdd,
		UpdateFunc: c.handleJobUpdate,
		// Job deletions we care about should only happen via a component build
		// being deleted
	})
	c.jobLister = jobInformer.Lister()
	c.jobListerSynced = jobInformer.Informer().HasSynced

	return c
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting component-build controller")
	defer glog.Infof("Shutting down component-build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.configListerSynced, c.componentBuildListerSynced, c.jobListerSynced) {
		return
	}

	glog.V(4).Info("caches synced, waiting for config to be set")

	// wait for config to be set
	<-c.configSetChan

	glog.V(4).Info("config set")

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

func (c *Controller) enqueueComponentBuild(build *latticev1.ComponentBuild) {
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

// syncComponentBuild will sync the ComponentBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncComponentBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("started syncing component build %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("finished syncing component build %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	build, err := c.componentBuildLister.ComponentBuilds(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("component build %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	// nothing to do if the component build is being deleted
	// TODO: should we clean up container artifacts as well?
	if build.DeletionTimestamp != nil {
		glog.V(5).Infof("%v is being deleted", build.Description(c.namespacePrefix))
		return nil
	}

	if isOrphaned(build) {
		return c.syncOrphanedComponentBuild(build)
	}

	stateInfo, err := c.calculateState(build)
	if err != nil {
		return err
	}

	glog.V(5).Infof("%v state: %v", build.Description(c.namespacePrefix), stateInfo.state)

	switch stateInfo.state {
	case stateJobNotCreated:
		return c.syncJoblessComponentBuild(build)
	case stateJobSucceeded:
		return c.syncSuccessfulComponentBuild(build, stateInfo.job)
	case stateJobFailed:
		return c.syncFailedComponentBuild(build)
	case stateJobRunning:
		return c.syncUnfinishedComponentBuild(build, stateInfo.job)
	default:
		return fmt.Errorf("%v in unexpected state %v", build.Description(c.namespacePrefix), stateInfo.state)
	}
}
