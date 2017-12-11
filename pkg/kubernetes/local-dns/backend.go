package local_dns

import (
	latticeresource "github.com/mlab-lattice/system/pkg/kubernetes/customresource"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"

	corev1informers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

//This should watch for changes in containers (?), and rewrite /etc/resolv.conf (or specified file) according to
// the new envoy target ip:port, using the envoy mapping.
// kube ip got from the config file. envoy ip from the mapping

//TODO :: Naming unsure stil, impl local-dns/backend
type KubernetesLocalDNSBackend struct {

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
		return nil, err
	}

	latticeResourceClient, _, err := latticeresource.NewClient(config) // I.e all of our lattice resources
	if err != nil {
		return nil, err
	}

	listerWatcher := cache.NewListWatchFromClient(
		latticeResourceClient,
		crv1.ServiceResourcePlural, //I.e. watches for lattice services. We want normal services?
		string(coreconstants.UserSystemNamespace),
		fields.Everything(),
	)

	//What are we watching for here? Just services?
	lSvcInformer := cache.NewSharedInformer(
		listerWatcher,
		%crv1.Service{},
		time.Duration(12*time.Hour), // May need to be more frequent?
	)

	return nil, nil
}

func (kldb *KubernetesLocalDNSBackend) Ready() bool {
	return true
}

func (kldb *KubernetesLocalDNSBackend) Services() bool {
	return true
}