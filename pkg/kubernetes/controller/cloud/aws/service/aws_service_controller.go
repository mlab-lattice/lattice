package service

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/listers/lattice/v1"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	coreinformers "k8s.io/client-go/informers/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("Service")

type Controller struct {
	syncHandler    func(bKey string) error
	enqueueService func(cb *crv1.Service)

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	configLister       latticelisters.ConfigLister
	configListerSynced cache.InformerSynced
	configSetChan      chan struct{}
	configSet          bool
	configLock         sync.RWMutex
	config             crv1.ConfigSpec

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	kubeServiceLister       corelisters.ServiceLister
	kubeServiceListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	terraformModulePath string
}

func NewController(
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	configInformer latticeinformers.ConfigInformer,
	serviceInformer latticeinformers.ServiceInformer,
	kubeServiceInformer coreinformers.ServiceInformer,
	terraformModulePath string,
) *Controller {
	sc := &Controller{
		kubeClient:          kubeClient,
		latticeClient:       latticeClient,
		configSetChan:       make(chan struct{}),
		queue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service"),
		terraformModulePath: terraformModulePath,
	}

	sc.syncHandler = sc.syncService
	sc.enqueueService = sc.enqueue

	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    sc.handleConfigAdd,
		UpdateFunc: sc.handleConfigUpdate,
		// TODO: for now it is assumed that ComponentBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ComponentBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document CB GC ideas (need to write down last used date, lock properly, etc)
	})
	sc.configLister = configInformer.Lister()
	sc.configListerSynced = configInformer.Informer().HasSynced

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleServiceAdd,
		UpdateFunc: sc.handleServiceUpdate,
		DeleteFunc: sc.handleServiceDelete,
	})
	sc.serviceLister = serviceInformer.Lister()
	sc.serviceListerSynced = serviceInformer.Informer().HasSynced

	kubeServiceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleKubeServiceAdd,
		UpdateFunc: sc.handleKubeServiceUpdate,
		DeleteFunc: sc.handleKubeServiceDelete,
	})
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

func (sc *Controller) enqueue(svc *crv1.Service) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(svc)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", svc, err))
		return
	}

	sc.queue.Add(key)
}

// handleKubeServiceAdd enqueues the Service that manages a kubeService when the kubeService is created.
func (sc *Controller) handleKubeServiceAdd(obj interface{}) {
	kubeSvc := obj.(*corev1.Service)

	if kubeSvc.DeletionTimestamp != nil {
		// On a restart of the controller manager, it'kubeSvc possible for an object to
		// show up in a state that is already pending deletion.
		sc.handleKubeServiceDelete(kubeSvc)
		return
	}

	// If it has a ControllerRef, that'kubeSvc all that matters.
	if controllerRef := metav1.GetControllerOf(kubeSvc); controllerRef != nil {
		svc := sc.resolveControllerRef(kubeSvc.Namespace, controllerRef)

		// Not a Service kubeService.
		if svc == nil {
			return
		}

		glog.V(4).Infof("kubeService %v added.", kubeSvc.Name)
		sc.enqueueService(svc)
		return
	}

	// Otherwise, it's an orphan, and therefore not of interest to us here.
}

// handleKubeServiceUpdate figures out what Service manages a kubeService when the kubeService
// is updated and enqueues it.
func (sc *Controller) handleKubeServiceUpdate(old, cur interface{}) {
	glog.V(5).Info("Got kubeService update")
	oldKSvc := old.(*corev1.Service)
	curKSvc := cur.(*corev1.Service)
	if curKSvc.ResourceVersion == oldKSvc.ResourceVersion {
		// Periodic resync will send update events for all known Deployments.
		// Two different versions of the same Deployment will always have different RVs.
		glog.V(5).Info("kubeService ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curKSvc)
	oldControllerRef := metav1.GetControllerOf(oldKSvc)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// The ControllerRef was changed. If this is a Service Deployment, this shouldn't happen.
		if b := sc.resolveControllerRef(oldKSvc.Namespace, oldControllerRef); b != nil {
			// FIXME: send error event here, this should not happen
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		svc := sc.resolveControllerRef(curKSvc.Namespace, curControllerRef)

		// Not a Service kubeService.
		if svc == nil {
			return
		}

		sc.enqueueService(svc)
		return
	}

	// Otherwise, it's an orphan, and therefore not of interest to us here.
}

// handleDeploymentDelete enqueues the Service that manages a Deployment when
// the Deployment is deleted.
func (sc *Controller) handleKubeServiceDelete(obj interface{}) {
	kubeSvc, ok := obj.(*corev1.Service)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		kubeSvc, ok = tombstone.Obj.(*corev1.Service)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a kubeService %#v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(kubeSvc)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	svc := sc.resolveControllerRef(kubeSvc.Namespace, controllerRef)

	// Not a Service kubeService
	if svc == nil {
		return
	}

	glog.V(4).Infof("kubeService %s deleted.", kubeSvc.Name)
	sc.enqueueService(svc)
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

	svc, err := sc.serviceLister.Services(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

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
	if !cache.WaitForCacheSync(stopCh, sc.serviceListerSynced, sc.kubeServiceListerSynced) {
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

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	svc, err := sc.serviceLister.Services(namespace).Get(name)
	if errors.IsNotFound(err) {
		//svcObj, exists, err := sc.serviceLister.GetByKey(key)
		glog.V(2).Infof("Service %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	svcCopy := svc.DeepCopy()

	// Before we start doing any work, we need to add our finalizer to the Service so that we can
	// clean up anything we have created when the Service gets deleted.
	err = sc.addFinalizer(svcCopy)
	if err != nil {
		return err
	}

	// Next, we need to find ensure that the kubeServices for this Service have been created.
	kubeSvc, necessary, err := sc.getKubeServiceForService(svcCopy)
	if err != nil {
		return err
	}

	// If this service has been deleted, we should deprovision the resources and remove the finalizer.
	if svcCopy.DeletionTimestamp != nil {
		return sc.deprovisionService(svcCopy)
	}

	// If this Service requires a kubeService be created and it has not yet been, we'll just say
	// we're done working on it for now.
	// When the kubeService gets created, the kubeService Informer handlers will requeue this
	// service.
	if necessary && kubeSvc == nil {
		return nil
	}

	// Otherwise we either don't need a kubeService or it's already created, so we're good to go.
	err = sc.provisionService(svcCopy)
	if err != nil {
		return err
	}

	return nil
}
