package systemlifecycle

import (
	"fmt"
	"time"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	syncutil "github.com/mlab-lattice/lattice/pkg/util/sync"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	coreinformers "k8s.io/client-go/informers/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type Controller struct {
	syncHandler func(sysRolloutKey string) error

	namespacePrefix string

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	lifecycleActions       *syncutil.LifecycleActionManager
	lifecycleActionsSynced chan struct{}

	deployLister       latticelisters.DeployLister
	deployListerSynced cache.InformerSynced

	teardownLister       latticelisters.TeardownLister
	teardownListerSynced cache.InformerSynced

	systemLister       latticelisters.SystemLister
	systemListerSynced cache.InformerSynced

	buildLister       latticelisters.BuildLister
	buildListerSynced cache.InformerSynced

	containerBuildLister       latticelisters.ContainerBuildLister
	containerBuildListerSynced cache.InformerSynced

	kubeNamespaceLister       corelisters.NamespaceLister
	kubeNamespaceListerSynced cache.InformerSynced

	deployQueue   workqueue.RateLimitingInterface
	teardownQueue workqueue.RateLimitingInterface
}

func NewController(
	namespacePrefix string,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	deployInformer latticeinformers.DeployInformer,
	teardownInformer latticeinformers.TeardownInformer,
	systemInformer latticeinformers.SystemInformer,
	buildInformer latticeinformers.BuildInformer,
	containerBuildInformer latticeinformers.ContainerBuildInformer,
	kubeNamespaceInformer coreinformers.NamespaceInformer,
) *Controller {
	c := &Controller{
		namespacePrefix: namespacePrefix,

		kubeClient:    kubeClient,
		latticeClient: latticeClient,

		lifecycleActions:       syncutil.NewLifecycleActionManager(),
		lifecycleActionsSynced: make(chan struct{}),

		deployQueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "deploy"),
		teardownQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "teardown"),
	}

	c.syncHandler = c.syncDeploy

	deployInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleDeployAdd,
		UpdateFunc: c.handleDeployUpdate,
		// TODO: for now it is assumed that SystemRollouts are not deleted. Revisit this.
	})
	c.deployLister = deployInformer.Lister()
	c.deployListerSynced = deployInformer.Informer().HasSynced

	teardownInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleTeardownAdd,
		UpdateFunc: c.handleTeardownUpdate,
		// TODO: for now it is assumed that SystemRollouts are not deleted. Revisit this.
	})
	c.teardownLister = teardownInformer.Lister()
	c.teardownListerSynced = teardownInformer.Informer().HasSynced

	systemInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSystemAdd,
		UpdateFunc: c.handleSystemUpdate,
		// TODO: for now it is assumed that Systems are not deleted. Revisit this.
	})
	c.systemLister = systemInformer.Lister()
	c.systemListerSynced = systemInformer.Informer().HasSynced

	buildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleBuildAdd,
		UpdateFunc: c.handleBuildUpdate,
		// TODO: for now it is assumed that SystemBuilds are not deleted. Revisit this.
	})
	c.buildLister = buildInformer.Lister()
	c.buildListerSynced = buildInformer.Informer().HasSynced

	c.containerBuildLister = containerBuildInformer.Lister()
	c.containerBuildListerSynced = containerBuildInformer.Informer().HasSynced

	c.kubeNamespaceLister = kubeNamespaceInformer.Lister()
	c.kubeNamespaceListerSynced = kubeNamespaceInformer.Informer().HasSynced

	return c
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.deployQueue.ShutDown()
	defer c.teardownQueue.ShutDown()

	glog.Infof("starting system lifecycle controller")
	defer glog.Infof("shutting down system lifecycle controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(
		stopCh,
		c.deployListerSynced,
		c.teardownListerSynced,
		c.systemListerSynced,
		c.buildListerSynced,
		c.containerBuildListerSynced,
		c.kubeNamespaceListerSynced,
	) {
		return
	}

	glog.V(4).Info("caches synced, syncing owning lifecycle actions")

	// It's okay that we're racing with the System and Build informer add/update functions here.
	// handleDeployAdd and handleDeployUpdate will enqueue all of the existing SystemRollouts already
	// so it's okay if the other informers don't.
	if err := c.syncLifecycleActions(); err != nil {
		glog.Errorf("error syncing owning actions: %v", err)
		return
	}

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(c.runDeployWorker, time.Second, stopCh)
	}

	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(c.runTeardownWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (c *Controller) enqueueDeploy(deploy *latticev1.Deploy) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(deploy)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", deploy, err))
		return
	}

	c.deployQueue.Add(key)
}

func (c *Controller) enqueueTeardown(teardown *latticev1.Teardown) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(teardown)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", teardown, err))
		return
	}

	c.teardownQueue.Add(key)
}

func (c *Controller) syncLifecycleActions() error {
	deploys, err := c.deployLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, deploy := range deploys {
		if deploy.Status.State != latticev1.DeployStateInProgress {
			continue
		}

		if deploy.Status.BuildID == nil {
			// this shouldn't happen
			continue
		}

		build, err := c.buildLister.Builds(deploy.Namespace).Get(string(*deploy.Status.BuildID))
		if err != nil {
			return err
		}

		path := tree.RootPath()
		if build.Spec.Path != nil {
			path = *build.Spec.Path
		}

		err = c.acquireDeployLock(deploy, path)
		if err != nil {
			return fmt.Errorf(
				"error attempting to acquire lock for %v %v: %v",
				deploy.Description(c.namespacePrefix),
				build.Spec.Path.String(),
				err,
			)
		}
	}

	teardowns, err := c.teardownLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, teardown := range teardowns {
		if teardown.Status.State != latticev1.TeardownStateInProgress {
			continue
		}

		err = c.acquireTeardownLock(teardown)
		if err != nil {
			return fmt.Errorf(
				"error attempting to acquire lock for %v: %v",
				teardown.Description(c.namespacePrefix),
				err,
			)
		}
	}

	close(c.lifecycleActionsSynced)
	return nil
}

func (c *Controller) runDeployWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for c.processNextWorkItem(c.deployQueue, c.syncDeploy) {
	}
}

func (c *Controller) runTeardownWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for c.processNextWorkItem(c.teardownQueue, c.syncTeardown) {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (c *Controller) processNextWorkItem(queue workqueue.RateLimitingInterface, syncHandler func(string) error) bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		queue.Forget(key)
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
	queue.AddRateLimited(key)

	return true
}

// syncSystemBuild will sync the Build with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncDeploy(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("started syncing deploy %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("finished syncing deploy %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	deploy, err := c.deployLister.Deploys(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("deploy %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	glog.V(5).Infof("%v state: %v", deploy.Description(c.namespacePrefix), deploy.Status.State)

	switch deploy.Status.State {
	case latticev1.DeployStateSucceeded, latticev1.DeployStateFailed:
		glog.V(4).Infof("%v already completed", deploy.Description(c.namespacePrefix))
		return nil

	case latticev1.DeployStateInProgress:
		return c.syncInProgressDeploy(deploy)

	case latticev1.DeployStateAccepted:
		return c.syncAcceptedDeploy(deploy)

	case latticev1.DeployStatePending:
		return c.syncPendingDeploy(deploy)

	default:
		return fmt.Errorf("%v has unexpected state: %v", deploy.Description(c.namespacePrefix), deploy.Status.State)
	}
}

// syncSystemBuild will sync the Build with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncTeardown(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("started syncing teardown %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("finished syncing teardown %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	teardown, err := c.teardownLister.Teardowns(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("teardown %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	switch teardown.Status.State {
	case latticev1.TeardownStateSucceeded, latticev1.TeardownStateFailed:
		glog.V(4).Infof("%v already completed", teardown.Description(c.namespacePrefix))
		return nil

	case latticev1.TeardownStateInProgress:
		return c.syncInProgressTeardown(teardown)

	case latticev1.TeardownStatePending:
		return c.syncPendingTeardown(teardown)

	default:
		return fmt.Errorf("%v has unexpected state: %v", teardown.Description(c.namespacePrefix), teardown.Status.State)
	}
}
