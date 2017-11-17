package service

import (
	"fmt"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("Service")

// We'll use LatticeService to differentiate between kubernetes' Service
type ServiceController struct {
	syncHandler    func(bKey string) error
	enqueueService func(cb *crv1.Service)

	latticeResourceRestClient rest.Interface
	kubeClient                clientset.Interface

	configStore       cache.Store
	configStoreSynced cache.InformerSynced
	configSetChan     chan struct{}
	configSet         bool
	configLock        sync.RWMutex
	config            crv1.ConfigSpec

	serviceStore       cache.Store
	serviceStoreSynced cache.InformerSynced

	kubeServiceLister       corelisters.ServiceLister
	kubeServiceListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewServiceController(
	kubeClient clientset.Interface,
	latticeResourceRestClient rest.Interface,
	configInformer cache.SharedInformer,
	serviceInformer cache.SharedInformer,
	kubeServiceInformer coreinformers.ServiceInformer,
) *ServiceController {
	sc := &ServiceController{
		latticeResourceRestClient: latticeResourceRestClient,
		kubeClient:                kubeClient,
		configSetChan:             make(chan struct{}),
		queue:                     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service"),
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

	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleServiceAdd,
		UpdateFunc: sc.handleServiceUpdate,
		DeleteFunc: sc.handleServiceDelete,
	})
	sc.serviceStore = serviceInformer.GetStore()
	sc.serviceStoreSynced = serviceInformer.HasSynced

	//kubeServiceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc:    sc.handleKubeServiceAdd,
	//	UpdateFunc: sc.handleKubeServiceUpdate,
	//	DeleteFunc: sc.handleKubeServiceDelete,
	//})
	sc.kubeServiceLister = kubeServiceInformer.Lister()
	sc.kubeServiceListerSynced = kubeServiceInformer.Informer().HasSynced

	return sc
}

func (sc *ServiceController) handleConfigAdd(obj interface{}) {
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

func (sc *ServiceController) handleConfigUpdate(old, cur interface{}) {
	oldConfig := old.(*crv1.Config)
	curConfig := cur.(*crv1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	sc.configLock.Lock()
	defer sc.configLock.Unlock()
	sc.config = curConfig.DeepCopy().Spec
}

func (sc *ServiceController) handleServiceAdd(obj interface{}) {
	svc := obj.(*crv1.Service)
	glog.V(4).Infof("Adding Service %s", svc.Name)
	sc.enqueueService(svc)
}

func (sc *ServiceController) handleServiceUpdate(old, cur interface{}) {
	oldSvc := old.(*crv1.Service)
	curSvc := cur.(*crv1.Service)
	glog.V(4).Infof("Updating Service %s", oldSvc.Name)
	sc.enqueueService(curSvc)
}

func (sc *ServiceController) handleServiceDelete(obj interface{}) {
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

func (sc *ServiceController) enqueue(svc *crv1.Service) {
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
func (sc *ServiceController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.Service {
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

func (sc *ServiceController) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer sc.queue.ShutDown()

	glog.Infof("Starting service controller")
	defer glog.Infof("Shutting down service controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, sc.serviceStoreSynced, sc.kubeServiceListerSynced) {
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

func (sc *ServiceController) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for sc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (sc *ServiceController) processNextWorkItem() bool {
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
func (sc *ServiceController) syncService(key string) error {
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
	svcCopy := svc.DeepCopy()

	// Before we start doing any work, we need to add our finalizer to the Service so that we can
	// clean up anything we have created when the Service gets deleted.
	err = sc.addFinalizer(svcCopy)
	if err != nil {
		return nil
	}

	// Next, we need to find ensure that the kubeServices for this Service have been created.
	kubeSvc, necessary, err := sc.getKubeServiceForService(svcCopy)

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
