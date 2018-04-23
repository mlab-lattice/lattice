package main

import (
	"flag"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local/dns/controller"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
)

var (
	kubeconfig        string
	hostsFilePath     string
	dnsmasqConfigPath string
	latticeID         string
	namespacePrefix   string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&latticeID, "lattice-id", "", "ID of the lattice")
	flag.StringVar(&namespacePrefix, "namespace-prefix", "", "namespace prefix for the lattice")
	flag.StringVar(&dnsmasqConfigPath, "dnsmasq-config-path", local.DnsmasqConfigFile, "path to the additional dnsmasq configuration file")
	flag.StringVar(&hostsFilePath, "hosts-file-path", local.DNSHostsFile, "path to the additional dnsmasq hosts")
	flag.Parse()
}

func main() {
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

	stop := make(chan struct{})

	latticeClient := latticeclientset.NewForConfigOrDie(config)
	latticeInformers := latticeinformers.NewSharedInformerFactory(latticeClient, time.Duration(12*time.Hour))

	glog.V(1).Info("Starting dns controller")

	go controller.NewController(
		namespacePrefix,
		v1.LatticeID(latticeID),
		dnsmasqConfigPath,
		hostsFilePath,
		latticeClient,
		kubeclientset.NewForConfigOrDie(config),
		latticeInformers.Lattice().V1().Configs(),
		latticeInformers.Lattice().V1().Addresses(),
		latticeInformers.Lattice().V1().Services(),
	).Run(stop)

	glog.V(1).Info("Starting informer factory")
	latticeInformers.Start(stop)

	select {}
}
