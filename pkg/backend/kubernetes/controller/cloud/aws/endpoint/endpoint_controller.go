package endpoint

import (
	"fmt"
	"time"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/terraform"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type Controller struct {
	syncHandler     func(bKey string) error
	enqueueEndpoint func(cb *latticev1.Endpoint)

	latticeID types.LatticeID

	latticeClient latticeclientset.Interface

	awsCloudProvider        aws.CloudProvider
	terraformModuleRoot     string
	terraformBackendOptions *terraform.BackendOptions

	endpointLister       latticelisters.EndpointLister
	endpointListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	latticeID types.LatticeID,
	awsCloudProvider aws.CloudProvider,
	terraformModuleRoot string,
	terraformBackendOptions *terraform.BackendOptions,
	latticeClient latticeclientset.Interface,
	endpointInformer latticeinformers.EndpointInformer,
) *Controller {
	sc := &Controller{
		latticeID:               latticeID,
		latticeClient:           latticeClient,
		awsCloudProvider:        awsCloudProvider,
		terraformModuleRoot:     terraformModuleRoot,
		terraformBackendOptions: terraformBackendOptions,
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service"),
	}

	sc.syncHandler = sc.syncEndpoint
	sc.enqueueEndpoint = sc.enqueue

	endpointInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.handleEndpointAdd,
		UpdateFunc: sc.handleEndpointUpdate,
		DeleteFunc: sc.handleEndpointDelete,
	})
	sc.endpointLister = endpointInformer.Lister()
	sc.endpointListerSynced = endpointInformer.Informer().HasSynced

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
	if !cache.WaitForCacheSync(stopCh, c.endpointListerSynced) {
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

func (c *Controller) handleEndpointAdd(obj interface{}) {
	endpoint := obj.(*latticev1.Endpoint)
	glog.V(4).Infof("Endpoint %v/%v added", endpoint.Namespace, endpoint.Name)

	if endpoint.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleEndpointDelete(endpoint)
		return
	}

	c.enqueueEndpoint(endpoint)
}

func (c *Controller) handleEndpointUpdate(old, cur interface{}) {
	oldEndpoint := old.(*latticev1.Endpoint)
	curEndpoint := cur.(*latticev1.Endpoint)
	glog.V(5).Info("Got Endpoint %v/%v update", curEndpoint.Namespace, curEndpoint.Name)
	if curEndpoint.ResourceVersion == oldEndpoint.ResourceVersion {
		// Periodic resync will send update events for all known Services.
		// Two different versions of the same Service will always have different RVs.
		glog.V(5).Info("Endpoint %v/%v ResourceVersions are the same", curEndpoint.Namespace, curEndpoint.Name)
		return
	}

	c.enqueueEndpoint(curEndpoint)
}

func (c *Controller) handleEndpointDelete(obj interface{}) {
	endpoint, ok := obj.(*latticev1.Endpoint)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		endpoint, ok = tombstone.Obj.(*latticev1.Endpoint)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}

	c.enqueueEndpoint(endpoint)
}

func (c *Controller) enqueue(endpoint *latticev1.Endpoint) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(endpoint)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", endpoint, err))
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

// syncEndpoint will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncEndpoint(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing Endpoint %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing Endpoint %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	endpoint, err := c.endpointLister.Endpoints(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			glog.V(2).Infof("Endpoint %v has been deleted", key)
			return nil
		}

		return err
	}

	if endpoint.DeletionTimestamp != nil {
		return c.syncDeletedEndpoint(endpoint)
	}

	endpoint, err = c.addFinalizer(endpoint)
	if err != nil {
		return err
	}

	err = c.provisionEndpoint(endpoint)
	if err != nil {
		return err
	}

	status := latticev1.EndpointStatus{
		State: latticev1.EndpointStateCreated,
	}

	_, err = c.updateEndpointStatus(endpoint, status)
	return err
}
