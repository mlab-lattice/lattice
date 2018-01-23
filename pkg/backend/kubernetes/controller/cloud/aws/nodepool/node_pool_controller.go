package nodepool

import (
	"fmt"
	"sync"
	"time"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/system/pkg/terraform"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type Controller struct {
	syncHandler     func(bKey string) error
	enqueueNodePool func(cb *latticev1.NodePool)

	clusterID types.ClusterID

	latticeClient latticeclientset.Interface

	awsCloudProvider        aws.CloudProvider
	terraformModuleRoot     string
	terraformBackendOptions *terraform.BackendOptions

	configLister       latticelisters.ConfigLister
	configListerSynced cache.InformerSynced
	configSetChan      chan struct{}
	configSet          bool
	configLock         sync.RWMutex
	config             latticev1.ConfigSpec

	nodePoolLister       latticelisters.NodePoolLister
	nodePoolListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	clusterID types.ClusterID,
	awsCloudProvider aws.CloudProvider,
	terraformModuleRoot string,
	terraformBackendOptions *terraform.BackendOptions,
	latticeClient latticeclientset.Interface,
	configInformer latticeinformers.ConfigInformer,
	nodePoolInformer latticeinformers.NodePoolInformer,
) *Controller {
	sc := &Controller{
		clusterID:               clusterID,
		latticeClient:           latticeClient,
		awsCloudProvider:        awsCloudProvider,
		terraformModuleRoot:     terraformModuleRoot,
		terraformBackendOptions: terraformBackendOptions,
		configSetChan:           make(chan struct{}),
		queue:                   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service"),
	}

	sc.syncHandler = sc.syncNodePool
	sc.enqueueNodePool = sc.enqueue

	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    sc.handleConfigAdd,
		UpdateFunc: sc.handleConfigUpdate,
		// TODO(kevinrosendahl): for now it is assumed that ComponentBuilds are not deleted.
	})
	sc.configLister = configInformer.Lister()
	sc.configListerSynced = configInformer.Informer().HasSynced

	nodePoolInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleNodePoolAdd,
		UpdateFunc: sc.handleNodePoolUpdate,
		DeleteFunc: sc.handleNodePoolDelete,
	})
	sc.nodePoolLister = nodePoolInformer.Lister()
	sc.nodePoolListerSynced = nodePoolInformer.Informer().HasSynced

	return sc
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting endpoint controller")
	defer glog.Infof("Shutting down endpoint controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.configListerSynced, c.nodePoolListerSynced) {
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
	config := obj.(*latticev1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = config.DeepCopy().Spec

	c.awsCloudProvider = c.newAWSCloudProvider()

	if !c.configSet {
		c.configSet = true
		close(c.configSetChan)
	}
}

func (c *Controller) handleConfigUpdate(old, cur interface{}) {
	oldConfig := old.(*latticev1.Config)
	curConfig := cur.(*latticev1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = curConfig.DeepCopy().Spec

	c.awsCloudProvider = c.newAWSCloudProvider()
}

func (c *Controller) newAWSCloudProvider() aws.CloudProvider {
	awsOptions := &aws.Options{
		Region:    c.awsCloudProvider.Region(),
		AccountID: c.awsCloudProvider.AccountID(),
		VPCID:     c.awsCloudProvider.VPCID(),

		Route53PrivateZoneID:      c.awsCloudProvider.Route53PrivateZoneID(),
		SubnetIDs:                 c.awsCloudProvider.SubnetIDs(),
		MasterNodeSecurityGroupID: c.awsCloudProvider.MasterNodeSecurityGroupID(),

		BaseNodeAMIID: c.config.CloudProvider.AWS.BaseNodeAMIID,
		KeyName:       c.config.CloudProvider.AWS.KeyName,
	}

	return aws.NewCloudProvider(awsOptions)
}

func (c *Controller) handleNodePoolAdd(obj interface{}) {
	nodePool := obj.(*latticev1.NodePool)
	glog.V(4).Infof("NodePool %v/%v added", nodePool.Namespace, nodePool.Name)

	if nodePool.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleNodePoolDelete(nodePool)
		return
	}

	c.enqueueNodePool(nodePool)
}

func (c *Controller) handleNodePoolUpdate(old, cur interface{}) {
	oldNodePool := old.(*latticev1.NodePool)
	curNodePool := cur.(*latticev1.NodePool)
	glog.V(5).Info("Got NodePool %v/%v update", curNodePool.Namespace, curNodePool.Name)
	if curNodePool.ResourceVersion == oldNodePool.ResourceVersion {
		// Periodic resync will send update events for all known Services.
		// Two different versions of the same Service will always have different RVs.
		glog.V(5).Info("NodePool %v/%v ResourceVersions are the same", curNodePool.Namespace, curNodePool.Name)
		return
	}

	c.enqueueNodePool(curNodePool)
}

func (c *Controller) handleNodePoolDelete(obj interface{}) {
	nodePool, ok := obj.(*latticev1.NodePool)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		nodePool, ok = tombstone.Obj.(*latticev1.NodePool)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	c.enqueueNodePool(nodePool)
}

func (c *Controller) enqueue(nodePool *latticev1.NodePool) {
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
	glog.V(4).Infof("Started syncing NodePool %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing NodePool %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	nodePool, err := c.nodePoolLister.NodePools(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			glog.V(2).Infof("NodePool %v has been deleted", key)
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

	nodePool, err = c.syncNodePoolState(nodePool)
	if err != nil {
		return nil
	}

	_, err = c.provisionNodePool(nodePool)
	return err
}
