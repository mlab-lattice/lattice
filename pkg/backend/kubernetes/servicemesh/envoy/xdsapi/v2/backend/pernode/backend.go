package pernode

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
	xdsservicenode "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2/service_node"
)

// TODO: add events

type KubernetesPerNodeBackend struct {
	serviceMesh *envoy.DefaultEnvoyServiceMesh

	kubeEndpointLister       corelisters.EndpointsLister
	kubeEndpointListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	count     int
	countLock sync.Mutex

	lock sync.Mutex

	serviceNodes map[string]*xdsservicenode.ServiceNode
	xdsCache     envoycache.SnapshotCache

	stopCh <-chan struct{}
}

func NewKubernetesPerNodeBackend(kubeconfig string, stopCh <-chan struct{}) (*KubernetesPerNodeBackend, error) {
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}

	rest.AddUserAgent(config, "envoy-api-backend")
	kubeClient, err := kubeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	kubeInformers := kubeinformers.NewSharedInformerFactory(kubeClient, time.Duration(12*time.Hour))

	latticeClient, err := latticeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	latticeInformers := latticeinformers.NewSharedInformerFactory(latticeClient, time.Duration(12*time.Hour))

	go kubeInformers.Start(stopCh)
	go latticeInformers.Start(stopCh)

	kubeEndpointInformer := kubeInformers.Core().V1().Endpoints()
	serviceInformer := latticeInformers.Lattice().V1().Services()

	b := &KubernetesPerNodeBackend{
		serviceMesh:              envoy.NewEnvoyServiceMesh(&envoy.Options{}),
		kubeEndpointLister:       kubeEndpointInformer.Lister(),
		kubeEndpointListerSynced: kubeEndpointInformer.Informer().HasSynced,
		serviceLister:            serviceInformer.Lister(),
		serviceListerSynced:      serviceInformer.Informer().HasSynced,
		queue:                    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "envoy-api-backend"),
		stopCh:                   stopCh,
	}
	b.serviceNodes = make(map[string]*xdsservicenode.ServiceNode)
	b.xdsCache = envoycache.NewSnapshotCache(true, b, b)

	serviceInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			task, err := b.enqueueCacheUpdateTask(xdsapi.LatticeEntityType, xdsapi.InformerAddEvent, obj)
			if err != nil {
				runtime.HandleError(err)
			} else {
				glog.V(4).Infof("Got Lattice \"Add\" event: %s", task)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			_old := old.(metav1.Object)
			_cur := cur.(metav1.Object)
			glog.V(4).Infof("old version: %s, new version: %s", _old.GetResourceVersion(), _cur.GetResourceVersion())
			if _old.GetResourceVersion() != _cur.GetResourceVersion() {
				task, err := b.enqueueCacheUpdateTask(xdsapi.LatticeEntityType, xdsapi.InformerUpdateEvent, cur)
				if err != nil {
					runtime.HandleError(err)
				} else {
					glog.V(4).Infof("Got Lattice \"Update\" event: %s", task)
				}
			} else {
				glog.V(4).Infof("Skipping Lattice \"Update\" event: old and current objects are equal")
			}
		},
		DeleteFunc: func(obj interface{}) {
			task, err := b.enqueueCacheUpdateTask(xdsapi.LatticeEntityType, xdsapi.InformerDeleteEvent, obj)
			if err != nil {
				runtime.HandleError(err)
			} else {
				glog.V(4).Infof("Got Lattice \"Delete\" event: %s", task)
			}
		},
	}, time.Duration(12*time.Hour))

	return b, nil
}

// getters

func (b *KubernetesPerNodeBackend) XDSCache() envoycache.Cache {
	return b.xdsCache
}

// methods

func (b *KubernetesPerNodeBackend) getNextVersion() string {
	b.countLock.Lock()
	defer b.countLock.Unlock()
	b.count++
	return fmt.Sprintf("%d", b.count)
}

func (b *KubernetesPerNodeBackend) SetXDSCacheSnapshot(id string, endpoints, clusters, routes, listeners []envoycache.Resource) error {
	// NOTE: do not call b.lock.Lock, xdsCache guarded by internal lock
	b.xdsCache.SetSnapshot(id, envoycache.NewSnapshot(b.getNextVersion(), endpoints, clusters, routes, listeners))

	return nil
}

func (b *KubernetesPerNodeBackend) ClearXDSCacheSnapshot(id string) error {
	// NOTE: do not call b.lock.Lock, xdsCache guarded by internal lock
	b.xdsCache.ClearSnapshot(id)

	return nil
}

func (b *KubernetesPerNodeBackend) enqueueCacheUpdateTask(_type xdsapi.EntityType, event xdsapi.Event, obj interface{}) (string, error) {
	var err error
	var ok bool
	var name string
	var task []byte

	switch _type {
	case xdsapi.EnvoyEntityType:
		name, ok = obj.(string)
		if !ok {
			return "", err
		}
	case xdsapi.LatticeEntityType:
		if event == xdsapi.InformerDeleteEvent {
			// generates name in the format "<namespace>/<name>"
			name, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if _obj, ok := obj.(cache.DeletedFinalStateUnknown); ok {
				// XXX: does this require any other special handling when it comes to listing?
				// XXX: handle "tombstones"? https://github.com/kubernetes/sample-controller/blob/master/controller.go#L351
				_objOut, _ := json.MarshalIndent(_obj, "", "  ")
				glog.Warningf("Got DeletedFinalStateUnknown obj:\n%s", string(_objOut))
			}
		} else {
			// generates name in the format "<namespace>/<name>"
			name, err = cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return "", err
			}
		}
	default:
		return "", fmt.Errorf("got unkown entity type <%d>", _type)
	}

	task, err = json.Marshal(xdsapi.CacheUpdateTask{
		Name:  name,
		Type:  _type,
		Event: event,
	})
	if err != nil {
		return "", err
	}

	taskKey := string(task[:])
	b.queue.Add(taskKey)
	return taskKey, nil
}

func (b *KubernetesPerNodeBackend) Ready() bool {
	return cache.WaitForCacheSync(b.stopCh, b.kubeEndpointListerSynced, b.serviceListerSynced)
}

func (b *KubernetesPerNodeBackend) Run(threadiness int) error {
	defer runtime.HandleCrash()
	defer b.queue.ShutDown()

	glog.Info("Starting per-node backend...")
	glog.Info("Waiting for caches to sync")

	if ok := b.Ready(); !ok {
		return fmt.Errorf("failed to sync caches")
	}

	glog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(b.Worker, time.Second, b.stopCh)
	}

	glog.Info("Per-node backend started")
	<-b.stopCh
	glog.Info("Per-node backend stopped")
	return nil
}

func (b *KubernetesPerNodeBackend) Worker() {
	glog.Info("Per-node backend worker starting...")
	for {
		obj, shutdown := b.queue.Get()
		if shutdown {
			glog.Info("Per-node backend worker shutting down")
			return
		}
		err := func(obj interface{}) error {
			defer b.queue.Done(obj)

			var key string
			var ok bool

			if key, ok = obj.(string); !ok {
				b.queue.Forget(obj)
				return fmt.Errorf("per-node backend worker got: %#v", obj)
			}

			if err := b.syncXDSCache(key); err != nil {
				return fmt.Errorf("per-node backend got error syncing XDS cache for '%s': %s", key, err.Error())
			}

			b.queue.Forget(obj)
			return nil
		}(obj)

		if err != nil {
			runtime.HandleError(err)
		}
	}
}

func (b *KubernetesPerNodeBackend) getServiceNode(id string) (*xdsservicenode.ServiceNode, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	service, ok := b.serviceNodes[id]
	if !ok {
		return nil, fmt.Errorf("couldn't find Envoy service with ID <%s>", id)
	}
	return service, nil
}

func (b *KubernetesPerNodeBackend) handleEnvoySyncXDSCache(entityName string) error {
	glog.Infof("Per-node backend handling envoy sync task")
	service, err := b.getServiceNode(entityName)
	if err != nil {
		return err
	}
	// find and set the lattice service name that corresponds to this envoy service node to aid in cleanup
	if name := service.GetLatticeServiceName(); name == "" {
		selector := labels.NewSelector()
		requirement, err := labels.NewRequirement(
			latticev1.ServicePathLabelKey, selection.Equals, []string{service.Domain()})
		if err != nil {
			return err
		}
		selector = selector.Add(*requirement)
		services, err := b.serviceLister.Services(service.ServiceCluster()).List(selector)
		if err != nil {
			return err
		}
		if len(services) != 1 {
			return fmt.Errorf("found %d services matching %v, expected 1", len(services), selector)
		}
		service.SetLatticeServiceName(services[0].Name)
	}
	return service.Update(b)
}

func (b *KubernetesPerNodeBackend) getServicesForServiceCluster(serviceCluster string) []*xdsservicenode.ServiceNode {
	var services []*xdsservicenode.ServiceNode

	b.lock.Lock()
	defer b.lock.Unlock()

	for serviceID, service := range b.serviceNodes {
		if _serviceCluster, _, err := cache.SplitMetaNamespaceKey(serviceID); err == nil && _serviceCluster == serviceCluster {
			services = append(services, service)
		}
	}

	return services
}

func (b *KubernetesPerNodeBackend) handleLatticeSyncXDSCache(entityName string, event xdsapi.Event) error {
	glog.Infof("Per-node backend handling lattice sync task")
	serviceCluster, serviceName, err := cache.SplitMetaNamespaceKey(entityName)
	if err != nil {
		return err
	}
	if event == xdsapi.InformerDeleteEvent {
		err = func() error {
			b.lock.Lock()
			defer b.lock.Unlock()
			var serviceNodeID string
			for _serviceNodeID, serviceNode := range b.serviceNodes {
				// use the lattice service name we set earlier to identify the envoy service node
				// to clean up
				if serviceNode.GetLatticeServiceName() == serviceName {
					serviceNodeID = _serviceNodeID
					break
				}
			}
			if serviceNodeID == "" {
				return fmt.Errorf("couldn't find <%v> on delete", serviceName)
			}
			glog.Infof("Deleting node <%v> and clearing its cache", serviceNodeID)
			b.serviceNodes[serviceNodeID].Cleanup(b)
			delete(b.serviceNodes, serviceNodeID)
			return nil
		}()
		if err != nil {
			return err
		}
	}
	var serviceNodeKeys []string
	b.lock.Lock()
	for k := range b.serviceNodes {
		serviceNodeKeys = append(serviceNodeKeys, k)
	}
	b.lock.Unlock()
	glog.V(4).Infof("Remaining service nodes: %v", serviceNodeKeys)
	glog.V(4).Infof("Remaining cache nodes: %v", b.xdsCache.GetStatusKeys())
	for _, service := range b.getServicesForServiceCluster(serviceCluster) {
		err = service.Update(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *KubernetesPerNodeBackend) syncXDSCache(key string) error {
	glog.V(4).Infof("Per-node backend syncing '%s'", key)
	var err error

	task := xdsapi.CacheUpdateTask{}

	err = json.Unmarshal([]byte(key), &task)
	if err != nil {
		return err
	}

	switch task.Type {
	case xdsapi.EnvoyEntityType:
		err = b.handleEnvoySyncXDSCache(task.Name)
	case xdsapi.LatticeEntityType:
		err = b.handleLatticeSyncXDSCache(task.Name, task.Event)
	default:
		return fmt.Errorf("got unkown entity type <%d>", task.Type)
	}

	if err == nil {
		glog.V(4).Infof("Per-node backend synced '%s'", key)
	}
	return err
}

func (b *KubernetesPerNodeBackend) SystemServices(serviceCluster string) (map[tree.NodePath]*xdsapi.Service, error) {
	namespace := serviceCluster
	result := make(map[tree.NodePath]*xdsapi.Service)

	services, err := b.serviceLister.Services(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		path, err := service.PathLabel()
		if err != nil {
			// FIXME: this shouldn't happen. send an error somewhere?
			continue
		}

		kubeServiceName := kubernetes.GetKubeServiceNameForService(service.Name)
		endpoint, err := b.kubeEndpointLister.Endpoints(service.Namespace).Get(kubeServiceName)
		if err != nil {
			return nil, err
		}

		egressPorts, err := b.serviceMesh.EgressPorts(service)
		if err != nil {
			return nil, err
		}

		xdsService := &xdsapi.Service{
			EgressPorts: *egressPorts,
			Components:  map[string]xdsapi.Component{},
			IPAddresses: []string{},
		}

		addressSet := make(map[string]bool)
		for _, subset := range endpoint.Subsets {
			for _, address := range subset.Addresses {
				// FIXME: check if this is necessary (i.e. does Endpoint ever repeat IPAddresses)
				if _, ok := addressSet[address.IP]; !ok {
					addressSet[address.IP] = true
					xdsService.IPAddresses = append(xdsService.IPAddresses, address.IP)
				}
			}
		}

		// FIXME: we should reevaluate these structures. Component isn't a thing anymore.
		mainContainer := xdsapi.Component{
			Ports: make(map[int32]int32),
		}
		for port := range service.Spec.Definition.Ports {
			envoyPort, err := b.serviceMesh.ServiceMeshPort(service, port)
			if err != nil {
				return nil, err
			}

			mainContainer.Ports[port] = envoyPort
		}
		xdsService.Components[kubernetes.UserMainContainerName] = mainContainer

		for name, sidecar := range service.Spec.Definition.Sidecars {
			c := xdsapi.Component{
				Ports: make(map[int32]int32),
			}
			for port := range sidecar.Ports {
				envoyPort, err := b.serviceMesh.ServiceMeshPort(service, port)
				if err != nil {
					return nil, err
				}

				c.Ports[port] = envoyPort
			}
			xdsService.Components[kubernetes.UserSidecarContainerName(name)] = c
		}

		result[path] = xdsService
	}

	return result, nil
}

// interface implementations

// github.com/envoyproxy/go-control-plane/pkg/cache#NodeHash{} -- for b.xdsCache

func (b *KubernetesPerNodeBackend) ID(node *envoycore.Node) string {
	return node.GetCluster() + "/" + node.GetId()
}

// github.com/envoyproxy/go-control-plane/pkg/log#Logger{} -- for b.xdsCache

func (b *KubernetesPerNodeBackend) Infof(format string, args ...interface{}) {
	glog.Infof(format, args...)
}

func (b *KubernetesPerNodeBackend) Errorf(format string, args ...interface{}) {
	glog.Errorf(format, args...)
}

// github.com/envoyproxy/go-control-plane/pkg/server#Callbacks{} -- for github.com/envoyproxy/go-control-plane/pkg/server#Server

func (b *KubernetesPerNodeBackend) OnStreamRequest(id int64, req *envoyv2.DiscoveryRequest) {
	reqStr, _ := json.MarshalIndent(req, "", "  ")
	glog.V(4).Infof("OnStreamRequest called: %d\n%v", id, string(reqStr[:]))
	node := req.GetNode()
	serviceID := b.ID(node)

	b.lock.Lock()
	if _, ok := b.serviceNodes[serviceID]; !ok {
		b.serviceNodes[serviceID] = xdsservicenode.NewServiceNode(serviceID, node)
	}
	b.lock.Unlock()

	glog.V(4).Infof("Got node <%s>: %v", serviceID, node)

	task, err := b.enqueueCacheUpdateTask(xdsapi.EnvoyEntityType, xdsapi.EnvoyStreamRequestEvent, serviceID)
	if err != nil {
		runtime.HandleError(err)
	} else {
		glog.V(4).Infof("Got new Envoy connection task: %s", task)
	}
}

func (b *KubernetesPerNodeBackend) OnStreamOpen(id int64, urlType string) {
	glog.V(4).Infof("OnStreamOpen called: %d, %v", id, urlType)
}

func (b *KubernetesPerNodeBackend) OnStreamClosed(id int64) {
	glog.V(4).Infof("OnStreamClosed called: %d", id)
}

func (b *KubernetesPerNodeBackend) OnStreamResponse(id int64, req *envoyv2.DiscoveryRequest, res *envoyv2.DiscoveryResponse) {
	glog.V(4).Infof("OnStreamResponse called: %d, %v, %v", id, req, res)
}

func (b *KubernetesPerNodeBackend) OnFetchRequest(req *envoyv2.DiscoveryRequest) {
	glog.V(4).Infof("OnFetchRequest called: %v", req)
}
func (b *KubernetesPerNodeBackend) OnFetchResponse(req *envoyv2.DiscoveryRequest, res *envoyv2.DiscoveryResponse) {
	glog.V(4).Infof("OnFetchRequest called: %v", req, res)
}
