package app

import (
	goflag "flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/basecontrollers"
	awscontrollers "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/cloudcontrollers/aws"
	localcontrollers "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/cloudcontrollers/local"
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"
	"github.com/mlab-lattice/system/pkg/util/cli"
	"github.com/mlab-lattice/system/pkg/util/terraform"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	kubeconfig string
	latticeID  string

	cloudProviderName string
	cloudProviderVars []string

	serviceMeshProvider     string
	serviceMeshProviderVars []string

	networkingProviderName string
	networkingProviderVars []string

	terraformModulePath  string
	terraformBackend     string
	terraformBackendVars []string
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

		latticeID := v1.LatticeID(latticeID)

		cloudSystemBootstrapper, err := cloudprovider.SystemBootstrapperFromFlags(cloudProviderName, cloudProviderVars)
		if err != nil {
			panic(err)
		}

		serviceMeshSystemBootstrapper, err := servicemesh.SystemBootstrapperFromFlags(serviceMeshProvider, serviceMeshProviderVars)
		if err != nil {
			panic(err)
		}

		systemBoostrappers := []bootstrapper.Interface{
			serviceMeshSystemBootstrapper,
			cloudSystemBootstrapper,
		}

		// TODO: setting stop as nil for now, won't actually need it until leader-election is used
		ctx, err := CreateControllerContext(latticeID, systemBoostrappers, config, nil, terraformModulePath)
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
	RootCmd.Flags().StringVar(&latticeID, "lattice-id", "", "id of the lattice")
	RootCmd.MarkFlagRequired("lattice-id")

	RootCmd.Flags().StringVar(&cloudProviderName, "cloud-provider", "", "cloud provider that lattice is being run on")
	RootCmd.MarkFlagRequired("cloud-provider")
	RootCmd.Flags().StringArrayVar(&cloudProviderVars, "cloud-provider-var", nil, "additional variables for the cloud provider")

	RootCmd.Flags().StringVar(&serviceMeshProvider, "service-mesh", "", "service mesh provider to use")
	RootCmd.MarkFlagRequired("service-mesh")
	RootCmd.Flags().StringArrayVar(&serviceMeshProviderVars, "service-mesh-var", nil, "additional variables for the cloud provider")

	RootCmd.Flags().StringVar(&networkingProviderName, "networking-provider", "", "provider to use for networking")
	RootCmd.Flags().StringArrayVar(&networkingProviderVars, "networking-provider-var", nil, "additional variables for the networking provider")

	RootCmd.Flags().StringVar(&terraformModulePath, "terraform-module-path", "/etc/terraform/modules", "path to terraform modules")
	RootCmd.Flags().StringVar(&terraformBackend, "terraform-backend", "", "backend to use for terraform")
	RootCmd.Flags().StringArrayVar(&terraformBackendVars, "terraform-backend-var", nil, "additional variables for the terraform backend")
}

func initCmd() {
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	goflag.CommandLine.Parse([]string{})
}

func CreateControllerContext(
	latticeID v1.LatticeID,
	systemBootstrappers []bootstrapper.Interface,
	kubeconfig *rest.Config,
	stop <-chan struct{},
	terraformModulePath string,
) (controller.Context, error) {
	cloudProviderOptions, err := parseCloudProviderVars()
	if err != nil {
		return controller.Context{}, err
	}

	terraformBackendOptions, err := parseTerraformVars()
	if err != nil {
		return controller.Context{}, err
	}

	cloudProvider, err := cloudprovider.NewCloudProvider(cloudProviderOptions)
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
		LatticeID:     latticeID,
		CloudProvider: cloudProvider,

		SystemBootstrappers: systemBootstrappers,

		TerraformBackendOptions: terraformBackendOptions,

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
	case cloudprovider.AWS:
		for name, initializer := range awscontrollers.GetControllerInitializers() {
			initializers["cloud-local-"+name] = initializer
		}

	case cloudprovider.Local:
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

func parseCloudProviderVars() (*cloudprovider.Options, error) {
	var options *cloudprovider.Options
	switch cloudProviderName {
	case cloudprovider.Local:
		localOptions, err := parseCloudProviderVarsLocal()
		if err != nil {
			return nil, err
		}
		options = &cloudprovider.Options{
			Local: localOptions,
		}

	case cloudprovider.AWS:
		awsOptions, err := parseCloudProviderVarsAWS()
		if err != nil {
			return nil, err
		}
		options = &cloudprovider.Options{
			AWS: awsOptions,
		}

	default:
		return nil, fmt.Errorf("unsupported cloudProviderName: %v", cloudProviderName)
	}

	return options, nil
}

func parseCloudProviderVarsLocal() (*local.Options, error) {
	options := &local.Options{}
	flags := cli.EmbeddedFlag{
		Target: &options,
		Expected: map[string]cli.EmbeddedFlagValue{
			"ip": {
				Required:     true,
				EncodingName: "IP",
			},
		},
	}

	err := flags.Parse(cloudProviderVars)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func parseCloudProviderVarsAWS() (*aws.Options, error) {
	options := &aws.Options{}
	flags := cli.EmbeddedFlag{
		Target: &options,
		Expected: map[string]cli.EmbeddedFlagValue{
			"region": {
				Required:     true,
				EncodingName: "Region",
			},
			"account-id": {
				Required:     true,
				EncodingName: "AccountID",
			},
			"vpc-id": {
				Required:     true,
				EncodingName: "VPCID",
			},
			"route53-private-zone-id": {
				Required:     true,
				EncodingName: "Route53PrivateZoneID",
			},
			"subnet-ids": {
				Required:     true,
				EncodingName: "SubnetIDs",
				ValueParser: func(value string) (interface{}, error) {
					return strings.Split(value, ","), nil
				},
			},
			"master-node-security-group-id": {
				Required:     true,
				EncodingName: "MasterNodeSecurityGroupID",
			},
		},
	}

	err := flags.Parse(cloudProviderVars)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func parseTerraformVars() (*terraform.BackendOptions, error) {
	if terraformBackend == "" {
		return nil, nil
	}

	var backend *terraform.BackendOptions
	switch terraformBackend {
	case terraform.BackendS3:
		s3Config, err := parseTerraformVarsS3()
		if err != nil {
			return nil, err
		}
		backend = &terraform.BackendOptions{
			S3: s3Config,
		}
	default:
		return nil, fmt.Errorf("unsupported terraform backend: %v", terraformBackend)
	}

	return backend, nil
}

func parseTerraformVarsS3() (*terraform.BackendOptionsS3, error) {
	s3Config := &terraform.BackendOptionsS3{}
	flags := cli.EmbeddedFlag{
		Target: &s3Config,
		Expected: map[string]cli.EmbeddedFlagValue{
			"bucket": {
				EncodingName: "Bucket",
				Required:     true,
			},
		},
	}

	err := flags.Parse(terraformBackendVars)
	if err != nil {
		return nil, err
	}
	return s3Config, nil
}
