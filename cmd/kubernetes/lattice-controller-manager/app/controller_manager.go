package app

import (
	goflag "flag"
	"fmt"
	"os"
	"time"

	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/basecontrollers"
	localcontrollers "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/cloudcontrollers/local"
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	kubeconfig          string
	clusterIDString     string
	cloudProviderName   string
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
		ctx, err := CreateControllerContext(clusterID, config, nil, terraformModulePath)
		if err != nil {
			panic(err)
		}

		glog.V(1).Info("Starting controllers")
		StartControllers(ctx, GetControllerInitializers(cloudProviderName))

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

	// https://flowerinthenight.com/blog/2017/12/01/golang-cobra-glog
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	RootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	RootCmd.Flags().StringVar(&clusterIDString, "cluster-id", "", "id of the cluster")
	RootCmd.MarkFlagRequired("cluster-id")
	RootCmd.Flags().StringVar(&cloudProviderName, "cloud-provider", "", "cloud provider that lattice is being run on")
	RootCmd.MarkFlagRequired("cloud-provider")
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
) (controller.Context, error) {
	cloudProvider, err := cloudprovider.NewCloudProvider(clusterID, cloudProviderName, &crv1.ConfigCloudProvider{})
	if err != nil {
		return controller.Context{}, err
	}

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

	ctx := controller.Context{
		ClusterID:     clusterID,
		CloudProvider: cloudProvider,

		KubeInformerFactory:    kubeInformers,
		LatticeInformerFactory: latticeInformers,
		KubeClientBuilder:      kcb,
		LatticeClientBuilder:   lcb,

		Stop: stop,

		TerraformModulePath: terraformModulePath,
	}
	return ctx, nil
}

func GetControllerInitializers(provider string) map[string]controller.Initializer {
	initializers := map[string]controller.Initializer{}

	for name, initializer := range basecontrollers.GetControllerInitializers() {
		initializers["base-"+name] = initializer
	}

	switch provider {
	case constants.ProviderAWS:
		// nothing for aws yet

	case constants.ProviderLocal:
		for name, initializer := range localcontrollers.GetControllerInitializers() {
			initializers["cloud-local-"+name] = initializer
		}
	default:
		panic("unsupported cloud provider " + provider)
	}

	return initializers
}

func StartControllers(ctx controller.Context, initializers map[string]controller.Initializer) {
	for controllerName, initializer := range initializers {
		glog.V(1).Infof("Starting %q", controllerName)
		initializer(ctx)
	}
}
