package main

import (
	"flag"
	"time"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local/dns/controller"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
)

var (
	kubeconfig        string
	hostsFilePath     string
	dnsmasqConfigPath string
	latticeID         string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&latticeID, "lattice-id", "", "ID of the lattice")
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

	versionedLatticeClient := latticeclientset.NewForConfigOrDie(config)
	latticeInformers := latticeinformers.NewSharedInformerFactory(versionedLatticeClient, time.Duration(12*time.Hour))

	glog.V(1).Info("Starting dns controller")

	go controller.NewController(
		dnsmasqConfigPath,
		hostsFilePath,
		v1.LatticeID(latticeID),
		versionedLatticeClient,
		clientset.NewForConfigOrDie(config),
		latticeInformers.Lattice().V1().Endpoints(),
	).Run(stop)

	glog.V(1).Info("Starting informer factory")
	latticeInformers.Start(stop)

	select {}
}
