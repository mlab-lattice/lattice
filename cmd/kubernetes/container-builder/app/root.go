package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubecontainerbuilder "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/containerbuilder"
	"github.com/mlab-lattice/lattice/pkg/containerbuilder"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/aws"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"github.com/spf13/cobra"
)

const (
	sshAuthSockEnvVarName   = "SSH_AUTH_SOCK"
	gitRepoSSHKeyEnvVarName = "GIT_REPO_SSH_KEY"
)

var (
	workDirectory    string
	containerBuildID string
	systemIDString   string
	namespacePrefix  string

	dockerRegistry         string
	dockerRegistryAuthType string
	dockerRepository       string
	dockerTag              string
	dockerPush             bool

	kubeconfig               string
	containerBuildDefinition string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "bootstrap-lattice",
	Short: "Bootstraps a kubernetes cluster to run lattice",
	Run: func(cmd *cobra.Command, args []string) {
		cb := &definitionv1.ContainerBuild{}
		err := json.Unmarshal([]byte(containerBuildDefinition), cb)
		if err != nil {
			log.Fatal("error unmarshaling container build: " + err.Error())
		}

		dockerOptions := &containerbuilder.DockerOptions{
			Registry:   dockerRegistry,
			Repository: dockerRepository,
			Tag:        dockerTag,
			Push:       dockerPush,
		}

		if dockerRegistryAuthType == aws.EC2RoleDockerRegistryAuth {
			dockerOptions.RegistryAuthProvider = &aws.ECRRegistryAuthProvider{}
		}

		systemID := v1.SystemID(systemIDString)

		statusUpdater, err := kubecontainerbuilder.NewKubernetesStatusUpdater(namespacePrefix, kubeconfig)
		if err != nil {
			log.Fatal("error getting status updater: " + err.Error())
		}

		gitRepoSSHKey := os.Getenv(gitRepoSSHKeyEnvVarName)
		var gitResolverOptions *git.Options
		if gitRepoSSHKey != "" {
			gitResolverOptions = &git.Options{
				SSHKey: []byte(gitRepoSSHKey),
			}
		}

		setupSSH()

		builder, err := containerbuilder.NewBuilder(
			v1.ContainerBuildID(containerBuildID),
			systemID,
			workDirectory,
			dockerOptions,
			gitResolverOptions,
			statusUpdater,
		)
		if err != nil {
			log.Fatal("error getting builder: " + err.Error())
		}

		if err = builder.Build(cb); err != nil {
			os.Exit(1)
		}
	},
}

func setupSSH() {
	// Get the SSH_AUTH_SOCK.
	// This probably isn't the best way of going about it.
	// First tried "eval ssh-agent > /dev/null && echo $SSH_AUTH_SOCK"
	// but since the subcommand isn't executed in a shell, this obviously didn't work.
	out, err := exec.Command("/usr/bin/ssh-agent", "-c").Output()
	if err != nil {
		log.Fatal("error setting up ssh-agent: " + err.Error())
	}

	// This expects the output to look like:
	// setenv SSH_AUTH_SOCK <file>;
	// ...
	lines := strings.Split(string(out), "\n")
	sshAuthSockSplit := strings.Split(lines[0], " ")
	sshAuthSock := strings.Split(sshAuthSockSplit[2], ";")[0]
	os.Setenv(sshAuthSockEnvVarName, sshAuthSock)

	out, err = exec.Command("/usr/bin/ssh-keyscan", "github.com").Output()
	if err != nil {
		log.Fatal("error setting up ssh-agent: " + err.Error())
	}

	err = ioutil.WriteFile("/etc/ssh/ssh_known_hosts", out, 0666)
	if err != nil {
		log.Fatal("error writing /etc/ssh/ssh_known_hosts: " + err.Error())
	}
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
	RootCmd.Flags().StringVar(&containerBuildID, "container-build-id", "", "ID of the container build")
	RootCmd.MarkFlagRequired("container-build-id")
	RootCmd.Flags().StringVar(&namespacePrefix, "namespace-prefix", "", "namespace prefix of the lattice")
	RootCmd.MarkFlagRequired("lattice-id")
	RootCmd.Flags().StringVar(&systemIDString, "system-id", "", "ID of the system")
	RootCmd.MarkFlagRequired("system-id")
	RootCmd.Flags().StringVar(&containerBuildDefinition, "container-build-definition", "", "JSON serialized version of the container build definition block")
	RootCmd.MarkFlagRequired("container-build-definition")

	RootCmd.Flags().StringVar(&dockerRegistry, "docker-registry", "", "registry to tag the docker image artifact with")
	RootCmd.MarkFlagRequired("docker-registry")
	RootCmd.Flags().StringVar(&dockerRegistryAuthType, "docker-registry-auth-type", "", "information about how to auth to the docker registry")
	RootCmd.Flags().StringVar(&dockerRepository, "docker-repository", "", "repository to tag the docker image artifact with")
	RootCmd.MarkFlagRequired("docker-repository")
	RootCmd.Flags().StringVar(&dockerTag, "docker-tag", "", "tag to tag the docker image artifact with")
	RootCmd.MarkFlagRequired("docker-tag")
	RootCmd.Flags().BoolVar(&dockerPush, "docker-push", false, "whether or not the image should be pushed to the registry")

	RootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig")
	RootCmd.Flags().StringVar(&workDirectory, "work-directory", "/tmp/container-build", "path to use to store build artifacts")
}
