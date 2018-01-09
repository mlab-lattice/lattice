package systemlifecycle

import (
	"fmt"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type lifecycleAction struct {
	rollout  *crv1.SystemRollout
	teardown *crv1.SystemTeardown
}

type Controller struct {
	syncHandler func(sysRolloutKey string) error

	latticeClient latticeclientset.Interface

	// Each LatticeNamespace may only have one rollout going on at a time.
	// A rollout is an "owning" rollout while it is rolling out, and until
	// it has completed and a new rollout has been accepted and becomes the
	// owning rollout. (e.g. we create SystemRollout A. It is accepted and becomes
	// the owning rollout. It then runs to completion. It is still the owning
	// rollout. Then SystemRollout B is created. It is accepted because the existing
	// owning rollout is not running. Now B is the owning rollout)
	owningLifecycleActionsLock   sync.RWMutex
	owningLifecycleActions       map[string]*lifecycleAction
	owningLifecycleActionsSynced chan struct{}

	systemRolloutLister       latticelisters.SystemRolloutLister
	systemRolloutListerSynced cache.InformerSynced

	systemTeardownLister       latticelisters.SystemTeardownLister
	systemTeardownListerSynced cache.InformerSynced

	systemLister       latticelisters.SystemLister
	systemListerSynced cache.InformerSynced

	systemBuildLister       latticelisters.SystemBuildLister
	systemBuildListerSynced cache.InformerSynced

	serviceBuildLister       latticelisters.ServiceBuildLister
	serviceBuildListerSynced cache.InformerSynced

	componentBuildLister       latticelisters.ComponentBuildLister
	componentBuildListerSynced cache.InformerSynced

	rolloutQueue  workqueue.RateLimitingInterface
	teardownQueue workqueue.RateLimitingInterface
}

func NewController(
	latticeClient latticeclientset.Interface,
	systemRolloutInformer latticeinformers.SystemRolloutInformer,
	systemTeardownInformer latticeinformers.SystemTeardownInformer,
	systemInformer latticeinformers.SystemInformer,
	systemBuildInformer latticeinformers.SystemBuildInformer,
	serviceBuildInformer latticeinformers.ServiceBuildInformer,
	componentBuildInformer latticeinformers.ComponentBuildInformer,
) *Controller {
	src := &Controller{
		latticeClient:                latticeClient,
		owningLifecycleActions:       make(map[string]*lifecycleAction),
		owningLifecycleActionsSynced: make(chan struct{}),
		rolloutQueue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-rollout"),
		teardownQueue:                workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-teardown"),
	}

	src.syncHandler = src.syncSystemRollout

	systemRolloutInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleSystemRolloutAdd,
		UpdateFunc: src.handleSystemRolloutUpdate,
		// TODO: for now it is assumed that SystemRollouts are not deleted. Revisit this.
	})
	src.systemRolloutLister = systemRolloutInformer.Lister()
	src.systemRolloutListerSynced = systemRolloutInformer.Informer().HasSynced

	systemTeardownInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleSystemTeardownAdd,
		UpdateFunc: src.handleSystemTeardownUpdate,
		// TODO: for now it is assumed that SystemRollouts are not deleted. Revisit this.
	})
	src.systemTeardownLister = systemTeardownInformer.Lister()
	src.systemTeardownListerSynced = systemTeardownInformer.Informer().HasSynced

	systemInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleSystemAdd,
		UpdateFunc: src.handleSystemUpdate,
		// TODO: for now it is assumed that Systems are not deleted. Revisit this.
	})
	src.systemLister = systemInformer.Lister()
	src.systemListerSynced = systemInformer.Informer().HasSynced

	systemBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleSystemBuildAdd,
		UpdateFunc: src.handleSystemBuildUpdate,
		// TODO: for now it is assumed that SystemBuilds are not deleted. Revisit this.
	})
	src.systemBuildLister = systemBuildInformer.Lister()
	src.systemBuildListerSynced = systemBuildInformer.Informer().HasSynced

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
		c.systemRolloutListerSynced,
		c.systemTeardownListerSynced,
		c.systemListerSynced,
		c.systemBuildListerSynced,
		c.serviceBuildListerSynced,
		c.componentBuildListerSynced,
	) {
		return
	}

	glog.V(4).Info("Caches synced. Syncing owning SystemRollouts")

	// It's okay that we're racing with the System and SystemBuild informer add/update functions here.
	// handleSystemRolloutAdd and handleSystemRolloutUpdate will enqueue all of the existing SystemRollouts already
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

func (c *Controller) handleSystemRolloutAdd(obj interface{}) {
	rollout := obj.(*crv1.SystemRollout)
	glog.V(4).Infof("Adding SystemRollout %v/%v", rollout.Namespace, rollout.Name)
	c.enqueueSystemRollout(rollout)
}

func (c *Controller) handleSystemRolloutUpdate(old, cur interface{}) {
	oldRollout := old.(*crv1.SystemRollout)
	curRollout := cur.(*crv1.SystemRollout)
	if oldRollout.ResourceVersion == curRollout.ResourceVersion {
		// Periodic resync will send update events for all known SystemRollout.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("SystemRollout ResourceVersions are the same")
		return
	}

	glog.V(4).Infof("Updating SystemRollout %s/%s", curRollout.Namespace, curRollout.Name)
	c.enqueueSystemRollout(curRollout)
}

func (c *Controller) handleSystemTeardownAdd(obj interface{}) {
	syst := obj.(*crv1.SystemTeardown)
	glog.V(4).Infof("Adding SystemTeardown %s", syst.Name)
	c.enqueueSystemTeardown(syst)
}

func (c *Controller) handleSystemTeardownUpdate(old, cur interface{}) {
	oldTeardown := old.(*crv1.SystemTeardown)
	curTeardown := cur.(*crv1.SystemTeardown)
	if oldTeardown.ResourceVersion == curTeardown.ResourceVersion {
		// Periodic resync will send update events for all known SystemRollout.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("SystemRollout ResourceVersions are the same")
		return
	}

	glog.V(4).Infof("Updating SystemTeardown %s", oldTeardown.Name)
	c.enqueueSystemTeardown(curTeardown)
}

func (c *Controller) handleSystemAdd(obj interface{}) {
	<-c.owningLifecycleActionsSynced

	system := obj.(*crv1.System)
	glog.V(4).Infof("System %s added", system.Name)

	action, exists := c.getOwningAction(system.Namespace)
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
		c.enqueueSystemRollout(action.rollout)
		return
	}

	if action.teardown != nil {
		c.enqueueSystemTeardown(action.teardown)
		return
	}

	// FIXME: Send warn event
}

func (c *Controller) handleSystemUpdate(old, cur interface{}) {
	<-c.owningLifecycleActionsSynced

	glog.V(4).Info("Got System update")
	oldSystem := old.(*crv1.System)
	curSystem := cur.(*crv1.System)
	if oldSystem.ResourceVersion == curSystem.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("System ResourceVersions are the same")
		return
	}

	action, exists := c.getOwningAction(curSystem.Namespace)
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
		c.enqueueSystemRollout(action.rollout)
		return
	}

	if action.teardown != nil {
		c.enqueueSystemTeardown(action.teardown)
		return
	}

	// FIXME: Send warn event
}

func (c *Controller) handleSystemBuildAdd(obj interface{}) {
	<-c.owningLifecycleActionsSynced

	build := obj.(*crv1.SystemBuild)
	glog.V(4).Infof("SystemBuild %s added", build.Name)

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
		c.enqueueSystemRollout(action.rollout)
		return
	}

	// only need to update rollouts on builds finishing
}

func (c *Controller) handleSystemBuildUpdate(old, cur interface{}) {
	<-c.owningLifecycleActionsSynced

	glog.V(4).Infof("Got SystemBuild update")
	oldBuild := old.(*crv1.SystemBuild)
	curBuild := cur.(*crv1.SystemBuild)
	if oldBuild.ResourceVersion == curBuild.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("SystemBuild ResourceVersions are the same")
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
		c.enqueueSystemRollout(action.rollout)
		return
	}

	// only need to update rollouts on builds finishing
}

func (c *Controller) enqueueSystemRollout(sysRollout *crv1.SystemRollout) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sysRollout)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", sysRollout, err))
		return
	}

	c.rolloutQueue.Add(key)
}

func (c *Controller) enqueueSystemTeardown(syst *crv1.SystemTeardown) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(syst)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", syst, err))
		return
	}

	c.teardownQueue.Add(key)
}

func (c *Controller) syncOwningActions() error {
	c.owningLifecycleActionsLock.Lock()
	defer c.owningLifecycleActionsLock.Unlock()

	rollouts, err := c.systemRolloutLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, rollout := range rollouts {
		if rollout.Status.State != crv1.SystemRolloutStateInProgress {
			continue
		}

		_, exists := c.owningLifecycleActions[rollout.Namespace]
		if exists {
			return fmt.Errorf("System %v has multiple owning actions", rollout.Namespace)
		}

		c.owningLifecycleActions[rollout.Namespace] = &lifecycleAction{
			rollout: rollout,
		}
	}

	teardowns, err := c.systemTeardownLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, teardown := range teardowns {
		if teardown.Status.State != crv1.SystemTeardownStateInProgress {
			continue
		}

		_, exists := c.owningLifecycleActions[teardown.Namespace]
		if exists {
			return fmt.Errorf("System %v has multiple owning actions", teardown.Namespace)
		}

		c.owningLifecycleActions[teardown.Namespace] = &lifecycleAction{
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
	for c.processNextWorkItem(c.rolloutQueue, c.syncSystemRollout) {
	}
}

func (c *Controller) runTeardownWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for c.processNextWorkItem(c.teardownQueue, c.syncSystemTeardown) {
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

// syncSystemBuild will sync the SystemBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncSystemRollout(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing SystemRollout %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing SystemRollout %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	rollout, err := c.systemRolloutLister.SystemRollouts(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("SystemRollout %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	glog.V(5).Infof("SystemRollout %v state: %v", key, rollout.Status.State)

	switch rollout.Status.State {
	case crv1.SystemRolloutStateSucceeded, crv1.SystemRolloutStateFailed:
		glog.V(4).Infof("SystemRollout %s already completed", key)
		return nil

	case crv1.SystemRolloutStateInProgress:
		return c.syncInProgressRollout(rollout)

	case crv1.SystemRolloutStateAccepted:
		return c.syncAcceptedRollout(rollout)

	case crv1.SystemRolloutStatePending:
		return c.syncPendingRollout(rollout)

	default:
		return fmt.Errorf("SystemRollout %v has unexpected state: %v", key, rollout.Status.State)
	}
}

// syncSystemBuild will sync the SystemBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncSystemTeardown(key string) error {
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

	teardown, err := c.systemTeardownLister.SystemTeardowns(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("SystemTeardown %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	switch teardown.Status.State {
	case crv1.SystemTeardownStateSucceeded, crv1.SystemTeardownStateFailed:
		glog.V(4).Infof("SystemTeardown %s already completed", key)
		return nil

	case crv1.SystemTeardownStateInProgress:
		return c.syncInProgressTeardown(teardown)

	case crv1.SystemTeardownStatePending:
		return c.syncPendingTeardown(teardown)

	default:
		return fmt.Errorf("SystemTeardown %v has unexpected state: %v", key, teardown.Status.State)
	}
}
