package service

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	latticeclientset "github.com/mlab-lattice/system/pkg/kubernetes/customresource/client"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	appinformers "k8s.io/client-go/informers/apps/v1beta2"
	coreinformers "k8s.io/client-go/informers/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta2"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("Service")

// We'll use LatticeService to differentiate between kubernetes' Service
type Controller struct {
	syncHandler    func(bKey string) error
	enqueueService func(cb *crv1.Service)

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	configStore       cache.Store
	configStoreSynced cache.InformerSynced
	configSetChan     chan struct{}
	configSet         bool
	configLock        sync.RWMutex
	config            crv1.ConfigSpec

	// FIXME: remove when local DNS server working
	systemStore       cache.Store
	systemStoreSynced cache.InformerSynced

	serviceStore       cache.Store
	serviceStoreSynced cache.InformerSynced

	deploymentLister       appslisters.DeploymentLister
	deploymentListerSynced cache.InformerSynced

	kubeServiceLister       corelisters.ServiceLister
	kubeServiceListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	configInformer cache.SharedInformer,
	systemInformer cache.SharedInformer,
	serviceInformer cache.SharedInformer,
	deploymentInformer appinformers.DeploymentInformer,
	kubeServiceInformer coreinformers.ServiceInformer,
) *Controller {
	sc := &Controller{
		kubeClient:    kubeClient,
		latticeClient: latticeClient,
		configSetChan: make(chan struct{}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service"),
	}

	sc.syncHandler = sc.syncService
	sc.enqueueService = sc.enqueue

	configInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    sc.handleConfigAdd,
		UpdateFunc: sc.handleConfigUpdate,
		// TODO: for now it is assumed that ComponentBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ComponentBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document CB GC ideas (need to write down last used date, lock properly, etc)
	})
	sc.configStore = configInformer.GetStore()
	sc.configStoreSynced = configInformer.HasSynced

	// FIXME: remove when local DNS server working
	systemInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleSystemAdd,
		UpdateFunc: sc.handleSystemUpdate,
	})
	sc.systemStore = systemInformer.GetStore()
	sc.systemStoreSynced = systemInformer.HasSynced

	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleServiceAdd,
		UpdateFunc: sc.handleServiceUpdate,
		DeleteFunc: sc.handleServiceDelete,
	})
	sc.serviceStore = serviceInformer.GetStore()
	sc.serviceStoreSynced = serviceInformer.HasSynced

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleDeploymentAdd,
		UpdateFunc: sc.handleDeploymentUpdate,
		DeleteFunc: sc.handleDeploymentDelete,
	})
	sc.deploymentLister = deploymentInformer.Lister()
	sc.deploymentListerSynced = deploymentInformer.Informer().HasSynced

	sc.kubeServiceLister = kubeServiceInformer.Lister()
	sc.kubeServiceListerSynced = kubeServiceInformer.Informer().HasSynced

	return sc
}

func (sc *Controller) handleConfigAdd(obj interface{}) {
	config := obj.(*crv1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	sc.configLock.Lock()
	defer sc.configLock.Unlock()
	sc.config = config.DeepCopy().Spec

	if !sc.configSet {
		sc.configSet = true
		close(sc.configSetChan)
	}
}

func (sc *Controller) handleConfigUpdate(old, cur interface{}) {
	oldConfig := old.(*crv1.Config)
	curConfig := cur.(*crv1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	sc.configLock.Lock()
	defer sc.configLock.Unlock()
	sc.config = curConfig.DeepCopy().Spec
}

func (sc *Controller) handleSystemAdd(obj interface{}) {
	sys := obj.(*crv1.System)
	glog.V(4).Infof("Adding System %s", sys.Name)

	for _, svcInfo := range sys.Spec.Services {
		if svcInfo.ServiceName == nil {
			// FIXME: what to do here?
			// probably okay not to worry about, this should be temp until local DNS working
			continue
		}

		svcKey := fmt.Sprintf("%v/%v", sys.Namespace, svcInfo.ServiceName)
		svcObj, exists, err := sc.serviceStore.GetByKey(svcKey)
		if err != nil || !exists {
			// FIXME: what to do here?
			// probably okay not to worry about, this should be temp until local DNS working
			continue
		}

		svc := svcObj.(*crv1.Service)
		sc.enqueueService(svc)
	}
}

func (sc *Controller) handleSystemUpdate(old, cur interface{}) {
	oldSys := old.(*crv1.System)
	curSys := cur.(*crv1.System)
	if oldSys.ResourceVersion == curSys.ResourceVersion {
		// Periodic resync will send update events for all known Deployments.
		// Two different versions of the same Deployment will always have different RVs.
		glog.V(5).Info("System ResourceVersions are the same")
		return
	}

	glog.V(4).Infof("Updating System %s", oldSys.Name)
	for _, svcInfo := range curSys.Spec.Services {
		if svcInfo.ServiceName == nil {
			// FIXME: what to do here?
			// probably okay not to worry about, this should be temp until local DNS working
			continue
		}

		svcKey := fmt.Sprintf("%v/%v", curSys.Namespace, svcInfo.ServiceName)
		svcObj, exists, err := sc.serviceStore.GetByKey(svcKey)
		if err != nil || !exists {
			// FIXME: what to do here?
			// probably okay not to worry about, this should be temp until local DNS working
			continue
		}

		svc := svcObj.(*crv1.Service)
		sc.enqueueService(svc)
	}
}

func (sc *Controller) handleServiceAdd(obj interface{}) {
	svc := obj.(*crv1.Service)
	glog.V(4).Infof("Adding Service %s", svc.Name)
	sc.enqueueService(svc)
}

func (sc *Controller) handleServiceUpdate(old, cur interface{}) {
	oldSvc := old.(*crv1.Service)
	curSvc := cur.(*crv1.Service)
	glog.V(4).Infof("Updating Service %s", oldSvc.Name)
	sc.enqueueService(curSvc)
}

func (sc *Controller) handleServiceDelete(obj interface{}) {
	svc, ok := obj.(*crv1.Service)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		svc, ok = tombstone.Obj.(*crv1.Service)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}
	glog.V(4).Infof("Deleting Service %s", svc.Name)
	sc.enqueueService(svc)
}

// handleDeploymentAdd enqueues the Service that manages a Deployment when the Deployment is created.
func (sc *Controller) handleDeploymentAdd(obj interface{}) {
	d := obj.(*appsv1beta2.Deployment)

	if d.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		sc.handleDeploymentDelete(d)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(d); controllerRef != nil {
		svc := sc.resolveControllerRef(d.Namespace, controllerRef)

		// Not a Service Deployment.
		if svc == nil {
			return
		}

		glog.V(4).Infof("Deployment %s added.", d.Name)
		sc.enqueueService(svc)
		return
	}

	// Otherwise, it's an orphan. These should never exist. All deployments should be run by some
	// controller.
	// FIXME: send warn event
}

// handleDeploymentUpdate figures out what Service manages a Deployment when the Deployment
// is updated and enqueues it.
func (sc *Controller) handleDeploymentUpdate(old, cur interface{}) {
	glog.V(5).Info("Got Deployment update")
	oldD := old.(*appsv1beta2.Deployment)
	curD := cur.(*appsv1beta2.Deployment)
	if curD.ResourceVersion == oldD.ResourceVersion {
		// Periodic resync will send update events for all known Deployments.
		// Two different versions of the same Deployment will always have different RVs.
		glog.V(5).Info("Deployment ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curD)
	oldControllerRef := metav1.GetControllerOf(oldD)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// The ControllerRef was changed. If this is a Service Deployment, this shouldn't happen.
		if b := sc.resolveControllerRef(oldD.Namespace, oldControllerRef); b != nil {
			// FIXME: send error event here, this should not happen
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		svc := sc.resolveControllerRef(curD.Namespace, curControllerRef)

		// Not a Service Deployment
		if svc == nil {
			return
		}

		sc.enqueueService(svc)
		return
	}

	// Otherwise, it's an orphan. These should never exist. All deployments should be run by some
	// controller.
	// FIXME: send warn event
}

// handleDeploymentDelete enqueues the Service that manages a Deployment when
// the Deployment is deleted.
func (sc *Controller) handleDeploymentDelete(obj interface{}) {
	d, ok := obj.(*appsv1beta2.Deployment)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*appsv1beta2.Deployment)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Deployment %#v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(d)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	svc := sc.resolveControllerRef(d.Namespace, controllerRef)

	// Not a Service Deployment
	if svc == nil {
		return
	}

	glog.V(4).Infof("Deployment %s deleted.", d.Name)
	sc.enqueueService(svc)
}

func (sc *Controller) enqueue(svc *crv1.Service) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(svc)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", svc, err))
		return
	}

	sc.queue.Add(key)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (sc *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.Service {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}

	svcKey := fmt.Sprintf("%v/%v", namespace, controllerRef.Name)
	svcObj, exists, err := sc.serviceStore.GetByKey(svcKey)
	if err != nil || !exists {
		return nil
	}

	svc := svcObj.(*crv1.Service)

	if svc.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return svc
}

func (sc *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer sc.queue.ShutDown()

	glog.Infof("Starting service controller")
	defer glog.Infof("Shutting down service controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, sc.serviceStoreSynced, sc.deploymentListerSynced, sc.kubeServiceListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Waiting for config to be set")

	// wait for config to be set
	<-sc.configSetChan

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(sc.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (sc *Controller) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for sc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (sc *Controller) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := sc.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer sc.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := sc.syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		sc.queue.Forget(key)
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
	sc.queue.AddRateLimited(key)

	return true
}

// syncService will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (sc *Controller) syncService(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing Service %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing Service %q (%v)", key, time.Now().Sub(startTime))
	}()

	svcObj, exists, err := sc.serviceStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("Service %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	svc := svcObj.(*crv1.Service)

	// TODO: probably need to change this when adding Blue/Green rollouts or canaries, there will probably be
	// 		 multiple deployments per Service.
	d, err := sc.getDeploymentForService(svc)
	if err != nil {
		return err
	}

	if d == nil {
		glog.V(4).Infof("Did not find Deployment for Service %q, creating one", svc.Name)
		dResp, err := sc.createDeployment(svc)
		if err != nil {
			return err
		}
		d = dResp
	} else {
		glog.V(4).Infof("Found Deployment for Service %q, syncing its Spec", svc.Name)
		d, err = sc.syncDeploymentSpec(svc, d)
		if err != nil {
			return err
		}
	}

	ksvc, err := sc.getKubeServiceForService(svc)
	if err != nil {
		return err
	}

	// FIXME: may have to update kubeService if ports change?
	if ksvc == nil {
		glog.V(4).Infof("Did not find kubeService for Service %q, creating one", svc.Name)
		if _, err := sc.createKubeService(svc); err != nil {
			return err
		}
	}

	svcCopy := svc.DeepCopy()
	return sc.syncServiceWithDeployment(svcCopy, d)
}

func (sc *Controller) syncServiceWithDeployment(svc *crv1.Service, d *appsv1beta2.Deployment) error {
	newStatus := calculateServiceStatus(d)
	return sc.updateServiceStatus(svc, newStatus)
}

// TODO: this is overly simplistic
func calculateServiceStatus(d *appsv1beta2.Deployment) crv1.ServiceStatus {
	available := false
	//progressing := false
	failure := false

	for _, condition := range d.Status.Conditions {
		switch condition.Type {
		case appsv1beta2.DeploymentAvailable:
			if condition.Status == corev1.ConditionTrue {
				available = true
			}
		//case appsv1beta2.DeploymentProgressing:
		//	if condition.Status == corev1.ConditionTrue {
		//		progressing = true
		//	}
		case appsv1beta2.DeploymentReplicaFailure:
			if condition.Status == corev1.ConditionTrue {
				failure = true
			}
		}
	}

	if failure {
		return crv1.ServiceStatus{
			State: crv1.ServiceStateRolloutFailed,
		}
	}

	if available {
		return crv1.ServiceStatus{
			State: crv1.ServiceStateRolloutSucceeded,
		}
	}

	return crv1.ServiceStatus{
		State: crv1.ServiceStateRollingOut,
	}
}

func (sc *Controller) updateServiceStatus(svc *crv1.Service, newStatus crv1.ServiceStatus) error {
	if reflect.DeepEqual(svc.Status, newStatus) {
		return nil
	}

	svc.Status = newStatus
	_, err := sc.latticeClient.V1().Services(svc.Namespace).Update(svc)
	return err
}
