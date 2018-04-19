package systemlifecycle

import (
	"fmt"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type lifecycleAction struct {
	rollout  *latticev1.Deploy
	teardown *latticev1.Teardown
}

type Controller struct {
	syncHandler func(sysRolloutKey string) error

	namespacePrefix string

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	// Each LatticeNamespace may only have one rollout going on at a time.
	// A rollout is an "owning" rollout while it is rolling out, and until
	// it has completed and a new rollout has been accepted and becomes the
	// owning rollout. (e.g. we create Deploy A. It is accepted and becomes
	// the owning rollout. It then runs to completion. It is still the owning
	// rollout. Then Deploy B is created. It is accepted because the existing
	// owning rollout is not running. Now B is the owning rollout)
	// FIXME: rethink this. is there a simpler solution?
	owningLifecycleActionsLock   sync.RWMutex
	owningLifecycleActions       map[types.UID]*lifecycleAction
	owningLifecycleActionsSynced chan struct{}

	deployLister       latticelisters.DeployLister
	deployListerSynced cache.InformerSynced

	teardownLister       latticelisters.TeardownLister
	teardownListerSynced cache.InformerSynced

	systemLister       latticelisters.SystemLister
	systemListerSynced cache.InformerSynced

	buildLister       latticelisters.BuildLister
	buildListerSynced cache.InformerSynced

	serviceBuildLister       latticelisters.ServiceBuildLister
	serviceBuildListerSynced cache.InformerSynced

	componentBuildLister       latticelisters.ComponentBuildLister
	componentBuildListerSynced cache.InformerSynced

	rolloutQueue  workqueue.RateLimitingInterface
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
	serviceBuildInformer latticeinformers.ServiceBuildInformer,
	componentBuildInformer latticeinformers.ComponentBuildInformer,
) *Controller {
	src := &Controller{
		namespacePrefix:              namespacePrefix,
		kubeClient:                   kubeClient,
		latticeClient:                latticeClient,
		owningLifecycleActions:       make(map[types.UID]*lifecycleAction),
		owningLifecycleActionsSynced: make(chan struct{}),
		rolloutQueue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-rollout"),
		teardownQueue:                workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-teardown"),
	}

	src.syncHandler = src.syncDeploy

	deployInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleDeployAdd,
		UpdateFunc: src.handleDeployUpdate,
		// TODO: for now it is assumed that SystemRollouts are not deleted. Revisit this.
	})
	src.deployLister = deployInformer.Lister()
	src.deployListerSynced = deployInformer.Informer().HasSynced

	teardownInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleTeardownAdd,
		UpdateFunc: src.handleTeardownUpdate,
		// TODO: for now it is assumed that SystemRollouts are not deleted. Revisit this.
	})
	src.teardownLister = teardownInformer.Lister()
	src.teardownListerSynced = teardownInformer.Informer().HasSynced

	systemInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleSystemAdd,
		UpdateFunc: src.handleSystemUpdate,
		// TODO: for now it is assumed that Systems are not deleted. Revisit this.
	})
	src.systemLister = systemInformer.Lister()
	src.systemListerSynced = systemInformer.Informer().HasSynced

	buildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleBuildAdd,
		UpdateFunc: src.handleBuildUpdate,
		// TODO: for now it is assumed that SystemBuilds are not deleted. Revisit this.
	})
	src.buildLister = buildInformer.Lister()
	src.buildListerSynced = buildInformer.Informer().HasSynced

	src.serviceBuildLister = serviceBuildInformer.Lister()
	src.serviceBuildListerSynced = serviceBuildInformer.Informer().HasSynced

	src.componentBuildLister = componentBuildInformer.Lister()
	src.componentBuildListerSynced = componentBuildInformer.Informer().HasSynced

	return src
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.rolloutQueue.ShutDown()
	defer c.teardownQueue.ShutDown()

	glog.Infof("Starting system-rollout controller")
	defer glog.Infof("Shutting down system-rollout controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(
		stopCh,
		c.deployListerSynced,
		c.teardownListerSynced,
		c.systemListerSynced,
		c.buildListerSynced,
		c.serviceBuildListerSynced,
		c.componentBuildListerSynced,
	) {
		return
	}

	glog.V(4).Info("Caches synced. Syncing owning SystemRollouts")

	// It's okay that we're racing with the System and Build informer add/update functions here.
	// handleDeployAdd and handleDeployUpdate will enqueue all of the existing SystemRollouts already
	// so it's okay if the other informers don't.
	if err := c.syncOwningActions(); err != nil {
		glog.Errorf("error syncing owning actions: %v", err)
		return
	}

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(c.runRolloutWorker, time.Second, stopCh)
	}

	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(c.runTeardownWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (c *Controller) handleDeployAdd(obj interface{}) {
	deploy := obj.(*latticev1.Deploy)
	glog.V(4).Infof("Adding Deploy %v/%v", deploy.Namespace, deploy.Name)
	c.enqueueDeploy(deploy)
}

func (c *Controller) handleDeployUpdate(old, cur interface{}) {
	oldDeploy := old.(*latticev1.Deploy)
	curDeploy := cur.(*latticev1.Deploy)
	if oldDeploy.ResourceVersion == curDeploy.ResourceVersion {
		// Periodic resync will send update events for all known Deploy.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("Deploy ResourceVersions are the same")
		return
	}

	glog.V(4).Infof("Updating Deploy %s/%s", curDeploy.Namespace, curDeploy.Name)
	c.enqueueDeploy(curDeploy)
}

func (c *Controller) handleTeardownAdd(obj interface{}) {
	teardown := obj.(*latticev1.Teardown)
	glog.V(4).Infof("Adding Teardown %s", teardown.Name)
	c.enqueueTeardown(teardown)
}

func (c *Controller) handleTeardownUpdate(old, cur interface{}) {
	oldTeardown := old.(*latticev1.Teardown)
	curTeardown := cur.(*latticev1.Teardown)
	if oldTeardown.ResourceVersion == curTeardown.ResourceVersion {
		// Periodic resync will send update events for all known Deploy.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("Deploy ResourceVersions are the same")
		return
	}

	glog.V(4).Infof("Updating Teardown %s", oldTeardown.Name)
	c.enqueueTeardown(curTeardown)
}

func (c *Controller) handleSystemAdd(obj interface{}) {
	<-c.owningLifecycleActionsSynced

	system := obj.(*latticev1.System)
	glog.V(4).Infof("System %s added", system.Name)

	systemNamespace := kubeutil.SystemNamespace(c.namespacePrefix, v1.SystemID(system.Name))
	action, exists := c.getOwningAction(systemNamespace)
	if !exists {
		// No ongoing action
		glog.V(4).Infof("System %v/%v has no owning actions, skipping", system.Namespace, system.Name)
		return
	}

	if action == nil {
		// FIXME: send warn event
		return
	}

	if action.rollout != nil {
		c.enqueueDeploy(action.rollout)
		return
	}

	if action.teardown != nil {
		c.enqueueTeardown(action.teardown)
		return
	}

	// FIXME: Send warn event
}

func (c *Controller) handleSystemUpdate(old, cur interface{}) {
	<-c.owningLifecycleActionsSynced

	glog.V(4).Info("Got System update")
	oldSystem := old.(*latticev1.System)
	curSystem := cur.(*latticev1.System)
	if oldSystem.ResourceVersion == curSystem.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("System ResourceVersions are the same")
		return
	}

	systemNamespace := kubeutil.SystemNamespace(c.namespacePrefix, v1.SystemID(curSystem.Name))
	action, exists := c.getOwningAction(systemNamespace)
	if !exists {
		glog.V(4).Infof("System %v/%v has no owning actions, skipping", curSystem.Namespace, curSystem.Name)
		// No ongoing action
		return
	}

	if action == nil {
		// FIXME: send warn event
		return
	}

	if action.rollout != nil {
		c.enqueueDeploy(action.rollout)
		return
	}

	if action.teardown != nil {
		c.enqueueTeardown(action.teardown)
		return
	}

	// FIXME: Send warn event
}

func (c *Controller) handleBuildAdd(obj interface{}) {
	<-c.owningLifecycleActionsSynced

	build := obj.(*latticev1.Build)
	glog.V(4).Infof("Build %s added", build.Name)

	action, exists := c.getOwningAction(build.Namespace)
	if !exists {
		// No ongoing action
		return
	}

	if action == nil {
		// FIXME: send warn event
		return
	}

	if action.rollout != nil {
		c.enqueueDeploy(action.rollout)
		return
	}

	// only need to update rollouts on builds finishing
}

func (c *Controller) handleBuildUpdate(old, cur interface{}) {
	<-c.owningLifecycleActionsSynced

	glog.V(4).Infof("Got Build update")
	oldBuild := old.(*latticev1.Build)
	curBuild := cur.(*latticev1.Build)
	if oldBuild.ResourceVersion == curBuild.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("Build ResourceVersions are the same")
		return
	}

	action, exists := c.getOwningAction(curBuild.Namespace)
	if !exists {
		// No ongoing action
		return
	}

	if action == nil {
		// FIXME: send warn event
		return
	}

	if action.rollout != nil {
		c.enqueueDeploy(action.rollout)
		return
	}

	// only need to update rollouts on builds finishing
}

func (c *Controller) enqueueDeploy(deploy *latticev1.Deploy) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(deploy)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", deploy, err))
		return
	}

	c.rolloutQueue.Add(key)
}

func (c *Controller) enqueueTeardown(teardown *latticev1.Teardown) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(teardown)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", teardown, err))
		return
	}

	c.teardownQueue.Add(key)
}

func (c *Controller) syncOwningActions() error {
	c.owningLifecycleActionsLock.Lock()
	defer c.owningLifecycleActionsLock.Unlock()

	rollouts, err := c.deployLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, rollout := range rollouts {
		if rollout.Status.State != latticev1.DeployStateInProgress {
			continue
		}

		namespace, err := c.kubeClient.CoreV1().Namespaces().Get(rollout.Namespace, metav1.GetOptions{})
		if err != nil {
			return err
		}

		_, exists := c.owningLifecycleActions[namespace.UID]
		if exists {
			return fmt.Errorf("System %v has multiple owning actions", rollout.Namespace)
		}

		c.owningLifecycleActions[namespace.UID] = &lifecycleAction{
			rollout: rollout,
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

		namespace, err := c.kubeClient.CoreV1().Namespaces().Get(teardown.Namespace, metav1.GetOptions{})
		if err != nil {
			return err
		}

		_, exists := c.owningLifecycleActions[namespace.UID]
		if exists {
			return fmt.Errorf("System %v has multiple owning actions", teardown.Namespace)
		}

		c.owningLifecycleActions[namespace.UID] = &lifecycleAction{
			teardown: teardown,
		}
	}

	close(c.owningLifecycleActionsSynced)
	return nil
}

func (c *Controller) runRolloutWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for c.processNextWorkItem(c.rolloutQueue, c.syncDeploy) {
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
	glog.V(4).Infof("Started syncing Deploy %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing Deploy %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	deploy, err := c.deployLister.Deploies(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("Deploy %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	glog.V(5).Infof("Deploy %v state: %v", key, deploy.Status.State)

	switch deploy.Status.State {
	case latticev1.DeployStateSucceeded, latticev1.DeployStateFailed:
		glog.V(4).Infof("Deploy %s already completed", key)
		return nil

	case latticev1.DeployStateInProgress:
		return c.syncInProgressDeploy(deploy)

	case latticev1.DeployStateAccepted:
		return c.syncAcceptedDeploy(deploy)

	case latticev1.DeployStatePending:
		return c.syncPendingDeploy(deploy)

	default:
		return fmt.Errorf("Deploy %v has unexpected state: %v", key, deploy.Status.State)
	}
}

// syncSystemBuild will sync the Build with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncTeardown(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing SystemTeardown %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing SystemTeardown %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	teardown, err := c.teardownLister.Teardowns(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("SystemTeardown %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	switch teardown.Status.State {
	case latticev1.TeardownStateSucceeded, latticev1.TeardownStateFailed:
		glog.V(4).Infof("SystemTeardown %s already completed", key)
		return nil

	case latticev1.TeardownStateInProgress:
		return c.syncInProgressTeardown(teardown)

	case latticev1.TeardownStatePending:
		return c.syncPendingTeardown(teardown)

	default:
		return fmt.Errorf("SystemTeardown %v has unexpected state: %v", key, teardown.Status.State)
	}
}
