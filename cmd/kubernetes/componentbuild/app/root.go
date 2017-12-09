package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	kubecomponentbuild "github.com/mlab-lattice/system/pkg/backend/kubernetes/componentbuild"
	"github.com/mlab-lattice/system/pkg/componentbuild"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/definition/block"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/aws"

	"github.com/spf13/cobra"
)

var (
	workDirectory    string
	componentBuildID string

	dockerRegistry         string
	dockerRegistryAuthType string
	dockerRepository       string
	dockerTag              string
	dockerPush             bool

	kubeconfig               string
	componentBuildDefinition string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "bootstrap-lattice",
	Short: "Bootstraps a kubernetes cluster to run lattice",
	Run: func(cmd *cobra.Command, args []string) {
		cb := &block.ComponentBuild{}
		err := json.Unmarshal([]byte(componentBuildDefinition), cb)
		if err != nil {
			log.Fatal("error unmarshaling component build: " + err.Error())
		}

		dockerOptions := &componentbuild.DockerOptions{
			Registry:   dockerRegistry,
			Repository: dockerRepository,
			Tag:        dockerTag,
			Push:       dockerPush,
		}

		if dockerRegistryAuthType == constants.DockerRegistryAuthAWSEC2Role {
			dockerOptions.RegistryAuthProvider = &aws.ECRRegistryAuthProvider{}
		}

		statusUpdater, err := kubecomponentbuild.NewKubernetesStatusUpdater(kubeconfig)
		if err != nil {
			log.Fatal("error getting status updater: " + err.Error())
		}

		builder, err := componentbuild.NewBuilder(types.ComponentBuildID(componentBuildID), workDirectory, dockerOptions, nil, cb, statusUpdater)
		if err != nil {
			log.Fatal("error getting builder: " + err.Error())
		}

		if err = builder.Build(); err != nil {
			os.Exit(1)
		}
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
	RootCmd.Flags().StringVar(&componentBuildID, "component-build-id", "", "ID of the component build")
	RootCmd.Flags().StringVar(&componentBuildDefinition, "component-build-definition", "", "JSON serialized version of the component build definition block")

	RootCmd.Flags().StringVar(&dockerRegistry, "docker-registry", "", "registry to tag the docker image artifact with")
	RootCmd.Flags().StringVar(&dockerRegistryAuthType, "docker-registry-auth-type", "", "information about how to auth to the docker registry")
	RootCmd.Flags().StringVar(&dockerRepository, "docker-repository", "", "repository to tag the docker image artifact with")
	RootCmd.Flags().StringVar(&dockerTag, "docker-tag", "", "tag to tag the docker image artifact with")
	RootCmd.Flags().BoolVar(&dockerPush, "docker-push", false, "whether or not the image should be pushed to the registry")

	RootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig")
	RootCmd.Flags().StringVar(&workDirectory, "work-directory", "/tmp/component-build", "path to use to store build artifacts")
}
