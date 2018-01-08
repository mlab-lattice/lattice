package dnscontroller

import (
	"fmt"
	"os"
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

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
}

var (
	updateWaitBeforeFlushTimerSeconds = 15
)

func NewController(
	serverConfigPath string,
	hostConfigPath string,
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

	c.endpointister = endpointInformer.Lister()
	c.endpointListerSynced = endpointInformer.Informer().HasSynced

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

	c.eventRecorder.Event(nil, "Normal", "DnsRewrite", "The DNS Server was rewritten")

	if err != nil {
		runtime.HandleError(err)
	}

	for _, endpoint := range c.currentEndpoints {
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

	keys := set.NewSet()

	for k, _ := range endpointListA {
		keys.Add(k)
	}

	for k, _ := range endpointListB {
		keys.Add(k)
	}

	if len(endpointListA) == 0 && len(endpointListB) == 0 {
		return false
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
