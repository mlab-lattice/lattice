package address

import (
	"fmt"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"

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

type Controller struct {
	syncHandler func(key string) error
	enqueue     func(address *latticev1.Address)

	namespacePrefix string
	latticeID       v1.LatticeID

	// NOTE: you must get a read lock on the configLock for the duration
	//       of your use of the cloudProvider
	staticCloudProviderOptions *cloudprovider.Options
	cloudProvider              cloudprovider.Interface

	// NOTE: you must get a read lock on the configLock for the duration
	//       of your use of the serviceMesh
	staticServiceMeshOptions *servicemesh.Options
	serviceMesh              servicemesh.Interface

	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	configLister       latticelisters.ConfigLister
	configListerSynced cache.InformerSynced
	configSetChan      chan struct{}
	configSet          bool
	configLock         sync.RWMutex
	config             latticev1.ConfigSpec

	addressLister       latticelisters.AddressLister
	addressListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	kubeServiceLister       corelisters.ServiceLister
	kubeServiceListerSynced cache.InformerSynced

	nodePoolLister       latticelisters.NodePoolLister
	nodePoolListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	namespacePrefix string,
	latticeID v1.LatticeID,
	cloudProviderOptions *cloudprovider.Options,
	serviceMeshOptions *servicemesh.Options,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	configInformer latticeinformers.ConfigInformer,
	addressInformer latticeinformers.AddressInformer,
	serviceInformer latticeinformers.ServiceInformer,
	kubeServiceInformer coreinformers.ServiceInformer,
	nodePoolInformer latticeinformers.NodePoolInformer,
) *Controller {
	c := &Controller{
		namespacePrefix: namespacePrefix,
		latticeID:       latticeID,

		staticCloudProviderOptions: cloudProviderOptions,
		staticServiceMeshOptions:   serviceMeshOptions,

		latticeClient: latticeClient,
		kubeClient:    kubeClient,

		configSetChan: make(chan struct{}),

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "address"),
	}

	c.syncHandler = c.syncAddress
	c.enqueue = c.enqueueAddress

	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    c.handleConfigAdd,
		UpdateFunc: c.handleConfigUpdate,
	})
	c.configLister = configInformer.Lister()
	c.configListerSynced = configInformer.Informer().HasSynced

	addressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAddressAdd,
		UpdateFunc: c.handleAddressUpdate,
		DeleteFunc: c.handleAddressDelete,
	})
	c.addressLister = addressInformer.Lister()
	c.addressListerSynced = addressInformer.Informer().HasSynced

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleServiceAdd,
		UpdateFunc: c.handleServiceUpdate,
		DeleteFunc: c.handleServiceDelete,
	})
	c.serviceLister = serviceInformer.Lister()
	c.serviceListerSynced = serviceInformer.Informer().HasSynced

	c.kubeServiceLister = kubeServiceInformer.Lister()
	c.kubeServiceListerSynced = kubeServiceInformer.Informer().HasSynced

	c.nodePoolLister = nodePoolInformer.Lister()
	c.nodePoolListerSynced = nodePoolInformer.Informer().HasSynced

	return c
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("starting address controller")
	defer glog.Infof("shutting down service controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(
		stopCh,
		c.configListerSynced,
		c.addressListerSynced,
		c.serviceListerSynced,
		c.kubeServiceListerSynced,
		c.nodePoolListerSynced,
	) {
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

func (c *Controller) enqueueAddress(svc *latticev1.Address) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(svc)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", svc, err))
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

// syncAddress will sync the address with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncAddress(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("started syncing address %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("finished syncing address %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	address, err := c.addressLister.Addresses(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			glog.V(2).Infof("address %v has been deleted", key)
			return nil
		}

		return err
	}

	if address.DeletionTimestamp != nil {
		return c.syncDeletedAddress(address)
	}

	address, err = c.addFinalizer(address)
	if err != nil {
		return err
	}

	if address.Spec.Service != nil {
		return c.syncServiceAddress(address)
	}

	if address.Spec.ExternalName != nil {
		return c.syncExternalNameAddress(address)
	}

	return fmt.Errorf("%v has neither service nor external name", address.Description(c.namespacePrefix))
}
