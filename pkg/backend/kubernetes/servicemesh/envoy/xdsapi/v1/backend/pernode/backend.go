package pernode

import (
	"time"

	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy"
	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/labels"

	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesPerNodeBackend struct {
	serviceMesh *envoy.DefaultEnvoyServiceMesh

	kubeEndpointLister       corelisters.EndpointsLister
	kubeEndpointListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced
}

func NewKubernetesPerNodeBackend(kubeconfig string) (*KubernetesPerNodeBackend, error) {
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

	// FIXME: should we add a stopCh?
	go kubeInformers.Start(nil)
	go latticeInformers.Start(nil)

	kubeEndpointInformer := kubeInformers.Core().V1().Endpoints()
	serviceInformer := latticeInformers.Lattice().V1().Services()

	b := &KubernetesPerNodeBackend{
		serviceMesh:              envoy.NewEnvoyServiceMesh(&envoy.Options{}),
		kubeEndpointLister:       kubeEndpointInformer.Lister(),
		kubeEndpointListerSynced: kubeEndpointInformer.Informer().HasSynced,
		serviceLister:            serviceInformer.Lister(),
		serviceListerSynced:      serviceInformer.Informer().HasSynced,
	}
	return b, nil
}

func (b *KubernetesPerNodeBackend) Ready() bool {
	return cache.WaitForCacheSync(nil, b.kubeEndpointListerSynced, b.serviceListerSynced)
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

		kubeServiceName := kubeutil.GetKubeServiceNameForService(service.Name)
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
			Containers:  map[string]xdsapi.Container{},
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

		mainContainer := xdsapi.Container{
			Ports: make(map[int32]int32),
		}

		for portNum := range service.Spec.Definition.Ports {
			envoyPort, err := b.serviceMesh.ServiceMeshPort(service, portNum)
			if err != nil {
				return nil, err
			}

			mainContainer.Ports[portNum] = envoyPort
		}
		xdsService.Containers[kubeutil.UserMainContainerName] = mainContainer

		for sidecar, sidecarContainer := range service.Spec.Definition.Sidecars {
			sidecarXDSContainer := xdsapi.Container{
				Ports: make(map[int32]int32),
			}

			for portNum := range sidecarContainer.Ports {
				envoyPort, err := b.serviceMesh.ServiceMeshPort(service, portNum)
				if err != nil {
					return nil, err
				}

				sidecarXDSContainer.Ports[portNum] = envoyPort
			}
			xdsService.Containers[kubeutil.UserSidecarContainerName(sidecar)] = sidecarXDSContainer
		}

		result[path] = xdsService
	}

	return result, nil
}
