package service

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/base/service/util"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	appinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("Service")

type Controller struct {
	syncHandler    func(bKey string) error
	enqueueService func(cb *crv1.Service)

	cloudProvider cloudprovider.Interface
	serviceMesh   servicemesh.Interface

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

	nodePoolLister       latticelisters.NodePoolLister
	nodePoolListerSynced cache.InformerSynced

	deploymentLister       appslisters.DeploymentLister
	deploymentListerSynced cache.InformerSynced

	kubeServiceLister       corelisters.ServiceLister
	kubeServiceListerSynced cache.InformerSynced

	serviceAddressLister       latticelisters.ServiceAddressLister
	serviceAddressListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	cloudProvider cloudprovider.Interface,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	configInformer latticeinformers.ConfigInformer,
	serviceInformer latticeinformers.ServiceInformer,
	nodePoolInformer latticeinformers.NodePoolInformer,
	deploymentInformer appinformers.DeploymentInformer,
	kubeServiceInformer coreinformers.ServiceInformer,
	serviceAddressInformer latticeinformers.ServiceAddressInformer,
) *Controller {
	sc := &Controller{
		cloudProvider: cloudProvider,

		kubeClient:    kubeClient,
		latticeClient: latticeClient,
		configSetChan: make(chan struct{}),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service"),
	}

	sc.syncHandler = sc.syncService
	sc.enqueueService = sc.enqueue

	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    sc.handleConfigAdd,
		UpdateFunc: sc.handleConfigUpdate,
		// TODO(kevinrosendahl): for now it is assumed that ComponentBuilds are not deleted.
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

	nodePoolInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleNodePoolAdd,
		UpdateFunc: sc.handleNodePoolUpdate,
	})
	sc.nodePoolLister = nodePoolInformer.Lister()
	sc.nodePoolListerSynced = nodePoolInformer.Informer().HasSynced

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleDeploymentAdd,
		UpdateFunc: sc.handleDeploymentUpdate,
		DeleteFunc: sc.handleDeploymentDelete,
	})
	sc.deploymentLister = deploymentInformer.Lister()
	sc.deploymentListerSynced = deploymentInformer.Informer().HasSynced

	kubeServiceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleKubeServiceAdd,
		UpdateFunc: sc.handleKubeServiceUpdate,
	})
	sc.kubeServiceLister = kubeServiceInformer.Lister()
	sc.kubeServiceListerSynced = kubeServiceInformer.Informer().HasSynced

	serviceAddressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleServiceAddressAdd,
		UpdateFunc: sc.handleServiceAddressUpdate,
	})
	sc.serviceAddressLister = serviceAddressInformer.Lister()
	sc.serviceAddressListerSynced = serviceAddressInformer.Informer().HasSynced

	return sc
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting service controller")
	defer glog.Infof("Shutting down service controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.serviceListerSynced, c.deploymentListerSynced, c.kubeServiceListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Waiting for config to be set")

	// wait for config to be set
	<-c.configSetChan

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

func (c *Controller) handleConfigAdd(obj interface{}) {
	config := obj.(*crv1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = config.DeepCopy().Spec

	serviceMesh, err := servicemesh.NewServiceMesh(&c.config.ServiceMesh)
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}

	c.serviceMesh = serviceMesh

	if !c.configSet {
		c.configSet = true
		close(c.configSetChan)
	}
}

func (c *Controller) handleConfigUpdate(old, cur interface{}) {
	oldConfig := old.(*crv1.Config)
	curConfig := cur.(*crv1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = curConfig.DeepCopy().Spec

	serviceMesh, err := servicemesh.NewServiceMesh(&c.config.ServiceMesh)
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}

	c.serviceMesh = serviceMesh
}

func (c *Controller) handleServiceAdd(obj interface{}) {
	svc := obj.(*crv1.Service)
	glog.V(4).Infof("Adding Service %s", svc.Name)
	c.enqueueService(svc)
}

func (c *Controller) handleServiceUpdate(old, cur interface{}) {
	oldSvc := old.(*crv1.Service)
	curSvc := cur.(*crv1.Service)
	glog.V(4).Infof("Updating Service %s", oldSvc.Name)
	c.enqueueService(curSvc)
}

func (c *Controller) handleServiceDelete(obj interface{}) {
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
	c.enqueueService(svc)
}

func (c *Controller) handleNodePoolAdd(obj interface{}) {
	nodePool := obj.(*crv1.NodePool)
	glog.V(4).Infof("Adding NodePool %s/%s", nodePool.Namespace, nodePool.Name)

	services, err := util.ServicesForNodePool(c.latticeClient, nodePool)
	if err != nil {
		// FIXME(kevinrosendahl): what to do here?
		return
	}

	for _, service := range services {
		c.enqueueService(&service)
	}
}

func (c *Controller) handleNodePoolUpdate(old, cur interface{}) {
	oldNodePool := old.(*crv1.NodePool)
	curNodePool := cur.(*crv1.NodePool)
	glog.V(4).Infof("Updating NodePool %s/%s", curNodePool.Namespace, curNodePool.Name)

	if oldNodePool.ResourceVersion == curNodePool.ResourceVersion {
		// Periodic resync will send update events for all known Deployments.
		// Two different versions of the same Deployment will always have different RVs.
		glog.V(5).Info("NodePool ResourceVersions are the same")
		return
	}

	services, err := util.ServicesForNodePool(c.latticeClient, curNodePool)
	if err != nil {
		// FIXME(kevinrosendahl): what to do here?
		return
	}

	for _, service := range services {
		c.enqueueService(&service)
	}
}

// handleDeploymentAdd enqueues the Service that manages a Deployment when the Deployment is created.
func (c *Controller) handleDeploymentAdd(obj interface{}) {
	d := obj.(*appsv1.Deployment)

	if d.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleDeploymentDelete(d)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(d); controllerRef != nil {
		svc := c.resolveControllerRef(d.Namespace, controllerRef)

		// Not a Service Deployment.
		if svc == nil {
			return
		}

		glog.V(4).Infof("Deployment %s added.", d.Name)
		c.enqueueService(svc)
		return
	}

	// Otherwise, it's an orphan. These should never exist. All deployments should be run by some
	// controller.
	// FIXME(kevinrosendahl): send warn event
}

// handleDeploymentUpdate figures out what Service manages a Deployment when the Deployment
// is updated and enqueues it.
func (c *Controller) handleDeploymentUpdate(old, cur interface{}) {
	glog.V(5).Info("Got Deployment update")
	oldD := old.(*appsv1.Deployment)
	curD := cur.(*appsv1.Deployment)
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
		if b := c.resolveControllerRef(oldD.Namespace, oldControllerRef); b != nil {
			// FIXME(kevinrosendahl): send error event here, this should not happen
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		svc := c.resolveControllerRef(curD.Namespace, curControllerRef)

		// Not a Service Deployment
		if svc == nil {
			return
		}

		c.enqueueService(svc)
		return
	}

	// Otherwise, it's an orphan. These should never exist. All deployments should be run by some
	// controller.
	// FIXME(kevinrosendahl): send warn event
}

// handleDeploymentDelete enqueues the Service that manages a Deployment when
// the Deployment is deleted.
func (c *Controller) handleDeploymentDelete(obj interface{}) {
	d, ok := obj.(*appsv1.Deployment)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*appsv1.Deployment)
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

	svc := c.resolveControllerRef(d.Namespace, controllerRef)

	// Not a Service Deployment
	if svc == nil {
		return
	}

	glog.V(4).Infof("Deployment %s deleted.", d.Name)
	c.enqueueService(svc)
}

func (c *Controller) handleKubeServiceAdd(obj interface{}) {
	kubeService := obj.(*corev1.Service)
	glog.V(4).Infof("kube Service %v/%v added", kubeService.Namespace, kubeService.Name)

	if kubeService.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleKubeServiceDelete(kubeService)
		return
	}

	name, err := kubeutil.GetServiceNameForKubeService(kubeService)
	if err != nil {
		// The kube service wasn't for a Service
		return
	}

	service, err := c.serviceLister.Services(kubeService.Namespace).Get(name)
	if err != nil {
		// FIXME(kevinrosendahl): send warn event
		return
	}

	c.enqueueService(service)
}

func (c *Controller) handleKubeServiceUpdate(old, cur interface{}) {
	glog.V(5).Info("Got kube Service update")
	oldKubeService := old.(*corev1.Service)
	curKubeService := cur.(*corev1.Service)
	if curKubeService.ResourceVersion == oldKubeService.ResourceVersion {
		// Periodic resync will send update events for all known Services.
		// Two different versions of the same Service will always have different RVs.
		glog.V(5).Info("kube Service %v/%v ResourceVersions are the same", curKubeService.Namespace, curKubeService.Name)
		return
	}

	name, err := kubeutil.GetServiceNameForKubeService(curKubeService)
	if err != nil {
		// The kube service wasn't for a Service
		return
	}

	service, err := c.serviceLister.Services(curKubeService.Namespace).Get(name)
	if err != nil {
		// FIXME(kevinrosendahl): send warn event
		return
	}

	c.enqueueService(service)
}

func (c *Controller) handleKubeServiceDelete(obj interface{}) {
	kubeService, ok := obj.(*corev1.Service)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		kubeService, ok = tombstone.Obj.(*corev1.Service)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	name, err := kubeutil.GetServiceNameForKubeService(kubeService)
	if err != nil {
		// The kube service wasn't for a Service
		return
	}

	service, err := c.serviceLister.Services(kubeService.Namespace).Get(name)
	if err != nil {
		// FIXME(kevinrosendahl): send warn event
		return
	}

	c.enqueueService(service)
}

func (c *Controller) handleServiceAddressAdd(obj interface{}) {
	address := obj.(*crv1.ServiceAddress)
	glog.V(4).Infof("ServiceAddress %v/%v added", address.Namespace, address.Name)

	if address.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleKubeServiceDelete(address)
		return
	}

	service, err := c.serviceLister.Services(address.Namespace).Get(address.Name)
	if err != nil {
		// FIXME(kevinrosendahl): send warn event
		return
	}

	c.enqueueService(service)
}

func (c *Controller) handleServiceAddressUpdate(old, cur interface{}) {
	glog.V(5).Info("Got kube Service update")
	oldAddress := old.(*crv1.ServiceAddress)
	curAddress := cur.(*crv1.ServiceAddress)
	if curAddress.ResourceVersion == oldAddress.ResourceVersion {
		// Periodic resync will send update events for all known Services.
		// Two different versions of the same Service will always have different RVs.
		glog.V(5).Info("kube Service %v/%v ResourceVersions are the same", curAddress.Namespace, curAddress.Name)
		return
	}

	service, err := c.serviceLister.Services(curAddress.Namespace).Get(curAddress.Name)
	if err != nil {
		// FIXME(kevinrosendahl): send warn event
		return
	}

	c.enqueueService(service)
}

func (c *Controller) handleServiceAddressDelete(obj interface{}) {
	address, ok := obj.(*crv1.ServiceAddress)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		address, ok = tombstone.Obj.(*crv1.ServiceAddress)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	service, err := c.serviceLister.Services(address.Namespace).Get(address.Name)
	if err != nil {
		// FIXME(kevinrosendahl): send warn event
		return
	}

	c.enqueueService(service)
}

func (c *Controller) enqueue(svc *crv1.Service) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(svc)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", svc, err))
		return
	}

	c.queue.Add(key)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (c *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.Service {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}

	svc, err := c.serviceLister.Services(namespace).Get(controllerRef.Name)
	if err != nil {
		// FIXME(kevinrosendahl): send error?
		return nil
	}

	if svc.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return svc
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

// syncService will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncService(key string) error {
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

	service, err := c.serviceLister.Services(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("Service %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	nodePool, err := c.syncServiceNodePool(service)
	if err != nil {
		return err
	}

	deployment, err := c.syncServiceDeployment(service, nodePool)
	if err != nil {
		return err
	}

	kubeService, err := c.syncServiceKubeService(service)
	if err != nil {
		return err
	}

	serviceAddress, err := c.syncServiceServiceAddress(service)
	if err != nil {
		return err
	}

	_, err = c.syncServiceStatus(service, deployment, kubeService, nodePool, serviceAddress)
	return err
}
