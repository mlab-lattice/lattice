package local_dns

import (
	"time"

	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	informers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//This should watch for changes in containers (?), and rewrite /etc/resolv.conf (or specified file) according to
// the new envoy target ip:port, using the envoy mapping.
// kube ip got from the config file. envoy ip from the mapping

//Create AddressInformer in informers directory to replace SystemInformer

//TODO :: Naming unsure stil, impl local-dns/backend
type KubernetesLocalDNSBackend struct {
	AddressWatcherInformer informers.SystemInformer
}

//Meaning of kubeconfig here
func NewKubernetesLocalDNSBackend(kubeconfig string) (*KubernetesLocalDNSBackend, error) {

	var config *rest.Config
	var err error

	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		panic(err)
	}

	lcb := controller.LatticeClientBuilder{
		Kubeconfig: config,
	}

	versionedLatticeClient := lcb.ClientOrDie("shared-kubeinformers")
	latticeInformers := latticeinformers.NewSharedInformerFactory(versionedLatticeClient, time.Duration(12*time.Hour))

	//Do I need to add in cmd/lcm/app do i need to add a new folder / file for the local case and add a new DNS controller

	//Create controllers

		//
		// ctx.LatticeInformerFactory.Lattice().V1().ServiceBuilds(),
		// ctx.LatticeInformerFactory.Lattice().V1().ComponentBuilds(),

		// new controller created with these params. controller then adds event handler to each thing
		// Get informer,Informer().AddEventHandler()

	//Still any use of the Ready / Services func

	return KubernetesLocalDNSBackend{
		AddressWatcherInformer:latticeInformers,
	}

	return nil, nil
}

func (kldb *KubernetesLocalDNSBackend) Ready() bool {
	return true
}

func (kldb *KubernetesLocalDNSBackend) Services() bool {
	return true
}