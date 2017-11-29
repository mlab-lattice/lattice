package app

import (
	"encoding/json"
	"fmt"
	"os"

	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"

	"github.com/mlab-lattice/system/pkg/componentbuild"

	"github.com/spf13/cobra"
)

var (
	workDirectory  string
	componentBuild string

	dockerRegistry   string
	dockerRepository string
	dockerTag        string
	dockerPush       bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "bootstrap-lattice",
	Short: "Bootstraps a kubernetes cluster to run lattice",
	Run: func(cmd *cobra.Command, args []string) {
		cb := &systemdefinitionblock.ComponentBuild{}
		err := json.Unmarshal([]byte(componentBuild), cb)
		if err != nil {
			panic("error unmarshaling component build: " + err.Error())
		}

		dockerOptions := &componentbuild.DockerOptions{
			Registry:   dockerRegistry,
			Repository: dockerRepository,
			Tag:        dockerTag,
			Push:       dockerPush,
		}

		builder, err := componentbuild.NewBuilder(workDirectory, dockerOptions, nil, cb, nil)
		if err != nil {
			panic("error getting builder: " + err.Error())
		}

		if err = builder.Build(); err != nil {
			panic("error building: " + err.Error())
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
	RootCmd.Flags().StringVar(&workDirectory, "work-directory", "/tmp/lattice-system/", "path to use to store build artifacts")
	RootCmd.Flags().StringVar(&componentBuild, "component-build", "", "JSON serialized version of the component build definition block")
	RootCmd.Flags().StringVar(&dockerRegistry, "docker-registry", "", "registry to tag the docker image artifact with")
	RootCmd.Flags().StringVar(&dockerRepository, "docker-repository", "", "repository to tag the docker image artifact with")
	RootCmd.Flags().StringVar(&dockerTag, "docker-tag", "", "tag to tag the docker image artifact with")
	RootCmd.Flags().BoolVar(&dockerPush, "docker-push", false, "whether or not the image should be pushed to the registry")
}
