package loadbalancer

import (
	"fmt"
	"time"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	coreinformers "k8s.io/client-go/informers/core/v1"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = latticev1.SchemeGroupVersion.WithKind("LoadBalancer")

type Controller struct {
	syncHandler         func(bKey string) error
	enqueueLoadBalancer func(cb *latticev1.LoadBalancer)

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	localCloudProvider local.CloudProvider

	loadBalancerLister       latticelisters.LoadBalancerLister
	loadBalancerListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	kubeServiceLister       corelisters.ServiceLister
	kubeServiceListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	localCloudProvider local.CloudProvider,
	loadBalancerInformer latticeinformers.LoadBalancerInformer,
	serviceInformer latticeinformers.ServiceInformer,
	kubeServiceInformer coreinformers.ServiceInformer,
) *Controller {
	sc := &Controller{
		kubeClient:         kubeClient,
		latticeClient:      latticeClient,
		localCloudProvider: localCloudProvider,
		queue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service"),
	}

	sc.syncHandler = sc.syncLoadBalancer
	sc.enqueueLoadBalancer = sc.enqueue

	loadBalancerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleLoadBalancerAdd,
		UpdateFunc: sc.handleLoadBalancerUpdate,
		DeleteFunc: sc.handleLoadBalancerDelete,
	})
	sc.loadBalancerLister = loadBalancerInformer.Lister()
	sc.loadBalancerListerSynced = loadBalancerInformer.Informer().HasSynced

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

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting node pool controller")
	defer glog.Infof("Shutting down node pool controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.loadBalancerListerSynced, c.serviceListerSynced, c.kubeServiceListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced")

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

func (c *Controller) handleLoadBalancerAdd(obj interface{}) {
	loadBalancer := obj.(*latticev1.LoadBalancer)
	glog.V(4).Infof("LoadBalancer %v/%v added", loadBalancer.Namespace, loadBalancer.Name)

	if loadBalancer.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleLoadBalancerDelete(loadBalancer)
		return
	}

	c.enqueueLoadBalancer(loadBalancer)
}

func (c *Controller) handleLoadBalancerUpdate(old, cur interface{}) {
	oldLoadBalancer := old.(*latticev1.LoadBalancer)
	curLoadBalancer := cur.(*latticev1.LoadBalancer)
	glog.V(5).Info("Got LoadBalancer %v/%v update", curLoadBalancer.Namespace, curLoadBalancer.Name)
	if curLoadBalancer.ResourceVersion == oldLoadBalancer.ResourceVersion {
		// Periodic resync will send update events for all known Services.
		// Two different versions of the same Service will always have different RVs.
		glog.V(5).Info("LoadBalancer %v/%v ResourceVersions are the same", curLoadBalancer.Namespace, curLoadBalancer.Name)
		return
	}

	c.enqueueLoadBalancer(curLoadBalancer)
}

func (c *Controller) handleLoadBalancerDelete(obj interface{}) {
	loadBalancer, ok := obj.(*latticev1.LoadBalancer)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		loadBalancer, ok = tombstone.Obj.(*latticev1.LoadBalancer)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	c.enqueueLoadBalancer(loadBalancer)
}

func (c *Controller) handleServiceAdd(obj interface{}) {
	service := obj.(*latticev1.Service)
	glog.V(4).Infof("Service %v/%v added", service.Namespace, service.Name)

	if service.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleServiceDelete(service)
		return
	}

	loadBalancer, err := c.loadBalancerLister.LoadBalancers(service.Namespace).Get(service.Name)
	if err != nil {
		// FIXME: handle error
		return
	}

	c.enqueueLoadBalancer(loadBalancer)
}

func (c *Controller) handleServiceUpdate(old, cur interface{}) {
	glog.V(5).Info("Got Service update")
	oldService := old.(*latticev1.Service)
	curService := cur.(*latticev1.Service)
	if curService.ResourceVersion == oldService.ResourceVersion {
		// Periodic resync will send update events for all known Services.
		// Two different versions of the same Service will always have different RVs.
		glog.V(5).Info("Service %v/%v ResourceVersions are the same", curService.Namespace, curService.Name)
		return
	}

	loadBalancer, err := c.loadBalancerLister.LoadBalancers(curService.Namespace).Get(curService.Name)
	if err != nil {
		// FIXME: handle error
		return
	}

	c.enqueueLoadBalancer(loadBalancer)
}

func (c *Controller) handleServiceDelete(obj interface{}) {
	service, ok := obj.(*latticev1.Service)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		service, ok = tombstone.Obj.(*latticev1.Service)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	loadBalancer, err := c.loadBalancerLister.LoadBalancers(service.Namespace).Get(service.Name)
	if err != nil {
		// FIXME: handle error
		return
	}

	c.enqueueLoadBalancer(loadBalancer)
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

	name, err := kubeutil.GetLoadBalancerNameForKubeService(kubeService)
	if err != nil {
		// The kube loadBalancer wasn't for a LoadBalancer
		return
	}

	loadBalancer, err := c.loadBalancerLister.LoadBalancers(kubeService.Namespace).Get(name)
	if err != nil {
		// FIXME(kevinrosendahl): send warn event
		return
	}

	c.enqueueLoadBalancer(loadBalancer)
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

	name, err := kubeutil.GetLoadBalancerNameForKubeService(curKubeService)
	if err != nil {
		// The kube loadBalancer wasn't for a LoadBalancer
		return
	}

	loadBalancer, err := c.loadBalancerLister.LoadBalancers(curKubeService.Namespace).Get(name)
	if err != nil {
		// FIXME(kevinrosendahl): send warn event
		return
	}

	c.enqueueLoadBalancer(loadBalancer)
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

	name, err := kubeutil.GetLoadBalancerNameForKubeService(kubeService)
	if err != nil {
		// The kube loadBalancer wasn't for a LoadBalancer
		return
	}

	loadBalancer, err := c.loadBalancerLister.LoadBalancers(kubeService.Namespace).Get(name)
	if err != nil {
		// FIXME(kevinrosendahl): send warn event
		return
	}

	c.enqueueLoadBalancer(loadBalancer)
}

func (c *Controller) enqueue(loadBalancer *latticev1.LoadBalancer) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(loadBalancer)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", loadBalancer, err))
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

// syncLoadBalancer will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncLoadBalancer(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing LoadBalancer %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing LoadBalancer %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	loadBalancer, err := c.loadBalancerLister.LoadBalancers(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("LoadBalancer %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	kubeService, err := c.syncLoadBalancerKubeService(loadBalancer)
	if err != nil {
		return err
	}

	// Copy so the shared cache isn't mutated
	loadBalancer = loadBalancer.DeepCopy()

	ports := map[int32]latticev1.LoadBalancerPort{}
	for _, port := range kubeService.Spec.Ports {
		ports[port.Port] = latticev1.LoadBalancerPort{
			Address: fmt.Sprintf("%v:%v", c.localCloudProvider.IP(), port.NodePort),
		}
	}

	loadBalancer.Status.Ports = ports
	loadBalancer.Status.State = latticev1.LoadBalancerStateCreated

	_, err = c.latticeClient.LatticeV1().LoadBalancers(loadBalancer.Namespace).Update(loadBalancer)
	return err
}