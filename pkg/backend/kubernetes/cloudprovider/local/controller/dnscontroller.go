package dnscontroller

import (
	"fmt"
	"os"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	util "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	"k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	set "github.com/deckarep/golang-set"
	"github.com/golang/glog"
	"io/ioutil"
)

type Controller struct {
	externalNameEndpoints set.Set
	ipEndpoints           set.Set

	latticeClient latticeclientset.Interface

	endpointister        latticelisters.EndpointLister
	endpointListerSynced cache.InformerSynced

	eventRecorder record.EventRecorder

	queue workqueue.RateLimitingInterface

	serverConfigPath string
	hostConfigPath   string
	clusterID        string
}

var (
	updateWaitBeforeFlushTimerSeconds = 15
)

// NewController returns a newly created DNS Controller.
func NewController(
	serverConfigPath string,
	hostConfigPath string,
	clusterID string,
	latticeClient latticeclientset.Interface,
	client clientset.Interface,
	endpointInformer latticeinformers.EndpointInformer,
) *Controller {

	c := &Controller{
		latticeClient: latticeClient,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "system"),
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(client.CoreV1().RESTClient()).Events("")})

	c.eventRecorder = eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "deployment-controller"})

	c.serverConfigPath = serverConfigPath
	c.hostConfigPath = hostConfigPath
	c.clusterID = clusterID

	c.endpointister = endpointInformer.Lister()
	c.endpointListerSynced = endpointInformer.Informer().HasSynced

	return c
}

// Run triggers a goroutine to refresh the cache and rewrite the config at set intervals, until it receives from stopCh
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
	if err != nil {
		runtime.HandleError(err)
		return
	}

	externalNameEndpointsSet, ipEndpointsSet, err := c.endpointsSets(endpoints)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	updateNeeded := false
	if externalNameEndpointsSet.Difference(c.externalNameEndpoints).Cardinality() > 0 {
		updateNeeded = true
	}

	if ipEndpointsSet.Difference(c.ipEndpoints).Cardinality() > 0 {
		updateNeeded = true
	}

	if !updateNeeded {
		return
	}

	glog.V(5).Infof("Endpoints have changed, rewriting DNS configuration...")

	err = c.RewriteDnsmasqConfig(externalNameEndpointsSet, ipEndpointsSet)
	if err != nil {
		runtime.HandleError(err)
	}

	for _, endpoint := range c.currentEndpoints {

		if endpoint.Status.State == crv1.EndpointStateCreated {
			continue
		}

		endpoint = endpoint.DeepCopy()
		endpoint.Status.State = crv1.EndpointStateCreated

		_, err := c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)

		if err != nil {
			runtime.HandleError(err)
		}
	}

	c.refreshOldCache()

}

// endpointsSets returns true if the two lists of endpoints are differnt
func (c *Controller) endpointsSets(endpoints []*crv1.Endpoint) (set.Set, set.Set, error) {
	externalNameEndpoints := set.NewSet()
	ipEndpoints := set.NewSet()

	for _, endpoint := range endpoints {
		key := fmt.Sprintf("%v/%v", endpoint.Namespace, endpoint.Name)
		if endpoint.Spec.ExternalName != nil {
			externalNameEndpoints.Add(key)
			continue
		}

		if endpoint.Spec.IP != nil {
			ipEndpoints.Add(key)
			continue
		}

		return nil, nil, fmt.Errorf("Endpoint %v had neither ExternalName nor IP set", key)
	}

	return externalNameEndpoints, ipEndpoints, nil
}

// CreateEmptyFile creates an empty file, removing the previous file if it existed
func CreateEmptyFile(filename string) (*os.File, error) {
	_, err := os.Stat(filename)

	// TODO: document this
	if err == nil {
		err := os.Remove(filename)

		if err != nil {
			return nil, err
		}
	}

	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0660)
}

// RewriteDnsmasqConfig rewrites the config files for the dns server
func (c *Controller) RewriteDnsmasqConfig(externalNameEndpointsSet, ipEndpointsSet set.Set) error {

	glog.V(4).Infof("Rewriting config %e, %e... ", c.hostConfigPath, c.serverConfigPath)

	//// FIXME: use https://golang.org/pkg/io/ioutil/#WriteFile
	//dnsmasqConfigFile, err := CreateEmptyFile(c.serverConfigPath)
	//defer dnsmasqConfigFile.Sync()
	//defer dnsmasqConfigFile.Close()

	//if err != nil {
	//	return err
	//}

	// FIXME: use https://golang.org/pkg/io/ioutil/#WriteFile
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

	// FIXME: we were here
	dnsmasqConfigFileContents := ""
	for setInterface := range externalNameEndpointsSet.Iter() {
		endpoint, ok := setInterface.(*crv1.Endpoint)
		if !ok {
			return fmt.Errorf("value in externalNameEndpointsSet was not a *crv1.Endpoint")
		}

		systemID, err := util.SystemID(endpoint.Namespace)
		if err != nil {
			return err
		}

		path := endpoint.Spec.Path.ToDomain(true)
		cname := fmt.Sprintf("%v.local.%v.%v.local", path, c.clusterID, systemID)

		dnsmasqConfigFileContents += fmt.Sprintf("cname=%v\n", cname)
	}

	err = ioutil.WriteFile(c.serverConfigPath, []byte(dnsmasqConfigFileContents), 0660)
	if err != nil {
		return err
	}

	for _, e := range c.currentEndpoints {
		cname := e.Spec.ExternalName
		endpoint := e.Spec.IP

		if err != nil {
			runtime.HandleError(err)
		}

		path := e.Spec.Path.ToDomain(true)
		path = path + ".local." + c.clusterID + "." + string(systemID) + ".local"

		if endpoint != nil && cname != nil {
			return fmt.Errorf("endpoint %e has both a cname and an endpoint", path)
		}

		if cname != nil {
			_, err = dnsmasqConfigFile.WriteString("cname=" + path + "," + *cname + "\n")
			glog.V(5).Infof("Added: cname=" + path + "," + *cname + "\n")
		}

		if endpoint != nil {
			_, err = hostsFile.WriteString(*endpoint + " " + path + "\n")
			glog.V(5).Infof("Added: " + *endpoint + " " + path + "\n")
		}

		if err != nil {
			return err
		}

		// c.eventRecorder.Event(c, "Normal", "DnsRewrite", "The endpoint was added.")
	}

	return nil
}

// refreshOldCache is called after rewriting the config, and updates the cached view of the in memory cname/host files.
func (c *Controller) refreshOldCache() {
	c.previousEndpoints = c.currentEndpoints
	c.currentEndpoints = nil
}

//func (c *Controller) GetObjectKind() schema.ObjectKind {
//	return schema.EmptyObjectKind
//}
//
//func (c *Controller) DeepCopyObject() rtime.Object {
//	return c
//}
