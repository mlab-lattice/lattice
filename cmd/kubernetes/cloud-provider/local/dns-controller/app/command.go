package app

import (
	"time"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local/dns/controller"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

func Command() *cli.Command {
	var kubeconfig string
	var namespacePrefix string
	var latticeID string
	var internalDNSDomain string
	var dnsmasqConfigPath string
	var dnsmasqHostsFilePath string

	var serviceMesh string
	serviceMeshFlag, serviceMeshOptions := servicemesh.Flag(&serviceMesh)

	command := &cli.Command{
		Name: "dns-controller",
		Flags: cli.Flags{
			&flags.String{
				Name:   "kubeconfig",
				Usage:  "path to kubeconfig file",
				Target: &kubeconfig,
			},
			&flags.String{
				Name:     "namespace-prefix",
				Usage:    "namespace prefix for the lattice",
				Required: true,
				Target:   &namespacePrefix,
			},
			&flags.String{
				Name:     "lattice-id",
				Usage:    "ID of the lattice",
				Required: true,
				Target:   &latticeID,
			},
			&flags.String{
				Name:     "internal-dns-domain",
				Usage:    "domain to use for internal dns",
				Required: true,
				Target:   &internalDNSDomain,
			},
			&flags.String{
				Name:    "dnsmasq-config-path",
				Usage:   "path to the additional dnsmasq configuration file",
				Default: "/var/run/lattice/dnsmasq.conf",
				Target:  &dnsmasqConfigPath,
			},
			&flags.String{
				Name:    "dnsmasq-hosts-file-path",
				Usage:   "path to the additional dnsmasq hosts",
				Default: "/var/run/lattice/hosts",
				Target:  &dnsmasqHostsFilePath,
			},

			&flags.String{
				Name:     "service-mesh",
				Required: true,
				Target:   &serviceMesh,
				Usage:    "service mesh to use",
			},
			serviceMeshFlag,
		},
		Run: func(args []string) {
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
				v1.LatticeID(latticeID),
				namespacePrefix,
				internalDNSDomain,
				dnsmasqConfigPath,
				dnsmasqHostsFilePath,
				serviceMeshOptions,
				latticeClient,
				kubeclientset.NewForConfigOrDie(config),
				latticeInformers.Lattice().V1().Configs(),
				latticeInformers.Lattice().V1().Addresses(),
				latticeInformers.Lattice().V1().Services(),
			).Run(stop)

			glog.V(1).Info("Starting informer factory")
			latticeInformers.Start(stop)

			select {}
		},
	}

	return command
}
