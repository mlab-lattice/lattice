package local

import (
    "fmt"
    "time"
    "os"
    "sync"

    crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
    latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
    latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
    latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

    "github.com/golang/glog"
    "k8s.io/client-go/tools/cache"
    "k8s.io/client-go/util/workqueue"

    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/util/runtime"
    "k8s.io/apimachinery/pkg/util/wait"
)

type Controller struct {
    //Contains the controller specific for updating DNS, Watches Address changes.
    syncAddressUpdate	func(bKey string) error
    enqueueAddressUpdate func(sysBuild *crv1.Endpoint)

    cnameList	map[string]string

    hasRecentlyUpdated 	bool
    lock 				sync.RWMutex

    latticeClient latticeclientset.Interface

    addressLister latticelisters.EndpointLister
    addressListerSynced cache.InformerSynced

    queue workqueue.RateLimitingInterface
}

type ServerConfig struct {
    serverConfigPath string
    serverResolvPath string
}

var (
    config ServerConfig
    updateWaitBeforeFlushTimer = 15
)

func NewController(
    serverConfigPath string,
    serverResolvPath string,
    latticeClient  	 latticeclientset.Interface,
    addressInformer  latticeinformers.EndpointInformer,
) *Controller {

    addrc := &Controller{
        latticeClient: latticeClient,
        queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system"),
    }

    config.serverConfigPath = serverConfigPath
    config.serverResolvPath = serverResolvPath

    addrc.syncAddressUpdate = addrc.SyncEndpointUpdate
    addrc.enqueueAddressUpdate = addrc.enqueue

    //Add event handlers
    addressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
        AddFunc:    addrc.addAddress,
        UpdateFunc: addrc.updateAddress,
        DeleteFunc: addrc.deleteAddress,
    })
    addrc.addressLister = addressInformer.Lister()
    addrc.addressListerSynced = addressInformer.Informer().HasSynced

    return addrc
}

func (addrc *Controller) enqueue(sysb *crv1.Endpoint) {
    key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sysb)
    if err != nil {
        runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", sysb, err))
        return
    }

    addrc.queue.Add(key)
}

func (addrc *Controller) addAddress(obj interface{}) {
    // New address resource has arrived
    glog.V(1).Infof("MyController just got an add")
    address := obj.(*crv1.Endpoint)

    addrc.enqueueAddressUpdate(address)
}

func (addrc *Controller) updateAddress(old, cur interface{}) {
    // Address object has been modified
    glog.V(1).Infof("MyController just got an update")
    address := cur.(*crv1.Endpoint)

    addrc.enqueueAddressUpdate(address)
}

func (addrc *Controller) deleteAddress(obj interface{}) {
    // Address object has been modified
    glog.V(1).Infof("MyController just got a delete")
    address := obj.(*crv1.Endpoint)

    addrc.enqueueAddressUpdate(address)
}

func (addrc *Controller) Run(workers int, stopCh <-chan struct{}) {
    // don't let panics crash the process
    defer runtime.HandleCrash()
    // make sure the work queue is shutdown which will trigger workers to end
    defer addrc.queue.ShutDown()

    glog.Infof(" Warning :: every endpoint is actually a service build")
    glog.Infof("Starting local-dns controller")
    defer glog.Infof("Shutting down local-dns controller")

    // wait for your secondary caches to fill before starting your work.
    // Do we need this if just using this lister?
    if !cache.WaitForCacheSync(stopCh, addrc.addressListerSynced) {
        return
    }

    glog.V(4).Info("Caches synced. Waiting for config to be set")

    // wait for config to be set.
    // Is this necessary for our case
    // <-addrc.configSetChan

    // start up your worker threads based on threadiness.  Some controllers
    // have multiple kinds of workers
    for i := 0; i < workers; i++ {
        // runWorker will loop until "something bad" happens.  The .Until will
        // then rekick the worker after one second
        go wait.Until(addrc.runWorker, time.Second, stopCh)
    }

    // wait until we're told to stop
    <-stopCh
}

func (addrc *Controller) runWorker() {
    // hot loop until we're told to stop.  processNextWorkItem will
    // automatically wait until there's work available, so we don't worry
    // about secondary waits
    for addrc.processNextWorkItem() {
    }
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (addrc *Controller) processNextWorkItem() bool {
    // pull the next work item from queue.  It should be a key we use to lookup
    // something in a cache
    key, quit := addrc.queue.Get()
    if quit {
        return false
    }
    // you always have to indicate to the queue that you've completed a piece of
    // work
    defer addrc.queue.Done(key)

    // do your work on the key.  This method will contains your "do stuff" logic
    err := addrc.syncAddressUpdate(key.(string))
    if err == nil {
        // if you had no error, tell the queue to stop tracking history for your
        // key. This will reset things like failure counts for per-item rate
        // limiting
        addrc.queue.Forget(key)
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
    addrc.queue.AddRateLimited(key)

    return true
}

func (addrc *Controller) SyncEndpointUpdate(key string) error {
    // List from the informer given in the controller.
    glog.V(1).Infof("Called rewrite DNS")
    defer func() {
        glog.V(4).Infof("Finished rewrite DNS")
    }()

    namespace, name, err := cache.SplitMetaNamespaceKey(key)
    if err != nil {
        return err
    }

    //Update our local list with just the new item, not the whole list
    endpoint, err := addrc.addressLister.Endpoints(namespace).Get(name)

    if errors.IsNotFound(err) {
        glog.V(2).Infof("Endpoint %v has been deleted", key)
        return nil
    }

    if err != nil {
        return err
    }

    glog.V(5).Infof("ServiceBuild %v state: %v", key, endpoint.Status.State)

    // If not recently updated, become responsible for flushing
    addrc.lock.RLock()
    if !addrc.hasRecentlyUpdated {
        addrc.lock.Lock()

        if !addrc.hasRecentlyUpdated {
            // Safe to write to boolean, and become responsible for updating
            addrc.hasRecentlyUpdated = true
            go time.AfterFunc(time.Second * time.Duration(updateWaitBeforeFlushTimer), addrc.FlushRewriteDNS)
        }
    }

    // Acquire the write lock to try and update the map
    addrc.lock.Lock()

    endpointPathURL := endpoint.Spec.Path.ToDomain(true)

    // TODO :: handle delete
    addrc.cnameList[endpointPathURL] = *endpoint.Spec.IP

    addrc.FlushRewriteDNS()

    return nil
}

func (addrc *Controller) FlushRewriteDNS() {
    // Called when it is time to actually rewrite the dns files.

    // Should be two separate go routines.
    err := addrc.RewriteDnsmasqConfig()

    if err != nil {
        println(err)
    }

    // No ping should be necessary given auto update.
    // However sending a SIGHUP would automatically reload resolv if necessary.
}

func CreateEmptyFile(filename string) (*os.File, error) {

    _, err := os.Stat(filename)

    if os.IsExist(err) {
        err := os.Remove(filename)

        if err != nil {
            panic(err)
        }
    }

    return os.OpenFile(filename, os.O_RDWR | os.O_CREATE, 0660)
}

func (addrc *Controller) RewriteDnsmasqConfig() error {

    fmt.Println("Rewriting config...")
    // Open dnsmasq.extra.conf and rewrite
    cname_file, err := CreateEmptyFile(config.serverConfigPath)

    defer func() {
        fmt.Println("Closing config file...")
        cname_file.Close()

        // Finished writing to the cache - can now unset the timer flag
        addrc.lock.Lock()
        addrc.hasRecentlyUpdated = false
        addrc.lock.Unlock()
    }()

    if err != nil {
        return err
    }

    // This is an extra config file, so contains only the options which must be rewritten.
    // Condition on cname is that it exists in the specified host file.
    // Each cname entry of the form cname=ALIAS,...(addn alias),TARGET

    for k, v := range addrc.cnameList {
        _, err := cname_file.WriteString("cname=" + k + "," + v + "\n")

        if err != nil {
            return err
        }
    }

    return nil
}
