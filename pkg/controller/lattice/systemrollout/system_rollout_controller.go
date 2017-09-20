package systemrollout

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	coretypes "github.com/mlab-lattice/core/pkg/types"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type SystemRolloutController struct {
	syncHandler          func(sysRolloutKey string) error
	enqueueSystemRollout func(sysRollout *crv1.SystemRollout)

	latticeResourceClient rest.Interface

	// Each LatticeNamespace may only have one rollout going on at a time.
	// A rollout is an "owning" rollout while it is rolling out, and until
	// it has completed and a new rollout has been accepted and becomes the
	// owning rollout. (e.g. we create SystemRollout A. It is accepted and becomes
	// the owning rollout. It then runs to completion. It is still the owning
	// rollout. Then SystemRollout B is created. It is accepted because the existing
	// owning rollout is not running. Now B is the owning rollout)
	owningRolloutsLock sync.RWMutex
	owningRollouts     map[coretypes.LatticeNamespace]*crv1.SystemRollout

	systemRolloutStore       cache.Store
	systemRolloutStoreSynced cache.InformerSynced

	systemStore       cache.Store
	systemStoreSynced cache.InformerSynced

	systemBuildStore       cache.Store
	systemBuildStoreSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewSystemRolloutController(
	latticeResourceClient rest.Interface,
	systemRolloutInformer cache.SharedInformer,
	systemInformer cache.SharedInformer,
	systemBuildInformer cache.SharedInformer,
) *SystemRolloutController {
	src := &SystemRolloutController{
		latticeResourceClient: latticeResourceClient,
		owningRollouts:        make(map[coretypes.LatticeNamespace]*crv1.SystemRollout),
		queue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system-rollout"),
	}

	src.enqueueSystemRollout = src.enqueue
	src.syncHandler = src.syncSystemRollout

	systemRolloutInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.addSystemRollout,
		UpdateFunc: src.updateSystemRollout,
		// TODO: for now it is assumed that SystemRollouts are not deleted. Revisit this.
	})
	src.systemRolloutStore = systemRolloutInformer.GetStore()
	src.systemRolloutStoreSynced = systemRolloutInformer.HasSynced

	systemInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.addSystem,
		UpdateFunc: src.updateSystem,
		// TODO: for now it is assumed that Systems are not deleted. Revisit this.
	})
	src.systemStore = systemInformer.GetStore()
	src.systemStoreSynced = systemInformer.HasSynced

	systemBuildInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    src.addSystemBuild,
		UpdateFunc: src.updateSystemBuild,
		// TODO: for now it is assumed that SystemBuilds are not deleted. Revisit this.
	})
	src.systemBuildStore = systemBuildInformer.GetStore()
	src.systemBuildStoreSynced = systemBuildInformer.HasSynced

	return src
}

func (src *SystemRolloutController) addSystemRollout(obj interface{}) {
	sysr := obj.(*crv1.SystemRollout)
	glog.V(4).Infof("Adding SystemRollout %s", sysr.Name)
	src.enqueueSystemRollout(sysr)
}

func (src *SystemRolloutController) updateSystemRollout(old, cur interface{}) {
	oldSysr := old.(*crv1.SystemRollout)
	curSysr := cur.(*crv1.SystemRollout)
	glog.V(4).Infof("Updating SystemRollout %s", oldSysr.Name)
	src.enqueueSystemRollout(curSysr)
}

func (src *SystemRolloutController) addSystem(obj interface{}) {
	sys := obj.(*crv1.System)
	glog.V(4).Infof("System %s added", sys.Name)

	src.owningRolloutsLock.RLock()
	defer src.owningRolloutsLock.RUnlock()
	owningRollout, ok := src.owningRollouts[sys.Spec.LatticeNamespace]
	if !ok {
		// TODO: send warn event
		return
	}

	src.enqueueSystemRollout(owningRollout)
}

func (src *SystemRolloutController) updateSystem(old, cur interface{}) {
	glog.V(4).Info("Got System update")
	oldSys := old.(*crv1.System)
	curSys := cur.(*crv1.System)
	if oldSys.ResourceVersion == curSys.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("System ResourceVersions are the same")
		return
	}

	src.owningRolloutsLock.RLock()
	defer src.owningRolloutsLock.RUnlock()
	owningRollout, ok := src.owningRollouts[curSys.Spec.LatticeNamespace]
	if !ok {
		// TODO: send warn event
		return
	}

	src.enqueueSystemRollout(owningRollout)
}

func (src *SystemRolloutController) addSystemBuild(obj interface{}) {
	sysb := obj.(*crv1.SystemBuild)
	glog.V(4).Infof("SystemBuild %s added", sysb.Name)

	src.owningRolloutsLock.RLock()
	defer src.owningRolloutsLock.RUnlock()
	owningRollout, ok := src.owningRollouts[sysb.Spec.LatticeNamespace]
	if !ok {
		// TODO: send warn event
		return
	}

	src.enqueueSystemRollout(owningRollout)
}

func (src *SystemRolloutController) updateSystemBuild(old, cur interface{}) {
	glog.V(4).Infof("Got SystemBuild update")
	oldSysb := old.(*crv1.SystemBuild)
	curSysb := cur.(*crv1.SystemBuild)
	if oldSysb.ResourceVersion == curSysb.ResourceVersion {
		// Periodic resync will send update events for all known ServiceBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("SystemBuild ResourceVersions are the same")
		return
	}

	src.owningRolloutsLock.RLock()
	defer src.owningRolloutsLock.RUnlock()
	owningRollout, ok := src.owningRollouts[curSysb.Spec.LatticeNamespace]
	if !ok {
		// TODO: send warn event
		return
	}

	src.enqueueSystemRollout(owningRollout)
}

func (src *SystemRolloutController) enqueue(sysRollout *crv1.SystemRollout) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sysRollout)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", sysRollout, err))
		return
	}

	src.queue.Add(key)
}

func (src *SystemRolloutController) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer src.queue.ShutDown()

	glog.Infof("Starting system-rollout controller")
	defer glog.Infof("Shutting down system-rollout controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, src.systemRolloutStoreSynced, src.systemStoreSynced, src.systemBuildStoreSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Syncing owning SystemRollouts")

	// It's okay that we're racing with the System and SystemBuild informer add/update functions here.
	// addSystemRollout and updateSystemRollout will enqueue all of the existing SystemRollouts already
	// so it's okay if the other informers don't.
	if err := src.syncOwningRollouts(); err != nil {
		return
	}

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(src.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (src *SystemRolloutController) syncOwningRollouts() error {
	src.owningRolloutsLock.Lock()
	defer src.owningRolloutsLock.Unlock()

	for _, sysrObj := range src.systemRolloutStore.List() {
		sysr := sysrObj.(*crv1.SystemRollout)
		if sysr.Status.State != crv1.SystemRolloutStateInProgress {
			continue
		}

		lns := sysr.Spec.LatticeNamespace
		if _, ok := src.owningRollouts[sysr.Spec.LatticeNamespace]; ok {
			return fmt.Errorf("LatticeNamespace %v has multiple owning rollouts", lns)
		}

		src.owningRollouts[lns] = sysr
	}

	return nil
}

func (src *SystemRolloutController) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for src.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (src *SystemRolloutController) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := src.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer src.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := src.syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		src.queue.Forget(key)
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
	src.queue.AddRateLimited(key)

	return true
}

// syncSystemBuild will sync the SystemBuild with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (src *SystemRolloutController) syncSystemRollout(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing SystemRollout %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing SystemRollout %q (%v)", key, time.Now().Sub(startTime))
	}()

	sysrObj, exists, err := src.systemRolloutStore.GetByKey(key)
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
		return src.syncInProgressRollout(sysr)

	case crv1.SystemRolloutStateAccepted:
		return src.syncAcceptedRollout(sysr)

	case crv1.SystemRolloutStatePending:
		return src.syncPendingRolloutState(sysr)

	default:
		panic("unreachable")
	}
}

func (src *SystemRolloutController) updateStatus(sysr *crv1.SystemRollout, newStatus crv1.SystemRolloutStatus) (*crv1.SystemRollout, error) {
	if reflect.DeepEqual(sysr.Status, newStatus) {
		return sysr, nil
	}

	sysr.Status = newStatus

	result := &crv1.SystemRollout{}
	err := src.latticeResourceClient.Put().
		Namespace(sysr.Namespace).
		Resource(crv1.SystemRolloutResourcePlural).
		Name(sysr.Name).
		Body(sysr).
		Do().
		Into(result)

	return result, err
}
