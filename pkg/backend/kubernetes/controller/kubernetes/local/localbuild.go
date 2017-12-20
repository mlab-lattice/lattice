package local

import (
	"fmt"
	"time"
	"math/rand"
	"os"

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
	enqueueAddressUpdate func(sysBuild *crv1.ServiceBuild)

	ips 	[]IP
	cnames 	[]CName

	latticeClient latticeclientset.Interface

	addressLister latticelisters.ServiceBuildLister
	addressListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

type ServerConfig struct {
	serverConfigPath string
	serverResolvPath string
}

type CName struct {
	alias 	  string
	canonical string
}

type IP struct {
	ip string
}

var (
	config ServerConfig

	resolvOptions = []string {
		"ndots:15",
	}
)

func NewController(
	serverConfigPath string,
	serverResolvPath string,
	latticeClient  	 latticeclientset.Interface,
	addressInformer  latticeinformers.ServiceBuildInformer,
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

func (addrc *Controller) enqueue(sysb *crv1.ServiceBuild) {
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
	address := obj.(*crv1.ServiceBuild)

	addrc.enqueueAddressUpdate(address)
}

func (addrc *Controller) updateAddress(old, cur interface{}) {
	// Address object has been modified
	glog.V(1).Infof("MyController just got an update")
	address := cur.(*crv1.ServiceBuild)

	addrc.enqueueAddressUpdate(address)
}

func (addrc *Controller) deleteAddress(obj interface{}) {
	// Address object has been modified
	glog.V(1).Infof("MyController just got a delete")
	address := obj.(*crv1.ServiceBuild)

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

	// Work with the cache here - get everything.
	// lister, err := addrc.addressLister.List(labels.Everything())
	// One at a time method is more suitable
	endpoint, err := addrc.addressLister.ServiceBuilds(namespace).Get(name)

	if errors.IsNotFound(err) {
		glog.V(2).Infof("Endpoint %v has been deleted", key)
		return nil
	}

	if err != nil {
		return err
	}

	// What would the endpoint variation of this be
	//stateInfo, err := addrc.calculateState(endpoint)
    //
	//if err != nil {
	//	return err
	//}

	//glog.V(5).Infof("ServiceBuild %v state: %v", key, stateInfo.state)

	// For now, no dealings with the switch b state
    //
	//switch stateInfo.state {
	//case stateHasFailedComponentBuilds:
	//	return c.syncFailedServiceBuild(build, stateInfo)
	//case stateHasOnlyRunningOrSucceededComponentBuilds:
	//	return c.syncRunningServiceBuild(build, stateInfo)
	//case stateNoFailuresNeedsNewComponentBuilds:
	//	return c.syncMissingComponentBuildsServiceBuild(build, stateInfo)
	//case stateAllComponentBuildsSucceeded:
	//	return c.syncSucceededServiceBuild(build, stateInfo)
	//default:
	//	return fmt.Errorf("ServiceBuild %v in unexpected state %v", key, stateInfo.state)
	//}

	/// TODO :: Switch based on EndpointSpec. Check cant be both.
	switch rand.Int() {
	case 1:
		// Create a cname
		cname := CName{
			canonical: "my_canonical",
			alias: "my_alias",
		}

		addrc.cnames = append(addrc.cnames, cname)
	case 2:
		// Create an IP
		ip := IP{
			ip: "192.168.9.9",
		}

		addrc.ips = append(addrc.ips, ip)
	}

	//Sync functions will include an update which will be a cache safe method to update the state.
	// TODO :: Implement sync / update
	addrc.FlushRewriteDNS()

	return nil
}

func (addrc *Controller) FlushRewriteDNS() {
	// Called when it is time to actually rewrite the dns files.

	// Should be two separate go routines.
	err := addrc.RewriteDnsmasqConfig()

	if err != nil {
		panic(err)
	}

	err = addrc.RewriteResolvConf()

	if err != nil {
		panic(err)
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
	// Open dnsmasq.extra.conf and rewrite
	cname_file, err := CreateEmptyFile(config.serverConfigPath)

	defer cname_file.Close()

	if err != nil {
		panic(err)
	}

	// This is an extra config file, so contains only the options which must be rewritten.
	// Condition on cname is that it exists in the specified host file.
	//Each cname entry of the form cname=ALIAS,...(addn alias),TARGET

	for _, v := range addrc.cnames {
		_, err := cname_file.WriteString("cname=" + v.alias + "," + v.canonical + "\n")

		if err != nil {
			panic(err)
		}
	}
}

func (addrc *Controller) RewriteResolvConf() error {

	// Open dnsmasq.resolv.conf and rewrite
	resolv_file, err := CreateEmptyFile(config.serverResolvPath)

	defer resolv_file.Close()

	if (err != nil) {
		panic(err)
	}

	// nameserver a
	// nameserver b

	// search a b c

	// Write any options

	resolv_file.WriteString("options")

	for _, v := range resolvOptions {
		_, err := resolv_file.WriteString(" " + v)

		if err != nil {
			panic(err)
		}
	}
}
