package app

import (
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local/dns/controller"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
)

func Command() *cli.RootCommand {
	var (
		kubeconfig           string
		namespacePrefix      string
		latticeID            string
		internalDNSDomain    string
		dnsmasqConfigPath    string
		dnsmasqHostsFilePath string

		serviceMesh string
	)
	serviceMeshFlag, serviceMeshOptions := servicemesh.Flag(&serviceMesh)

	command := &cli.RootCommand{
		Name: "dns-controller",
		Command: &cli.Command{
			Flags: cli.Flags{
				"kubeconfig": &flags.String{
					Usage:  "path to kubeconfig file",
					Target: &kubeconfig,
				},
				"namespace-prefix": &flags.String{
					Usage:    "namespace prefix for the lattice",
					Required: true,
					Target:   &namespacePrefix,
				},
				"lattice-id": &flags.String{
					Usage:    "ID of the lattice",
					Required: true,
					Target:   &latticeID,
				},
				"internal-dns-domain": &flags.String{
					Usage:    "domain to use for internal dns",
					Required: true,
					Target:   &internalDNSDomain,
				},
				"dnsmasq-config-path": &flags.String{
					Usage:   "path to the additional dnsmasq configuration file",
					Default: "/var/run/lattice/dnsmasq.conf",
					Target:  &dnsmasqConfigPath,
				},
				"dnsmasq-hosts-file-path": &flags.String{
					Usage:   "path to the additional dnsmasq hosts",
					Default: "/var/run/lattice/hosts",
					Target:  &dnsmasqHostsFilePath,
				},

				"service-mesh": &flags.String{
					Required: true,
					Target:   &serviceMesh,
					Usage:    "service mesh to use",
				},
				"service-mesh-var": serviceMeshFlag,
			},
			Run: func(args []string, flags cli.Flags) error {
				var config *rest.Config
				var err error
				if kubeconfig == "" {
					config, err = rest.InClusterConfig()
				} else {
					config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
				}
				if err != nil {
					return err
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
		},
	}

	return command
}
