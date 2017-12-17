package app

import (
	goflag "flag"
	"fmt"
	"os"
	"time"

	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/basecontrollers"
	awscontrollers "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/cloudcontrollers/aws"
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var (
	kubeconfig          string
	clusterIDString     string
	provider            string
	terraformModulePath string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "lattice-controller-manager",
	Run: func(cmd *cobra.Command, args []string) {
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

		clusterID := types.ClusterID(clusterIDString)

		// TODO: setting stop as nil for now, won't actually need it until leader-election is used
		ctx := CreateControllerContext(clusterID, config, nil, terraformModulePath)
		glog.V(1).Info("Starting controllers")
		StartControllers(ctx, GetControllerInitializers(provider))

		glog.V(4).Info("Starting informer factory kubeinformers")
		ctx.KubeInformerFactory.Start(ctx.Stop)
		ctx.LatticeInformerFactory.Start(ctx.Stop)

		select {}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initCmd)

	RootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	RootCmd.Flags().StringVar(&clusterIDString, "cluster-id", "", "id of the cluster")
	RootCmd.MarkFlagRequired("cluster-id")
	RootCmd.Flags().StringVar(&provider, "provider", "", "provider to use")
	RootCmd.MarkFlagRequired("provider")
	RootCmd.Flags().StringVar(&terraformModulePath, "terraform-module-path", "/etc/terraform/modules", "path to terraform modules")
}

func initCmd() {
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	goflag.CommandLine.Parse([]string{})
}

func CreateControllerContext(
	clusterID types.ClusterID,
	kubeconfig *rest.Config,
	stop <-chan struct{},
	terraformModulePath string,
) controller.Context {
	kcb := controller.KubeClientBuilder{
		Kubeconfig: kubeconfig,
	}
	lcb := controller.LatticeClientBuilder{
		Kubeconfig: kubeconfig,
	}

	versionedKubeClient := kcb.ClientOrDie("shared-kubeinformers")
	kubeInformers := kubeinformers.NewSharedInformerFactory(versionedKubeClient, time.Duration(12*time.Hour))

	versionedLatticeClient := lcb.ClientOrDie("shared-latticeinformers")
	latticeInformers := latticeinformers.NewSharedInformerFactory(versionedLatticeClient, time.Duration(12*time.Hour))

	return controller.Context{
		ClusterID: clusterID,

		KubeInformerFactory:    kubeInformers,
		LatticeInformerFactory: latticeInformers,
		KubeClientBuilder:      kcb,
		LatticeClientBuilder:   lcb,

		Stop: stop,

		TerraformModulePath: terraformModulePath,
	}
}

func GetControllerInitializers(provider string) map[string]controller.Initializer {
	initializers := map[string]controller.Initializer{}

	for name, initializer := range basecontrollers.GetControllerInitializers() {
		initializers["base-"+name] = initializer
	}

	switch provider {
	case constants.ProviderAWS:
		for name, initializer := range awscontrollers.GetControllerInitializers() {
			initializers["cloud-aws-"+name] = initializer
		}
	case constants.ProviderLocal:
		// Local case doesn't need any extra controllers
	default:
		panic("unsupported provider " + provider)
	}

	return initializers
}

func StartControllers(ctx controller.Context, initializers map[string]controller.Initializer) {
	for controllerName, initializer := range initializers {
		glog.V(1).Infof("Starting %q", controllerName)
		initializer(ctx)
	}
}
