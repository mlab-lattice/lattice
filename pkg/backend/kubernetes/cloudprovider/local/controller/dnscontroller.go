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
)

type Controller struct {
	previousEndpoints []*crv1.Endpoint
	currentEndpoints  []*crv1.Endpoint

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

	c.currentEndpoints = endpoints

	if err != nil {
		runtime.HandleError(err)
		return
	}

	if !haveEndpointsChanged(c.currentEndpoints, c.previousEndpoints) {
		return
	}

	glog.V(5).Infof("Endpoints have changed, rewriting DNS configuration...")
	err = c.RewriteDnsmasqConfig()

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

// haveEndpointsChanged returns true if the two lists of endpoints are differnt
func haveEndpointsChanged(endpointListA []*crv1.Endpoint, endpointListB []*crv1.Endpoint) bool {

	endpointMapA := make(map[string]crv1.Endpoint)
	endpointMapB := make(map[string]crv1.Endpoint)

	allEndpoints := set.NewSet()

	for _, endpoint := range endpointListA {
		key := endpoint.Name
		allEndpoints.Add(key)
		endpointMapA[key] = *endpoint

		ip := ""
		if endpoint.Spec.IP != nil {
			ip = *endpoint.Spec.IP
		}

		name := ""
		if endpoint.Spec.ExternalName != nil {
			name = *endpoint.Spec.ExternalName
		}

		glog.V(1).Infof("Endpoint before: path = %v, ip = %v, external name = %v", endpoint.Spec.Path.ToDomain(true), ip, name)
	}

	for _, endpoint := range endpointListB {
		key := endpoint.Name
		allEndpoints.Add(key)
		endpointMapB[key] = *endpoint

		ip := ""
		if endpoint.Spec.IP != nil {
			ip = *endpoint.Spec.IP
		}

		name := ""
		if endpoint.Spec.ExternalName != nil {
			name = *endpoint.Spec.ExternalName
		}

		glog.V(1).Infof("Endpoint before: path = %v, ip = %v, external name = %v", endpoint.Spec.Path.ToDomain(true), ip, name)
	}

	if len(endpointListA) == 0 && len(endpointListB) == 0 {
		return false
	}

	itr := allEndpoints.Iterator()

	for k := range itr.C {
		str := k.(string)

		endpointA, ok := endpointMapA[str]

		if !ok {
			return true
		}

		endpointB, ok := endpointMapB[str]

		if !ok {
			return true
		}

		// Simple equality based on file output or endpoint state.
		if endpointA.Spec.Path != endpointB.Spec.Path ||
			endpointA.Spec.IP != endpointB.Spec.IP ||
			endpointA.Spec.ExternalName != endpointB.Spec.ExternalName ||
			endpointA.Namespace != endpointB.Namespace ||
			endpointA.Status != endpointB.Status {
			return true
		}
	}

	return false
}

// CreateEmptyFile creates an empty file, removing the previous file if it existed
func CreateEmptyFile(filename string) (*os.File, error) {

	_, err := os.Stat(filename)

	if err == nil {
		err := os.Remove(filename)

		if err != nil {
			runtime.HandleError(err)
		}
	}

	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0660)
}

// RewriteDnsmasqConfig rewrites the config files for the dns server
func (c *Controller) RewriteDnsmasqConfig() error {

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
		cname := v.Spec.ExternalName
		endpoint := v.Spec.IP
		systemID, err := util.SystemID(v.Namespace)

		if err != nil {
			runtime.HandleError(err)
		}

		path := v.Spec.Path.ToDomain(true)
		path = path + ".local." + c.clusterID + "." + string(systemID) + ".local"

		if endpoint != nil && cname != nil {
			return fmt.Errorf("endpoint %v has both a cname and an endpoint", path)
		}

		if cname != nil {
			_, err = dnsmasqConfigFile.WriteString("cname=" + path + "," + *cname + "\n")
			glog.V(5).Infof("cname=" + path + "," + *cname + "\n")
		}

		if endpoint != nil {
			_, err = hostsFile.WriteString(*endpoint + " " + path + "\n")
			glog.V(5).Infof(*endpoint + " " + path + "\n")
		}

		if err != nil {
			return err
		}

		//c.eventRecorder.Event(v, "Normal", "DnsRewrite", "The endpoint was added.")
	}

	return nil
}

// refreshOldCache is called after rewriting the config, and updates the cached view of the in memory cname/host files.
func (c *Controller) refreshOldCache() {
	c.previousEndpoints = c.currentEndpoints
	c.currentEndpoints = nil
}
