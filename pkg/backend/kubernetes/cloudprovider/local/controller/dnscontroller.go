package dnscontroller

import (
	"fmt"
	"os"
	"sync"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/apimachinery/pkg/labels"

	set "github.com/deckarep/golang-set"
	"github.com/golang/glog"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
)

type Controller struct {
	previousEndpoints []*crv1.Endpoint
	currentEndpoints  []*crv1.Endpoint

	flushPending bool

	sharedVarsLock sync.RWMutex

	latticeClient latticeclientset.Interface

	endpointister        latticelisters.EndpointLister
	endpointListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	serverConfigPath string
	hostConfigPath   string
}

var (
	updateWaitBeforeFlushTimerSeconds = 15
)

func NewController(
	serverConfigPath string,
	hostConfigPath string,
	latticeClient latticeclientset.Interface,
	endpointInformer latticeinformers.EndpointInformer,
) *Controller {

	c := &Controller{
		latticeClient: latticeClient,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system"),
	}

	c.serverConfigPath = serverConfigPath
	c.hostConfigPath = hostConfigPath

	c.endpointister = endpointInformer.Lister()
	c.endpointListerSynced = endpointInformer.Informer().HasSynced

	c.cnamesCached = make(map[string]crv1.Endpoint)
	c.hostsCached = make(map[string]crv1.Endpoint)
	c.cnamesCurrent = make(map[string]crv1.Endpoint)
	c.hostsCurrent = make(map[string]crv1.Endpoint)

	return c
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Starting local-dns controller")
	defer glog.Infof("Shutting down local-dns controller")

	// wait for your secondary caches to fill before starting your work.
	if !cache.WaitForCacheSync(stopCh, c.endpointListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Waiting for config to be set")

	go wait.Until(c.calculateCache, time.Second*time.Duration(updateWaitBeforeFlushTimerSeconds), stopCh)

	// wait until we're told to stop
	<-stopCh
}

// calculateCache runs at regular intervals and compares the current cache to the current set of endpoints. If the current cache is not up to date, it is rewritten
func (c *Controller) calculateCache() {

	endpoints, err := c.endpointister.List(labels.Everything())

	c.currentEndpoints = endpoints

	if err != nil {
		runtime.HandleError(err)
		return
	}

	if !haveEndpointsChanged(c.currentEndpoints, c.previousEndpoints) {
		return
	}

	err = c.RewriteDnsmasqConfig()

	if err != nil {
		runtime.HandleError(err)
	}

}

// haveEndpointsChanged returns true if the two lists of endpoints are differnt
func haveEndpointsChanged(endpointListA []*crv1.Endpoint, endpointListB []*crv1.Endpoint) bool {

	endpointMapA := make(map[string]crv1.Endpoint)
	endpointMapB := make(map[string]crv1.Endpoint)

	keys := set.NewSet()

	for k, _ := range endpointMapA {
		keys.Add(k)
	}

	for k, _ := range endpointMapB {
		keys.Add(k)
	}

	itr := keys.Iterator()

	for k := range itr.C {
		str := k.(string)

		endpointA, ok := endpointMapA[str]

		if !ok {
			return false
		}

		endpointB, ok := endpointMapB[str]

		if !ok {
			return false
		}

		// Simple equality based on just what the file output depends on.
		if endpointA.Spec.Path != endpointB.Spec.Path {
			return false
		}

		if endpointA.Spec.IP != endpointB.Spec.IP {
			return false
		}

		if endpointA.Spec.ExternalName != endpointB.Spec.ExternalName {
			return false
		}
	}

	return true
}

// updateEndpointValue takes one Endpoint key and updates the current cache to represent the up to date version of this endpoint
func (c *Controller) updateEndpointValue(key string) error {
	glog.V(1).Infof("updating endpoint %v", key)
	defer func() {
		glog.V(4).Infof("Finished endpoint update")
	}()
	namespace, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		return err
	}

	// Cache lookup
	endpoint, err := c.endpointlister.Endpoints(namespace).Get(name)
	endpointPathURL := endpoint.Spec.Path.ToDomain(true)

	if errors.IsNotFound(err) {
		// Should probably just remove here
	}

	if endpoint.DeletionTimestamp != nil {
		return c.syncHandleFinalizer(key, endpoint, endpointPathURL)
	}

	if err != nil {
		return err
	}

	glog.V(5).Infof("Endpoint %v state: %v", key, endpoint.Status.State)

	if endpoint.Status.State == crv1.EndpointStateCreated {
		// Created, nothing to do.
		return nil
	}

	// Locks sharedVars for the entire duration. This ensures that the hosts and cnames are updated atomically alongside
	// cache flushes and prevents missed updates.
	c.sharedVarsLock.Lock()
	defer c.sharedVarsLock.Unlock()

	if !c.flushPending {
		glog.V(5).Infof("has not updated recently, will flush all updates in %v seconds", updateWaitBeforeFlushTimerSeconds)
		// Safe to write to this boolean as we have the write sharedVarsLock.
		c.flushPending = true
		go time.AfterFunc(time.Second*time.Duration(updateWaitBeforeFlushTimerSeconds), c.FlushRewriteDNS)
	}

	_, inHosts := c.hosts[key]
	_, inCname := c.cnames[key]

	if inHosts || inCname {
		glog.V(5).Infof("Endpoint %v already updated. Setting state to created...", key)

		if inCname {
			delete(c.cnames, endpointPathURL)
		}

		if inHosts {
			delete(c.hosts, endpointPathURL)
		}

		endpoint = endpoint.DeepCopy()
		endpoint.Status.State = crv1.EndpointStateCreated
		_, err := c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)

		return err
	}

	if endpoint.Spec.ExternalName != nil {
		return c.syncExternalEndpoint(key, endpoint)
	}

	if endpoint.Spec.IP != nil {
		return c.syncIPEndpoint(key, endpoint)
	}

	return fmt.Errorf("endpoint %v/%v does not have External or IP set", endpoint.Namespace, endpoint.Name)
}

func (c *Controller) syncExternalEndpoint(key string, endpoint *crv1.Endpoint) error {
	endpointPathURL := endpoint.Spec.Path.ToDomain(true)

	if _, ok := c.hosts[key]; ok {
		delete(c.hosts, key)
	}

	glog.V(2).Infof("Updating endpoint %v with cname %v...", endpointPathURL, *endpoint.Spec.ExternalName)
	c.cnames[key] = *endpoint

	return nil
}

func (c *Controller) syncIPEndpoint(key string, endpoint *crv1.Endpoint) error {
	endpointPathURL := endpoint.Spec.Path.ToDomain(true)

	if _, ok := c.cnames[key]; ok {
		delete(c.cnames, key)
	}

	glog.V(2).Infof("Updating endpoint %v with IP address %v...", endpointPathURL, *endpoint.Spec.IP)
	c.hosts[key] = *endpoint

	return nil
}

// syncHandleFinalizer is called when the controller needs to handle any finalization logic and remove its finalizer.
func (c *Controller) syncHandleFinalizer(key string, endpoint *crv1.Endpoint, endpointPathURL string) error {

	foundFinalizer := false
	finalizerIndex := -1

	for idx, finalizer := range endpoint.Finalizers {
		if finalizer == local.LocaldnsFinalizer {
			foundFinalizer = true
			finalizerIndex = idx
			break
		}
	}

	if !foundFinalizer {
		return fmt.Errorf("could not find expected finalizer %v for endpoint %v", local.LocaldnsFinalizer, endpoint.Name)
	}

	c.sharedVarsLock.Lock()
	defer c.sharedVarsLock.Unlock()

	_, inHosts := c.hosts[key]
	_, inCname := c.cnames[key]

	if inCname {
		delete(c.cnames, endpointPathURL)
	}

	if inHosts {
		delete(c.hosts, endpointPathURL)
	}

	endpoint = endpoint.DeepCopy()
	endpoint.Finalizers = append(endpoint.Finalizers[:finalizerIndex], endpoint.Finalizers[finalizerIndex+1:]...)
	_, err := c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)

	if err != nil {
		return err
	}

	// Can stop tracking once its finalizer is dealt with.
	return nil
}

func (c *Controller) FlushRewriteDNS() {
	err := c.RewriteDnsmasqConfig()

	if err != nil {
		runtime.HandleError(err)
	}
}

func CreateEmptyFile(filename string) (*os.File, error) {

	_, err := os.Stat(filename)

	if os.IsExist(err) {
		err := os.Remove(filename)

		if err != nil {
			runtime.HandleError(err)
		}
	}

	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0660)
}

func (c *Controller) RewriteDnsmasqConfig() error {

	// This logic takes a write lock for the entire duration of the update to simplify the logic and to prevent possible missed updates.
	c.sharedVarsLock.Lock()
	defer c.sharedVarsLock.Unlock()
	defer func() {
		// Finished writing to the cache - can now unset the timer flag
		c.flushPending = false
	}()

	glog.V(4).Infof("Rewriting config %v, %v... ", c.hostConfigPath, c.serverConfigPath)

	dnsmasqConfigFile, err := CreateEmptyFile(c.serverConfigPath)
	defer dnsmasqConfigFile.Sync()
	defer dnsmasqConfigFile.Close()

	if err != nil {
		return err
	}

	hostsFile, err := CreateEmptyFile(c.hostConfigPath)
	defer hostsFile.Sync()
	defer hostsFile.Close()

	if err != nil {
		return err
	}

	// This is an extra config file, so contains only the options which must be rewritten.
	// Condition on cname is that it exists in the specified host file, or references another cname.
	// Each cname entry of the form cname=ALIAS,...(addn aliases),TARGET
	// Full specification here: http://www.thekelleys.org.uk/dnsmasq/docs/dnsmasq-man.html

	for _, v := range c.currentEndpoints {
		cname := *v.Spec.ExternalName
		endpoint := *v.Spec.IP
		path := v.Spec.Path.ToDomain(true)

		if endpoint != "" && cname != "" {
			return fmt.Errorf("endpoint %v has both a cname and an endpoint", path)
		}

		if cname != "" {
			_, err = dnsmasqConfigFile.WriteString("cname=" + path + "," + cname + "\n")
			glog.V(5).Infof("cname=" + path + "," + cname + "\n")
		}

		if endpoint != "" {
			_, err = hostsFile.WriteString(endpoint + " " + path + "\n")
			glog.V(5).Infof(endpoint + " " + path + "\n")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// refreshOldCache is called after rewriting the config, and updates the cached view of the in memory cname/host files.
func (c *Controller) refreshOldCache() {
	c.previousEndpoints = c.currentEndpoints
	c.currentEndpoints = nil
}
