package nodepool

import (
	"fmt"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type Controller struct {
	syncHandler func(key string) error
	enqueue     func(nodePool *latticev1.NodePool)

	namespacePrefix string
	latticeID       v1.LatticeID

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	kubeInformerFactory    kubeinformers.SharedInformerFactory
	latticeInformerFactory latticeinformers.SharedInformerFactory

	staticCloudProviderOptions *cloudprovider.Options
	cloudProvider              cloudprovider.Interface

	configLister       latticelisters.ConfigLister
	configListerSynced cache.InformerSynced
	configSetChan      chan struct{}
	configSet          bool
	configLock         sync.RWMutex
	config             latticev1.ConfigSpec

	nodePoolLister       latticelisters.NodePoolLister
	nodePoolListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	kubeNodeLister       corelisters.NodeLister
	kubeNodeListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	namespacePrefix string,
	latticeID v1.LatticeID,
	cloudProviderOptions *cloudprovider.Options,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	latticeInformerFactory latticeinformers.SharedInformerFactory,
) *Controller {
	sc := &Controller{
		namespacePrefix: namespacePrefix,
		latticeID:       latticeID,

		kubeClient:    kubeClient,
		latticeClient: latticeClient,

		kubeInformerFactory:    kubeInformerFactory,
		latticeInformerFactory: latticeInformerFactory,

		staticCloudProviderOptions: cloudProviderOptions,

		configSetChan: make(chan struct{}),

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "nodepool"),
	}

	sc.syncHandler = sc.syncNodePool
	sc.enqueue = sc.enqueueNodePool

	configInformer := latticeInformerFactory.Lattice().V1().Configs()
	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    sc.handleConfigAdd,
		UpdateFunc: sc.handleConfigUpdate,
		// TODO(kevindrosendahl): for now it is assumed that configs are not deleted.
	})
	sc.configLister = configInformer.Lister()
	sc.configListerSynced = configInformer.Informer().HasSynced

	nodePoolInformer := latticeInformerFactory.Lattice().V1().NodePools()
	nodePoolInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleNodePoolAdd,
		UpdateFunc: sc.handleNodePoolUpdate,
		DeleteFunc: sc.handleNodePoolDelete,
	})
	sc.nodePoolLister = nodePoolInformer.Lister()
	sc.nodePoolListerSynced = nodePoolInformer.Informer().HasSynced

	serviceInformer := latticeInformerFactory.Lattice().V1().Services()
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleServiceAdd,
		UpdateFunc: sc.handleServiceUpdate,
		DeleteFunc: sc.handleServiceDelete,
	})
	sc.serviceLister = serviceInformer.Lister()
	sc.serviceListerSynced = serviceInformer.Informer().HasSynced

	kubeNodeInformer := kubeInformerFactory.Core().V1().Nodes()
	kubeNodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleKubeNodeAdd,
		UpdateFunc: sc.handleKubeNodeUpdate,
		DeleteFunc: sc.handleKubeNodeDelete,
	})
	sc.kubeNodeLister = kubeNodeInformer.Lister()
	sc.serviceListerSynced = kubeNodeInformer.Informer().HasSynced

	return sc
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("starting node pool controller")
	defer glog.Infof("shutting down endpoint controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.configListerSynced, c.nodePoolListerSynced, c.serviceListerSynced) {
		return
	}

	glog.V(4).Info("caches synced, waiting for config to be set")

	// wait for config to be set
	<-c.configSetChan

	glog.V(4).Info("config set")

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

func (c *Controller) enqueueNodePool(nodePool *latticev1.NodePool) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(nodePool)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", nodePool, err))
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

// syncNodePool will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncNodePool(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("started syncing node pool %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("finished syncing node pool %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	nodePool, err := c.nodePoolLister.NodePools(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			glog.V(2).Infof("node pool %v has been deleted", key)
			return nil
		}

		return err
	}

	if nodePool.DeletionTimestamp != nil {
		return c.syncDeletedNodePool(nodePool)
	}

	nodePool, err = c.addFinalizer(nodePool)
	if err != nil {
		return nil
	}

	return c.syncActiveNodePool(nodePool)
}
