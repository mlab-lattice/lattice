package app

import (
	goflag "flag"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/api/server/backend"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	kuberesolver "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/util/cli"

	kubeclientset "k8s.io/client-go/kubernetes"
	kuberest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/pflag"
)

func Command() *cli.Command {
	// https://flowerinthenight.com/blog/2017/12/01/golang-cobra-glog
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	var kubeconfig string
	var namespacePrefix string
	var workDirectory string
	var port int32
	var apiAuthKey string

	command := &cli.Command{
		Name: "api-server",
		Flags: cli.Flags{
			&cli.StringFlag{
				Name:   "kubeconfig",
				Usage:  "path to kubeconfig file",
				Target: &kubeconfig,
			},
			&cli.StringFlag{
				Name:     "namespace-prefix",
				Usage:    "namespace prefix of the lattice",
				Required: true,
				Target:   &namespacePrefix,
			},
			&cli.StringFlag{
				Name:    "work-directory",
				Usage:   "work directory to use",
				Default: "/tmp/lattice-api",
				Target:  &workDirectory,
			},
			&cli.Int32Flag{
				Name:    "port",
				Usage:   "port to bind to",
				Default: 8080,
				Target:  &port,
			},
			&cli.StringFlag{
				Name:    "api-auth-key",
				Usage:   "if supplied, the required value of the API_KEY header",
				Default: "",
				Target:  &apiAuthKey,
			},
		},
		Run: func(args []string) {
			// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
			goflag.CommandLine.Parse([]string{})

			var config *kuberest.Config
			var err error
			if kubeconfig == "" {
				config, err = kuberest.InClusterConfig()
			} else {
				config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			}
			if err != nil {
				panic(err)
			}

			kubeClient, err := kubeclientset.NewForConfig(config)
			if err != nil {
				panic(err)
			}

			latticeClient, err := latticeclientset.NewForConfig(config)
			if err != nil {
				panic(err)
			}

			backend := backend.NewKubernetesBackend(namespacePrefix, kubeClient, latticeClient)

			latticeInformers := latticeinformers.NewSharedInformerFactory(latticeClient, time.Duration(12*time.Hour))
			store := kuberesolver.NewKubernetesTemplateStore(namespacePrefix, latticeClient, latticeInformers, nil)
			resolver, err := resolver.NewComponentResolver(workDirectory, store)
			if err != nil {
				panic(err)
			}

			rest.RunNewRestServer(backend, resolver, port, apiAuthKey)
		},
	}

	return command
}
