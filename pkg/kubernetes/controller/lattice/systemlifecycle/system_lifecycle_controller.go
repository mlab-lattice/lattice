package systemlifecycle

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	latticeclientset "github.com/mlab-lattice/system/pkg/kubernetes/customresource/client"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
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
	owningLifecycleActionsLock sync.RWMutex
	owningLifecycleActions     map[types.LatticeNamespace]*lifecycleAction

	systemRolloutStore       cache.Store
	systemRolloutStoreSynced cache.InformerSynced

	systemTeardownStore       cache.Store
	systemTeardownStoreSynced cache.InformerSynced

	systemStore       cache.Store
	systemStoreSynced cache.InformerSynced

	systemBuildStore       cache.Store
	systemBuildStoreSynced cache.InformerSynced

	serviceBuildStore       cache.Store
	serviceBuildStoreSynced cache.InformerSynced

	componentBuildStore       cache.Store
	componentBuildStoreSynced cache.InformerSynced

	rolloutQueue  workqueue.RateLimitingInterface
	teardownQueue workqueue.RateLimitingInterface
}

func NewController(
	latticeClient latticeclientset.Interface,
	systemRolloutInformer cache.SharedInformer,
	systemTeardownInformer cache.SharedInformer,
	systemInformer cache.SharedInformer,
	systemBuildInformer cache.SharedInformer,
	serviceBuildInformer cache.SharedInformer,
	componentBuildInformer cache.SharedInformer,
) *Controller {
	src := &Controller{
		latticeClient:          latticeClient,
		owningLifecycleActions: make(map[types.LatticeNamespace]*lifecycleAction),
		rolloutQueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-rollout"),
		teardownQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-teardown"),
	}

	src.syncHandler = src.syncSystemRollout

	systemRolloutInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleSystemRolloutAdd,
		UpdateFunc: src.handleSystemRolloutUpdate,
		// TODO: for now it is assumed that SystemRollouts are not deleted. Revisit this.
	})
	src.systemRolloutStore = systemRolloutInformer.GetStore()
	src.systemRolloutStoreSynced = systemRolloutInformer.HasSynced

	systemTeardownInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleSystemTeardownAdd,
		UpdateFunc: src.handleSystemTeardownUpdate,
		// TODO: for now it is assumed that SystemRollouts are not deleted. Revisit this.
	})
	src.systemTeardownStore = systemTeardownInformer.GetStore()
	src.systemTeardownStoreSynced = systemTeardownInformer.HasSynced

	systemInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleSystemAdd,
		UpdateFunc: src.handleSystemUpdate,
		// TODO: for now it is assumed that Systems are not deleted. Revisit this.
	})
	src.systemStore = systemInformer.GetStore()
	src.systemStoreSynced = systemInformer.HasSynced

	systemBuildInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.handleSystemBuildAdd,
		UpdateFunc: src.handleSystemBuildUpdate,
		// TODO: for now it is assumed that SystemBuilds are not deleted. Revisit this.
	})
	src.systemBuildStore = systemBuildInformer.GetStore()
	src.systemBuildStoreSynced = systemBuildInformer.HasSynced

	src.serviceBuildStore = serviceBuildInformer.GetStore()
	src.serviceBuildStoreSynced = serviceBuildInformer.HasSynced

	src.componentBuildStore = componentBuildInformer.GetStore()
	src.componentBuildStoreSynced = componentBuildInformer.HasSynced

	return src
}

func (slc *Controller) handleSystemRolloutAdd(obj interface{}) {
	sysr := obj.(*crv1.SystemRollout)
	glog.V(4).Infof("Adding SystemRollout %s", sysr.Name)
	slc.enqueueSystemRollout(sysr)
}

func (slc *Controller) handleSystemRolloutUpdate(old, cur interface{}) {
	oldSysr := old.(*crv1.SystemRollout)
	curSysr := cur.(*crv1.SystemRollout)
	glog.V(4).Infof("Updating SystemRollout %s", oldSysr.Name)
	slc.enqueueSystemRollout(curSysr)
}

func (slc *Controller) handleSystemTeardownAdd(obj interface{}) {
	syst := obj.(*crv1.SystemTeardown)
	glog.V(4).Infof("Adding SystemTeardown %s", syst.Name)
	slc.enqueueSystemTeardown(syst)
}

func (slc *Controller) handleSystemTeardownUpdate(old, cur interface{}) {
	oldSyst := old.(*crv1.SystemTeardown)
	curSyst := cur.(*crv1.SystemTeardown)
	glog.V(4).Infof("Updating SystemTeardown %s", oldSyst.Name)
	slc.enqueueSystemTeardown(curSyst)
}

func (slc *Controller) handleSystemAdd(obj interface{}) {
	sys := obj.(*crv1.System)
	glog.V(4).Infof("System %s added", sys.Name)

	slc.owningLifecycleActionsLock.RLock()
	defer slc.owningLifecycleActionsLock.RUnlock()
	owningAction, ok := slc.owningLifecycleActions[types.LatticeNamespace(sys.Namespace)]
	if !ok {
		// TODO: send warn event
		return
	}

	if owningAction.rollout != nil {
		slc.enqueueSystemRollout(owningAction.rollout)
	} else {
		// TODO send warn here, a system shouldnt be added while a teardown is in progress
	}
}

func (slc *Controller) handleSystemUpdate(old, cur interface{}) {
	glog.V(4).Info("Got System update")
	oldSys := old.(*crv1.System)
	curSys := cur.(*crv1.System)
	if oldSys.ResourceVersion == curSys.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("System ResourceVersions are the same")
		return
	}

	slc.owningLifecycleActionsLock.RLock()
	defer slc.owningLifecycleActionsLock.RUnlock()
	owningAction, ok := slc.owningLifecycleActions[types.LatticeNamespace(curSys.Namespace)]
	if !ok {
		// TODO: send warn event
		return
	}

	if owningAction.rollout != nil {
		slc.enqueueSystemRollout(owningAction.rollout)
	} else {
		slc.enqueueSystemTeardown(owningAction.teardown)
	}
}

func (slc *Controller) handleSystemBuildAdd(obj interface{}) {
	sysb := obj.(*crv1.SystemBuild)
	glog.V(4).Infof("SystemBuild %s added", sysb.Name)

	slc.owningLifecycleActionsLock.RLock()
	defer slc.owningLifecycleActionsLock.RUnlock()
	owningAction, ok := slc.owningLifecycleActions[sysb.Spec.LatticeNamespace]
	if !ok {
		// TODO: send warn event
		return
	}

	if owningAction.rollout != nil {
		slc.enqueueSystemRollout(owningAction.rollout)
	}
}

func (slc *Controller) handleSystemBuildUpdate(old, cur interface{}) {
	glog.V(4).Infof("Got SystemBuild update")
	oldSysb := old.(*crv1.SystemBuild)
	curSysb := cur.(*crv1.SystemBuild)
	if oldSysb.ResourceVersion == curSysb.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("SystemBuild ResourceVersions are the same")
		return
	}

	slc.owningLifecycleActionsLock.RLock()
	defer slc.owningLifecycleActionsLock.RUnlock()
	owningAction, ok := slc.owningLifecycleActions[curSysb.Spec.LatticeNamespace]
	if !ok {
		// TODO: send warn event
		return
	}

	if owningAction.rollout != nil {
		slc.enqueueSystemRollout(owningAction.rollout)
	}
}

func (slc *Controller) enqueueSystemRollout(sysRollout *crv1.SystemRollout) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sysRollout)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", sysRollout, err))
		return
	}

	slc.rolloutQueue.Add(key)
}

func (slc *Controller) enqueueSystemTeardown(syst *crv1.SystemTeardown) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(syst)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", syst, err))
		return
	}

	slc.teardownQueue.Add(key)
}

func (slc *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer slc.rolloutQueue.ShutDown()
	defer slc.teardownQueue.ShutDown()

	glog.Infof("Starting system-rollout controller")
	defer glog.Infof("Shutting down system-rollout controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(
		stopCh,
		slc.systemRolloutStoreSynced,
		slc.systemTeardownStoreSynced,
		slc.systemStoreSynced,
		slc.systemBuildStoreSynced,
		slc.serviceBuildStoreSynced,
		slc.componentBuildStoreSynced,
	) {
		return
	}

	glog.V(4).Info("Caches synced. Syncing owning SystemRollouts")

	// It's okay that we're racing with the System and SystemBuild informer add/update functions here.
	// handleSystemRolloutAdd and handleSystemRolloutUpdate will enqueue all of the existing SystemRollouts already
	// so it's okay if the other informers don't.
	if err := slc.syncOwningActions(); err != nil {
		return
	}

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(slc.runRolloutWorker, time.Second, stopCh)
	}

	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(slc.runTeardownWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (slc *Controller) syncOwningActions() error {
	slc.owningLifecycleActionsLock.Lock()
	defer slc.owningLifecycleActionsLock.Unlock()

	for _, sysrObj := range slc.systemRolloutStore.List() {
		sysr := sysrObj.(*crv1.SystemRollout)
		if sysr.Status.State != crv1.SystemRolloutStateInProgress {
			continue
		}

		lns := sysr.Spec.LatticeNamespace
		if _, ok := slc.owningLifecycleActions[lns]; ok {
			return fmt.Errorf("LatticeNamespace %v has multiple owning rollouts", lns)
		}

		slc.owningLifecycleActions[lns] = &lifecycleAction{
			rollout: sysr,
		}
	}

	for _, systObj := range slc.systemTeardownStore.List() {
		syst := systObj.(*crv1.SystemTeardown)
		if syst.Status.State != crv1.SystemTeardownStateInProgress {
			continue
		}

		lns := syst.Spec.LatticeNamespace
		if _, ok := slc.owningLifecycleActions[lns]; ok {
			return fmt.Errorf("LatticeNamespace %v has multiple owning actions", lns)
		}

		slc.owningLifecycleActions[lns] = &lifecycleAction{
			teardown: syst,
		}
	}

	return nil
}

func (slc *Controller) runRolloutWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for slc.processNextWorkItem(slc.rolloutQueue, slc.syncSystemRollout) {
	}
}

func (slc *Controller) runTeardownWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for slc.processNextWorkItem(slc.teardownQueue, slc.syncSystemTeardown) {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (slc *Controller) processNextWorkItem(queue workqueue.RateLimitingInterface, syncHandler func(string) error) bool {
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
func (slc *Controller) syncSystemRollout(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing SystemRollout %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing SystemRollout %q (%v)", key, time.Now().Sub(startTime))
	}()

	sysrObj, exists, err := slc.systemRolloutStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("SystemRollout %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	sysr := sysrObj.(*crv1.SystemRollout)

	switch sysr.Status.State {
	case crv1.SystemRolloutStateSucceeded, crv1.SystemRolloutStateFailed:
		glog.V(4).Infof("SystemRollout %s already completed", key)
		return nil

	case crv1.SystemRolloutStateInProgress:
		return slc.syncInProgressRollout(sysr)

	case crv1.SystemRolloutStateAccepted:
		return slc.syncAcceptedRollout(sysr)

	case crv1.SystemRolloutStatePending:
		return slc.syncPendingRolloutState(sysr)

	default:
		panic("unreachable")
	}
}

func (slc *Controller) updateSystemRolloutStatus(sysr *crv1.SystemRollout, newStatus crv1.SystemRolloutStatus) (*crv1.SystemRollout, error) {
	if reflect.DeepEqual(sysr.Status, newStatus) {
		return sysr, nil
	}

	sysr.Status = newStatus
	return slc.latticeClient.V1().SystemRollouts(sysr.Namespace).Update(sysr)
}

// syncSystemBuild will sync the SystemBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (slc *Controller) syncSystemTeardown(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing SystemTeardown %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing SystemTeardown %q (%v)", key, time.Now().Sub(startTime))
	}()

	systObj, exists, err := slc.systemTeardownStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("SystemTeardown %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	syst := systObj.(*crv1.SystemTeardown)

	switch syst.Status.State {
	case crv1.SystemTeardownStateSucceeded, crv1.SystemTeardownStateFailed:
		glog.V(4).Infof("SystemTeardown %s already completed", key)
		return nil

	case crv1.SystemTeardownStateInProgress:
		return slc.syncInProgressTeardown(syst)

	case crv1.SystemTeardownStatePending:
		return slc.syncPendingTeardown(syst)

	default:
		panic("unreachable")
	}
}

func (slc *Controller) updateSystemTeardownStatus(syst *crv1.SystemTeardown, newStatus crv1.SystemTeardownStatus) (*crv1.SystemTeardown, error) {
	if reflect.DeepEqual(syst.Status, newStatus) {
		return syst, nil
	}

	syst.Status = newStatus
	return slc.latticeClient.V1().SystemTeardowns(syst.Namespace).Update(syst)
}
