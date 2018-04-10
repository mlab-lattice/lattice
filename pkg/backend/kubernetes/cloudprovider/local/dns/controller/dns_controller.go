package controller

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	endpointutil "github.com/mlab-lattice/lattice/pkg/util/endpoint"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	clientset "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	set "github.com/deckarep/golang-set"
	"github.com/golang/glog"
)

type Controller struct {
	externalNameEndpoints set.Set
	ipEndpoints           set.Set

	latticeClient latticeclientset.Interface

	endpointister        latticelisters.EndpointLister
	endpointListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	dnsmasqConfigPath string
	hostFilePath      string
	latticeID         v1.LatticeID
}

var (
	updateWaitBeforeFlushTimerSeconds = 5
)

// NewController returns a newly created DNS Controller.
func NewController(
	dnsmasqConfigPath string,
	hostConfigPath string,
	latticeID v1.LatticeID,
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

	c.dnsmasqConfigPath = dnsmasqConfigPath
	c.hostFilePath = hostConfigPath
	c.latticeID = latticeID

	c.endpointister = endpointInformer.Lister()
	c.endpointListerSynced = endpointInformer.Informer().HasSynced

	c.externalNameEndpoints = set.NewSet()
	c.ipEndpoints = set.NewSet()

	return c
}

// Run triggers a goroutine to refresh the cache and rewrite the config at set intervals, until it receives from stopCh
func (c *Controller) Run(stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Infof("Running controller")
	defer glog.Infof("Shutting down controller")

	// wait for your secondary caches to fill before starting your work.
	if !cache.WaitForCacheSync(stopCh, c.endpointListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced.")

	go wait.Until(c.updateConfigs, time.Second*time.Duration(updateWaitBeforeFlushTimerSeconds), stopCh)

	// wait until we're told to stop
	<-stopCh
}

// updateConfigs runs at regular intervals and compares the endpoints the configuration was written with against
// the current list of endpoints. If there is any difference, the configuration is rewritten.
func (c *Controller) updateConfigs() {
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
	if externalNameEndpointsSet.SymmetricDifference(c.externalNameEndpoints).Cardinality() > 0 {
		updateNeeded = true
	}

	if ipEndpointsSet.SymmetricDifference(c.ipEndpoints).Cardinality() > 0 {
		updateNeeded = true
	}

	if !updateNeeded {
		// In the case the previous update failed and some endpoints have not been updated to the created state
		// attempt to update them again here.
		for _, e := range endpoints {
			if e.Status.State != latticev1.EndpointStateCreated {
				endpoint := e.DeepCopy()
				endpoint.Status.State = latticev1.EndpointStateCreated

				_, err = c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)
				if err != nil {
					runtime.HandleError(err)
				}
			}
		}

		return
	}

	glog.V(5).Infof("Endpoints have changed, rewriting DNS configuration...")

	err = c.rewriteDnsmasqConfig(endpoints)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	c.externalNameEndpoints = externalNameEndpointsSet
	c.ipEndpoints = ipEndpointsSet

	for _, e := range endpoints {
		if e.Status.State == latticev1.EndpointStateCreated {
			continue
		}

		endpoint := e.DeepCopy()
		endpoint.Status.State = latticev1.EndpointStateCreated

		_, err := c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)
		if err != nil {
			runtime.HandleError(err)
		}
	}
}

// endpointsSets returns two sets of endpoints - the first for endpoints specified by an external name,
// and the second for endpoints specified by an ip.
func (c *Controller) endpointsSets(endpoints []*latticev1.Endpoint) (set.Set, set.Set, error) {
	externalNameEndpoints := set.NewSet()
	ipEndpoints := set.NewSet()

	for _, endpoint := range endpoints {
		key := fmt.Sprintf("%v/%v", endpoint.Namespace, endpoint.Name)
		if endpoint.Spec.ExternalName != nil {
			endpointKey := fmt.Sprintf("/%v" + *endpoint.Spec.ExternalName)
			externalNameEndpoints.Add(key + endpointKey)
			continue
		}

		if endpoint.Spec.IP != nil {
			ipKey := fmt.Sprintf("/%v" + *endpoint.Spec.IP)
			ipEndpoints.Add(key + ipKey)
			continue
		}

		return nil, nil, fmt.Errorf("endpoint %v had neither ExternalName nor IP set", key)
	}

	return externalNameEndpoints, ipEndpoints, nil
}

// rewriteDnsmasqConfig rewrites the config files for the dns server
func (c *Controller) rewriteDnsmasqConfig(endpoints []*latticev1.Endpoint) error {
	// This is an extra config file, so contains only the options which must be rewritten.
	// Condition on cname is that it exists in the specified host file, or references another cname.
	// Each cname entry of the form cname=ALIAS,...(addn aliases),TARGET
	// Full specification here: http://www.thekelleys.org.uk/dnsmasq/docs/dnsmasq-man.html
	dnsmasqConfigFileContents := ""
	hostConfigFileContents := ""

	for _, endpoint := range endpoints {
		systemID, err := kubeutil.SystemID(endpoint.Namespace)
		if err != nil {
			return err
		}

		domain := endpoint.Spec.Path.ToDomain()
		cname := endpointutil.DNSName(domain, systemID, c.latticeID)

		if endpoint.Spec.IP != nil {
			hostConfigFileContents += fmt.Sprintf("%v %v\n", *endpoint.Spec.IP, cname)
		}

		if endpoint.Spec.ExternalName != nil {
			dnsmasqConfigFileContents += fmt.Sprintf("cname=%v,%v\n", cname, *endpoint.Spec.ExternalName)
		}
	}

	err := ioutil.WriteFile(c.dnsmasqConfigPath, []byte(dnsmasqConfigFileContents), 0660)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.hostFilePath, []byte(hostConfigFileContents), 0660)
	if err != nil {
		return err
	}

	return nil
}
