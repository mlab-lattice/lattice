package controller

import (
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	kubeclientset "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	set "github.com/deckarep/golang-set"
	"github.com/golang/glog"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type Controller struct {
	externalNameAddresses set.Set
	serviceAddresses      set.Set

	latticeID       v1.LatticeID
	namespacePrefix string

	internalDNSDomain string

	dnsmasqConfigPath    string
	dnsmasqHostsFilePath string

	// NOTE: you must get a read lock on the configLock for the duration
	//       of your use of the serviceMesh
	staticServiceMeshOptions *servicemesh.Options
	serviceMesh              servicemesh.Interface

	latticeClient latticeclientset.Interface

	configLister       latticelisters.ConfigLister
	configListerSynced cache.InformerSynced
	configSetChan      chan struct{}
	configSet          bool
	configLock         sync.RWMutex
	config             latticev1.ConfigSpec

	addressLister       latticelisters.AddressLister
	addressListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced
}

var (
	updateWaitBeforeFlushTimerSeconds = 5
)

// NewController returns a newly created DNS Controller.
func NewController(
	latticeID v1.LatticeID,
	namespacePrefix string,
	internalDNSDomain string,
	dnsmasqConfigPath string,
	dnsmasqHostsFilePath string,
	serviceMeshOptions *servicemesh.Options,
	latticeClient latticeclientset.Interface,
	kubeClient kubeclientset.Interface,
	configInformer latticeinformers.ConfigInformer,
	addressInformer latticeinformers.AddressInformer,
	serviceInformer latticeinformers.ServiceInformer,
) *Controller {

	c := &Controller{
		latticeID:       latticeID,
		namespacePrefix: namespacePrefix,

		internalDNSDomain: internalDNSDomain,

		dnsmasqConfigPath:    dnsmasqConfigPath,
		dnsmasqHostsFilePath: dnsmasqHostsFilePath,

		staticServiceMeshOptions: serviceMeshOptions,

		latticeClient: latticeClient,

		configSetChan: make(chan struct{}),
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubeClient.CoreV1().RESTClient()).Events("")})

	configInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// It's assumed there is always one and only one config object.
		AddFunc:    c.handleConfigAdd,
		UpdateFunc: c.handleConfigUpdate,
	})
	c.configLister = configInformer.Lister()
	c.configListerSynced = configInformer.Informer().HasSynced

	c.addressLister = addressInformer.Lister()
	c.addressListerSynced = addressInformer.Informer().HasSynced

	c.serviceLister = serviceInformer.Lister()
	c.serviceListerSynced = serviceInformer.Informer().HasSynced

	c.externalNameAddresses = set.NewSet()
	c.serviceAddresses = set.NewSet()

	return c
}

// Run triggers a goroutine to refresh the cache and rewrite the config at set intervals, until it receives from stopCh
func (c *Controller) Run(stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()

	glog.Infof("Running controller")
	defer glog.Infof("Shutting down controller")

	// wait for your secondary caches to fill before starting your work.
	if !cache.WaitForCacheSync(stopCh, c.addressListerSynced, c.serviceListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced. Waiting for config to be set")

	// wait for config to be set
	<-c.configSetChan

	go wait.Until(c.syncAddresses, time.Second*time.Duration(updateWaitBeforeFlushTimerSeconds), stopCh)

	// wait until we're told to stop
	<-stopCh
}

func (c *Controller) handleConfigAdd(obj interface{}) {
	config := obj.(*latticev1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = config.DeepCopy().Spec

	err := c.newServiceMesh()
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}

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

	err := c.newServiceMesh()
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}
}

func (c *Controller) newServiceMesh() error {
	options, err := servicemesh.OverlayConfigOptions(c.staticServiceMeshOptions, &c.config.ServiceMesh)
	if err != nil {
		return err
	}

	serviceMesh, err := servicemesh.NewServiceMesh(options)
	if err != nil {
		return err
	}

	c.serviceMesh = serviceMesh
	return nil
}

// syncAddresses runs at regular intervals and compares the endpoints the configuration was written with against
// the current list of endpoints. If there is any difference, the configuration is rewritten.
func (c *Controller) syncAddresses() {
	addresses, err := c.addressLister.List(labels.Everything())
	if err != nil {
		runtime.HandleError(err)
		return
	}

	externalNameAddressSet, serviceAddressSet, err := c.addressSets(addresses)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	updateNeeded := false
	if externalNameAddressSet.SymmetricDifference(c.externalNameAddresses).Cardinality() > 0 {
		updateNeeded = true
	}

	if serviceAddressSet.SymmetricDifference(c.serviceAddresses).Cardinality() > 0 {
		updateNeeded = true
	}

	if !updateNeeded {
		return
	}

	glog.V(5).Infof("addresses have changed, rewriting dnsmasq configuration...")

	err = c.rewriteDnsmasqConfig(addresses)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	c.externalNameAddresses = externalNameAddressSet
	c.serviceAddresses = serviceAddressSet
}

// addressSets returns two sets of endpoints - the first for endpoints specified by an external name,
// and the second for endpoints specified by an value.
func (c *Controller) addressSets(addresses []*latticev1.Address) (set.Set, set.Set, error) {
	externalNameAddresses := set.NewSet()
	serviceAddresses := set.NewSet()

	for _, address := range addresses {
		key := fmt.Sprintf("%v/%v", address.Namespace, address.Name)
		if address.Spec.ExternalName != nil {
			endpointKey := fmt.Sprintf("/%v", *address.Spec.ExternalName)
			externalNameAddresses.Add(key + endpointKey)
			continue
		}

		if address.Spec.Service != nil {
			ipKey := fmt.Sprintf("/%v", address.Spec.Service.String())
			serviceAddresses.Add(key + ipKey)
			continue
		}

		return nil, nil, fmt.Errorf("address %v had neither external name nor service", key)
	}

	return externalNameAddresses, serviceAddresses, nil
}

// rewriteDnsmasqConfig rewrites the config files for the dns server
func (c *Controller) rewriteDnsmasqConfig(addresses []*latticev1.Address) error {
	// This is an extra config file, so contains only the options which must be rewritten.
	// Condition on cname is that it exists in the specified name file, or references another cname.
	// Each cname entry of the form cname=ALIAS,...(addn aliases),TARGET
	// Full specification here: http://www.thekelleys.org.uk/dnsmasq/docs/dnsmasq-man.html
	dnsmasqConfigFileContents := ""
	hostConfigFileContents := ""

	// Hold a consistent view of config throughout the rewrite
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	for _, address := range addresses {
		systemID, err := kubeutil.SystemID(c.namespacePrefix, address.Namespace)
		if err != nil {
			return err
		}

		path, err := address.PathLabel()
		if err != nil {
			glog.Errorf("error getting path for %v: %v, ignoring the address", address.Description(c.namespacePrefix), err)
			continue
		}

		domain := kubeutil.FullyQualifiedInternalAddressSubdomain(path.ToDomain(), systemID, c.latticeID, c.internalDNSDomain)

		if address.Spec.Service != nil {
			service, err := c.service(address.Namespace, *address.Spec.Service)
			if err != nil {
				glog.Errorf("error getting service for service %v for %v: %v, ignoring the address", *address.Spec.Service, address.Description(c.namespacePrefix), err)
				continue
			}

			if service == nil {
				glog.Warningf("service %v for %v does not exist, ignoring the address", *address.Spec.Service, address.Description(c.namespacePrefix))
				continue
			}

			ip, err := c.serviceMesh.HasWorkloadIP(address)
			if err != nil {
				glog.Errorf("error getting service for value for %v (%v): %v, ignoring the address", address.Description(c.namespacePrefix), service.Description(c.namespacePrefix), err)
				continue
			} else if ip == "" {
				glog.V(4).Infof("Service %v does not have a WorkloadIP assigned yet, skipping...", path)
				continue
			}

			hostConfigFileContents += fmt.Sprintf("%v %v\n", ip, domain)
		}

		if address.Spec.ExternalName != nil {
			dnsmasqConfigFileContents += fmt.Sprintf("cname=%v,%v\n", domain, *address.Spec.ExternalName)
		}
	}

	err := ioutil.WriteFile(c.dnsmasqConfigPath, []byte(dnsmasqConfigFileContents), 0660)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.dnsmasqHostsFilePath, []byte(hostConfigFileContents), 0660)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) service(namespace string, path tree.Path) (*latticev1.Service, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServicePathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	services, err := c.serviceLister.Services(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	if len(services) > 1 {
		return nil, fmt.Errorf("found multiple services for path %v in namespace %v", path.String(), namespace)
	}

	return services[0], nil
}
