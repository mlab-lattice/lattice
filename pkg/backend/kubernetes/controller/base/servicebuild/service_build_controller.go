package servicebuild

import (
	"fmt"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

type Controller struct {
	syncHandler func(svcBuildKey string) error
	enqueue     func(svcBuild *crv1.ServiceBuild)

	latticeClient latticeclientset.Interface

	serviceBuildLister       latticelisters.ServiceBuildLister
	serviceBuildListerSynced cache.InformerSynced

	componentBuildLister       latticelisters.ComponentBuildLister
	componentBuildListerSynced cache.InformerSynced

	// recentComponentBuilds holds a map of namespaces which map to a map of definition
	// hashes which map to the name of a ComponentBuild that was recently created
	// in the namespace. recentComponentBuilds should always hold the Name of the most
	// recently created ComponentBuild for a given definition hash.
	// See createComponentBuilds for more information.
	// FIXME: add some GC on this map so it doesn't grow forever (maybe remove in addComponentBuild)
	recentComponentBuildsLock sync.RWMutex
	recentComponentBuilds     map[string]map[string]string

	queue workqueue.RateLimitingInterface
}

func NewController(
	latticeClient latticeclientset.Interface,
	serviceBuildInformer latticeinformers.ServiceBuildInformer,
	componentBuildInformer latticeinformers.ComponentBuildInformer,
) *Controller {
	sbc := &Controller{
		latticeClient:         latticeClient,
		recentComponentBuilds: make(map[string]map[string]string),
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-build"),
	}

	sbc.syncHandler = sbc.syncServiceBuild
	sbc.enqueue = sbc.enqueueServiceBuild

	serviceBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addServiceBuild,
		UpdateFunc: sbc.updateServiceBuild,
		// TODO: for now it is assumed that ServiceBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ServiceBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document SvcB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.serviceBuildLister = serviceBuildInformer.Lister()
	sbc.serviceBuildListerSynced = serviceBuildInformer.Informer().HasSynced

	componentBuildInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sbc.addComponentBuild,
		UpdateFunc: sbc.updateComponentBuild,
		// TODO: for now it is assumed that ComponentBuilds are not deleted.
		// in the future we'll probably want to add a GC process for ComponentBuilds.
		// At that point we should listen here for those deletes.
		// FIXME: Document CB GC ideas (need to write down last used date, lock properly, etc)
	})
	sbc.componentBuildLister = componentBuildInformer.Lister()
	sbc.componentBuildListerSynced = componentBuildInformer.Informer().HasSynced

	return sbc
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting service-build controller")
	defer glog.Infof("Shutting down service-build controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, c.serviceBuildListerSynced, c.componentBuildListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced.")

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

func (c *Controller) addServiceBuild(obj interface{}) {
	build := obj.(*crv1.ServiceBuild)
	glog.V(4).Infof("Adding ServiceBuild %s", build.Name)
	c.enqueueServiceBuild(build)
}

func (c *Controller) updateServiceBuild(old, cur interface{}) {
	oldBuild := old.(*crv1.ServiceBuild)
	curBuild := cur.(*crv1.ServiceBuild)
	glog.V(4).Infof("Updating ServiceBuild %s", oldBuild.Name)
	c.enqueueServiceBuild(curBuild)
}

// addComponentBuild enqueues any ServiceBuilds which may be interested in it when
// a ComponentBuild is added.
func (c *Controller) addComponentBuild(obj interface{}) {
	componentBuild := obj.(*crv1.ComponentBuild)

	if componentBuild.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		// FIXME: send error event
		return
	}

	glog.V(4).Infof("ComponentBuild %s added.", componentBuild.Name)
	builds, err := c.getServiceBuildsForComponentBuild(componentBuild)
	if err != nil {
		// FIXME: send error event?
	}
	for _, build := range builds {
		c.enqueueServiceBuild(build)
	}
}

// updateComponentBuild enqueues any ServiceBuilds which may be interested in it when
// a ComponentBuild is updated.
func (c *Controller) updateComponentBuild(old, cur interface{}) {
	glog.V(5).Info("Got ComponentBuild update")
	oldComponentBuild := old.(*crv1.ComponentBuild)
	curComponentBuild := cur.(*crv1.ComponentBuild)
	if curComponentBuild.ResourceVersion == oldComponentBuild.ResourceVersion {
		// Periodic resync will send update events for all known ComponentBuilds.
		// Two different versions of the same job will always have different RVs.
		glog.V(5).Info("ComponentBuild ResourceVersions are the same")
		return
	}

	builds, err := c.getServiceBuildsForComponentBuild(curComponentBuild)
	if err != nil {
		// FIXME: send error event?
	}
	for _, build := range builds {
		c.enqueueServiceBuild(build)
	}
}

func (c *Controller) getServiceBuildsForComponentBuild(componentBuild *crv1.ComponentBuild) ([]*crv1.ServiceBuild, error) {
	// TODO: this could eventually get expensive
	builds, err := c.serviceBuildLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var matchingBuilds []*crv1.ServiceBuild
	for _, build := range builds {
		// Check to see if the ServiceBuild is waiting on this ComponentBuild
		if _, ok := build.Status.ComponentBuildStatuses[componentBuild.Name]; ok {
			matchingBuilds = append(matchingBuilds, build)
		}
	}

	return matchingBuilds, nil
}

func (c *Controller) enqueueServiceBuild(build *crv1.ServiceBuild) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(build)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", build, err))
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

// syncServiceBuild will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncServiceBuild(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing ServiceBuild %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing ServiceBuild %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	build, err := c.serviceBuildLister.ServiceBuilds(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.V(2).Infof("ServiceBuild %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	stateInfo, err := c.calculateState(build)
	if err != nil {
		return err
	}

	glog.V(5).Infof("ServiceBuild %v state: %v", build.Name, stateInfo.state)

	switch stateInfo.state {
	case stateHasFailedCBuilds:
		return c.syncFailedServiceBuild(build, stateInfo)
	case stateHasOnlyRunningOrSucceededCBuilds:
		return c.syncRunningServiceBuild(build, stateInfo)
	case stateNoFailuresNeedsNewCBuilds:
		return c.syncMissingComponentBuildsServiceBuild(build, stateInfo)
	case stateAllCBuildsSucceeded:
		return c.syncSucceededServiceBuild(build, stateInfo)
	default:
		return fmt.Errorf("ServiceBuild %v in unexpected state %v", key, stateInfo.state)
	}
}
