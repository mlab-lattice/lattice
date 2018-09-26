package app

import (
	goflag "flag"
	"time"

	"github.com/mlab-lattice/lattice/cmd/kubernetes/controller-manager/app/controllers"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	kuberesolver "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
)

func Command() *cli.RootCommand {
	// Need to do a little hacking to make glog play nice
	// https://flowerinthenight.com/blog/2017/12/01/golang-cobra-glog
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	goflag.CommandLine.Parse([]string{})

	var (
		kubeconfig        string
		namespacePrefix   string
		workDirectory     string
		latticeID         string
		internalDNSDomain string

		enabledControllers []string

		cloudProvider string
		serviceMesh   string
	)

	cloudProviderFlag, cloudProviderOptions := cloudprovider.Flag(&cloudProvider)
	serviceMeshFlag, serviceMeshOptions := servicemesh.Flag(&serviceMesh)

	command := &cli.RootCommand{
		Name: "lattice-controller-manager",
		Command: &cli.Command{
			Flags: cli.Flags{
				"kubeconfig": &flags.String{
					Usage:  "path to kubeconfig file",
					Target: &kubeconfig,
				},
				"namespace-prefix": &flags.String{
					Usage:    "namespace prefix of the lattice",
					Required: true,
					Target:   &namespacePrefix,
				},
				"work-directory": &flags.String{
					Usage:   "work directory to use",
					Default: "/tmp/lattice-api",
					Target:  &workDirectory,
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
				"controllers": &flags.StringSlice{
					Usage:   "controllers that should be run",
					Default: []string{"*"},
					Target:  &enabledControllers,
				},

				"cloud-provider": &flags.String{
					Required: true,
					Target:   &cloudProvider,
					Usage:    "cloud provider that the kubernetes cluster is running on",
				},
				"cloud-provider-var": cloudProviderFlag,

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

				// TODO: setting stop as nil for now, won't actually need it until leader-election is used
				ctx, err := createControllerContext(
					namespacePrefix,
					workDirectory,
					v1.LatticeID(latticeID),
					internalDNSDomain,
					config,
					cloudProviderOptions,
					serviceMeshOptions,
					nil,
				)
				if err != nil {
					return err
				}

				glog.V(1).Info("Starting enabled controllers")
				startControllers(ctx, enabledControllers)

				glog.V(4).Info("Starting informer factory kubeinformers")
				ctx.KubeInformerFactory.Start(ctx.Stop)
				ctx.LatticeInformerFactory.Start(ctx.Stop)

				select {}
			},
		},
	}

	return command
}

func createControllerContext(
	namespacePrefix string,
	workDirectory string,
	latticeID v1.LatticeID,
	internalDNSDomain string,
	kubeconfig *rest.Config,
	cloudProviderOptions *cloudprovider.Options,
	serviceMeshOptions *servicemesh.Options,
	stop <-chan struct{},
) (controllers.Context, error) {
	kcb := controllers.KubeClientBuilder{
		Kubeconfig: kubeconfig,
	}
	lcb := controllers.LatticeClientBuilder{
		Kubeconfig: kubeconfig,
	}

	versionedKubeClient := kcb.ClientOrDie("shared-kubeinformers")
	kubeInformers := kubeinformers.NewSharedInformerFactory(versionedKubeClient, time.Duration(12*time.Hour))

	versionedLatticeClient := lcb.ClientOrDie("shared-latticeinformers")
	latticeInformers := latticeinformers.NewSharedInformerFactory(versionedLatticeClient, time.Duration(12*time.Hour))

	templateStore := kuberesolver.NewKubernetesTemplateStore(namespacePrefix, lcb.ClientOrDie("controller-manager-component-resolver"), latticeInformers, nil)
	secretStore := kuberesolver.NewKubernetesSecretStore(namespacePrefix, kubeInformers, nil)
	gitResolver, err := git.NewResolver(workDirectory, false)
	if err != nil {
		return controllers.Context{}, err
	}

	r := resolver.NewComponentResolver(gitResolver, templateStore, secretStore)
	ctx := controllers.Context{
		NamespacePrefix: namespacePrefix,
		LatticeID:       latticeID,

		ComponentResolver: r,

		InternalDNSDomain: internalDNSDomain,

		CloudProviderOptions: cloudProviderOptions,
		ServiceMeshOptions:   serviceMeshOptions,

		KubeInformerFactory:    kubeInformers,
		LatticeInformerFactory: latticeInformers,
		KubeClientBuilder:      kcb,
		LatticeClientBuilder:   lcb,

		Stop: stop,
	}
	return ctx, nil
}

func startControllers(ctx controllers.Context, enabledControllers []string) {
	for controllerName, initializer := range controllers.Initializers {
		if !controllerEnabled(controllerName, enabledControllers) {
			glog.V(1).Infof("not starting %q", controllerName)
			continue
		}

		glog.V(1).Infof("starting %q", controllerName)
		initializer(ctx)
	}
}

// similar to https://github.com/kubernetes/kubernetes/blob/v1.10.1/cmd/kube-controller-manager/app/controllermanager.go#L251
func controllerEnabled(name string, enabledControllers []string) bool {
	hasStar := false
	for _, ctrl := range enabledControllers {
		if ctrl == name {
			return true
		}
		if ctrl == "-"+name {
			return false
		}
		if ctrl == "*" {
			hasStar = true
		}
	}

	// if we get here, there was no explicit choice
	if !hasStar {
		// nothing on by default
		return false
	}

	return true
}
