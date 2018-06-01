package pernode

import (
	"encoding/json"
	"fmt"
	// "reflect"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
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

	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
)

// TODO: add events

type KubernetesPerNodeBackend struct {
	serviceMesh *envoy.DefaultEnvoyServiceMesh

	kubeEndpointLister       corelisters.EndpointsLister
	kubeEndpointListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	xdsCache envoycache.SnapshotCache

	stopCh <-chan struct{}
}

// I0531 16:50:06.323342       1 server.go:16] Per-node GRPC server starting on port 8080
// I0531 16:50:06.323671       1 backend.go:111] Starting per-node backend...
// I0531 16:50:06.323687       1 backend.go:112] Waiting for caches to sync
// I0531 16:50:06.323989       1 reflector.go:202] Starting reflector *v1.Endpoints (12h0m0s) from external/io_k8s_client_go/tools/cache/reflector.go:99
// I0531 16:50:06.324019       1 reflector.go:240] Listing and watching *v1.Endpoints from external/io_k8s_client_go/tools/cache/reflector.go:99
// I0531 16:50:06.324605       1 reflector.go:202] Starting reflector *v1.Service (12h0m0s) from external/io_k8s_client_go/tools/cache/reflector.go:99
// I0531 16:50:06.324624       1 reflector.go:240] Listing and watching *v1.Service from external/io_k8s_client_go/tools/cache/reflector.go:99
// I0531 16:50:06.423922       1 shared_informer.go:123] caches populated
// I0531 16:50:06.423998       1 backend.go:118] Starting workers
// I0531 16:50:06.424010       1 backend.go:123] Per-node backend started
// I0531 16:58:07.345046       1 reflector.go:428] external/io_k8s_client_go/tools/cache/reflector.go:99: Watch close - *v1.Endpoints total 480 items received
// I0531 16:59:48.350041       1 reflector.go:428] external/io_k8s_client_go/tools/cache/reflector.go:99: Watch close - *v1.Service total 0 items received
// I0531 17:00:57.290362       1 backend.go:266] OnStreamOpen called: 1,
// I0531 17:00:57.290448       1 backend.go:262] OnStreamRequest called: 1, node:<build_version:"0bcdb5d7611a79fd22f823fd707a8b6f7b5f756e/1.7.0-dev/Clean/RELEASE" > type_url:"type.googleapis.com/envoy.api.v2.Cluster"
// I0531 17:00:57.290571       1 backend.go:252] open watch 1 for type.googleapis.com/envoy.api.v2.Cluster[] from nodeID "", version ""
// I0531 17:00:58.589383       1 backend.go:266] OnStreamOpen called: 2,
// I0531 17:00:58.591280       1 backend.go:262] OnStreamRequest called: 2, node:<build_version:"0bcdb5d7611a79fd22f823fd707a8b6f7b5f756e/1.7.0-dev/Clean/RELEASE" > type_url:"type.googleapis.com/envoy.api.v2.Cluster"
// I0531 17:00:58.591376       1 backend.go:252] open watch 2 for type.googleapis.com/envoy.api.v2.Cluster[] from nodeID "", version ""

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
	b.xdsCache = envoycache.NewSnapshotCache(true, b, b)

	kubeEndpointInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			task, err := b.enqueueCacheUpdateTask(xdsapi.KubeEntityType, obj)
			if err != nil {
				glog.Error(err)
				runtime.HandleError(err)
			} else {
				glog.Infof("Got Kube \"Add\" event: %s", task)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			_old := old.(metav1.Object)
			_cur := cur.(metav1.Object)
			glog.Infof("old version: %s, new version: %s", _old.GetResourceVersion(), _cur.GetResourceVersion())
			// if !reflect.DeepEqual(old, cur) {
			if _old.GetResourceVersion() != _cur.GetResourceVersion() {
				task, err := b.enqueueCacheUpdateTask(xdsapi.KubeEntityType, cur)
				if err != nil {
					glog.Error(err)
					runtime.HandleError(err)
				} else {
					glog.Infof("Got Kube \"Update\" event: %s", task)
				}
			} else {
				glog.Info("Skipping Kube \"Update\" event: old and current objects are equal")
			}
		},
		DeleteFunc: func(obj interface{}) {
			task, err := b.enqueueCacheUpdateTask(xdsapi.KubeEntityType, obj)
			if err != nil {
				glog.Error(err)
				runtime.HandleError(err)
			} else {
				glog.Infof("Got Kube \"Delete\" event: %s", task)
			}
		},
	}, time.Duration(1*time.Minute))
	serviceInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			task, err := b.enqueueCacheUpdateTask(xdsapi.LatticeEntityType, obj)
			if err != nil {
				glog.Error(err)
				runtime.HandleError(err)
			} else {
				glog.Infof("Got Lattice \"Add\" event: %s", task)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			_old := old.(metav1.Object)
			_cur := cur.(metav1.Object)
			glog.Infof("old version: %s, new version: %s", _old.GetResourceVersion(), _cur.GetResourceVersion())
			// if !reflect.DeepEqual(old, cur) {
			if _old.GetResourceVersion() != _cur.GetResourceVersion() {
				task, err := b.enqueueCacheUpdateTask(xdsapi.LatticeEntityType, cur)
				if err != nil {
					glog.Error(err)
					runtime.HandleError(err)
				} else {
					glog.Infof("Got Lattice \"Update\" event: %s", task)
				}
			} else {
				glog.Infof("Skipping Lattice \"Update\" event: old and current objects are equal")
			}
		},
		DeleteFunc: func(obj interface{}) {
			task, err := b.enqueueCacheUpdateTask(xdsapi.LatticeEntityType, obj)
			if err != nil {
				glog.Error(err)
				runtime.HandleError(err)
			} else {
				glog.Infof("Got Lattice \"Delete\" event: %s", task)
			}
		},
	}, time.Duration(1*time.Minute))

	return b, nil
}

// getters

func (b *KubernetesPerNodeBackend) XDSCache() envoycache.Cache {
	return b.xdsCache
}

// methods

func (b *KubernetesPerNodeBackend) enqueueCacheUpdateTask(_type xdsapi.EntityType, obj interface{}) (string, error) {
	var err error
	var name string
	var task []byte

	name, err = cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return "", err
	}
	task, err = json.Marshal(xdsapi.CacheUpdateTask{
		Name: name,
		Type: _type,
	})
	if err != nil {
		return "", err
	}

	taskKey := string(task[:])
	b.queue.Add(taskKey)
	return taskKey, nil
}

func (b *KubernetesPerNodeBackend) Ready() bool {
	return cache.WaitForCacheSync(nil, b.kubeEndpointListerSynced, b.serviceListerSynced)
}

func (b *KubernetesPerNodeBackend) Run(threadiness int) error {
	defer runtime.HandleCrash()
	defer b.queue.ShutDown()

	glog.Info("Starting per-node backend...")
	glog.Info("Waiting for caches to sync")

	if err := b.Ready(); !err {
		return fmt.Errorf("Failed to sync caches")
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
				return fmt.Errorf("Per-node backend worker got: %#v", obj)
			}

			if err := b.syncXDSCache(key); err != nil {
				return fmt.Errorf("Per-node backend got error syncing XDS cache for '%s': %s", key, err.Error())
			}

			b.queue.Forget(obj)
			return nil
		}(obj)

		if err != nil {
			runtime.HandleError(err)
		}
	}
}

func (b *KubernetesPerNodeBackend) syncXDSCache(key string) error {
	glog.Infof("Per-node backend syncing '%s'", key)
	glog.Infof("Per-node backend synced '%s'", key)
	return nil
}

func (b *KubernetesPerNodeBackend) Services(serviceCluster string) (map[tree.NodePath]*xdsapi.Service, error) {
	// TODO: probably want to have Services return a cached snapshot of the service state so we don't have to recompute this every time
	// 	     For example, could add hooks to the informers which creates a new Services map and changes the pointer to point to the new one
	//       so future Services() calls will return the new map.
	// 		 Could also have the backend have a channel passed into it and it could notify the API when an update has occurred.
	//       This could be useful for the GRPC streaming version of the API.
	// N.B.: keep an eye on https://github.com/envoyproxy/go-control-plane
	namespace := serviceCluster
	result := map[tree.NodePath]*xdsapi.Service{}

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

		egressPort, err := b.serviceMesh.EgressPort(service)
		if err != nil {
			return nil, err
		}

		xdsService := &xdsapi.Service{
			EgressPort:  egressPort,
			Components:  map[string]xdsapi.Component{},
			IPAddresses: []string{},
		}

		addressSet := map[string]bool{}
		for _, subset := range endpoint.Subsets {
			for _, address := range subset.Addresses {
				// FIXME: check if this is necessary (i.e. does Endpoint ever repeat IPAddresses)
				if _, ok := addressSet[address.IP]; !ok {
					addressSet[address.IP] = true
					xdsService.IPAddresses = append(xdsService.IPAddresses, address.IP)
				}
			}
		}

		for component, ports := range service.Spec.Ports {
			bc := xdsapi.Component{
				Ports: map[int32]int32{},
			}

			for _, port := range ports {
				envoyPort, err := b.serviceMesh.ServiceMeshPort(service, port.Port)
				if err != nil {
					return nil, err
				}

				bc.Ports[port.Port] = envoyPort
			}

			xdsService.Components[component] = bc
		}

		result[path] = xdsService
	}

	return result, nil
}

// interface implementations

// github.com/envoyproxy/go-control-plane/pkg/cache#NodeHash{} -- for b.xdsCache

func (b *KubernetesPerNodeBackend) ID(node *envoycore.Node) string {
	return node.GetId()
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
	glog.Infof("OnStreamRequest called: %d, %v", id, req)
}

func (b *KubernetesPerNodeBackend) OnStreamOpen(id int64, urlType string) {
	glog.Infof("OnStreamOpen called: %d, %v", id, urlType)
}

func (b *KubernetesPerNodeBackend) OnStreamClosed(id int64) {
	glog.Infof("OnStreamClosed called: %d", id)
}

func (b *KubernetesPerNodeBackend) OnStreamResponse(id int64, req *envoyv2.DiscoveryRequest, res *envoyv2.DiscoveryResponse) {
	glog.Infof("OnStreamResponse called: %d, %v, %v", id, req, res)
}

func (b *KubernetesPerNodeBackend) OnFetchRequest(req *envoyv2.DiscoveryRequest) {
	glog.Infof("OnFetchRequest called: %v", req)
}
func (b *KubernetesPerNodeBackend) OnFetchResponse(req *envoyv2.DiscoveryRequest, res *envoyv2.DiscoveryResponse) {
	glog.Infof("OnFetchRequest called: %v", req, res)
}
