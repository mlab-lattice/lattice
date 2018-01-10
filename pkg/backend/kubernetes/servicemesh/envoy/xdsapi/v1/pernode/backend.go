package pernode

import (
	"time"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh/envoy"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	xdsapi "github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesPerNodeBackend struct {
	serviceMesh envoy.ServiceMesh

	kubeEndpointLister       corelisters.EndpointsLister
	kubeEndpointListerSynced cache.InformerSynced

	serviceLister       latticelisters.ServiceLister
	serviceListerSynced cache.InformerSynced
}

func NewKubernetesPerNodeBackend(kubeconfig, namespace string) (*KubernetesPerNodeBackend, error) {
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
	//kubeInformers := kubeinformers.NewSharedInformerFactory(kubeClient, time.Duration(12*time.Hour))
	kubeInformers := kubeinformers.NewFilteredSharedInformerFactory(kubeClient, time.Duration(12*time.Hour), namespace, nil)

	latticeClient, err := latticeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	//latticeInformers := latticeinformers.NewSharedInformerFactory(latticeClient, time.Duration(12*time.Hour))
	latticeInformers := latticeinformers.NewFilteredSharedInformerFactory(latticeClient, time.Duration(12*time.Hour), namespace, nil)

	// FIXME: should we add a stopCh?
	go kubeInformers.Start(nil)
	go latticeInformers.Start(nil)

	kubeEndpointInformer := kubeInformers.Core().V1().Endpoints()
	serviceInformer := latticeInformers.Lattice().V1().Services()

	envoyConfig := &crv1.ConfigServiceMesh{
		Envoy: &crv1.ConfigEnvoy{},
	}
	serviceMesh, err := servicemesh.NewServiceMesh(envoyConfig)
	if err != nil {
		return nil, err
	}

	b := &KubernetesPerNodeBackend{
		serviceMesh:              serviceMesh.(envoy.ServiceMesh),
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
		kubeServiceName := kubernetes.GetKubeServiceNameForService(service.Name)

		endpoint, err := b.kubeEndpointLister.Endpoints(service.Namespace).Get(kubeServiceName)
		if err != nil {
			return nil, err
		}

		egressPort, err := b.serviceMesh.EgressPort(service)
		if err != nil {
			return nil, err
		}

		bsvc := &xdsapi.Service{
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
					bsvc.IPAddresses = append(bsvc.IPAddresses, address.IP)
				}
			}
		}

		for component, ports := range service.Spec.Ports {
			bc := xdsapi.Component{
				Ports: map[int32]int32{},
			}

			for _, port := range ports {
				envoyPort, err := b.serviceMesh.ServicePort(service, port.Port)
				if err != nil {
					return nil, err
				}

				bc.Ports[port.Port] = envoyPort
			}

			bsvc.Components[component] = bc
		}

		result[service.Spec.Path] = bsvc
	}

	return result, nil
}
