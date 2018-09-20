package app

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"encoding/json"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubecontainerbuilder "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/containerbuilder"
	"github.com/mlab-lattice/lattice/pkg/containerbuilder"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/aws"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

const (
	sshAuthSockEnvVarName   = "SSH_AUTH_SOCK"
	gitRepoSSHKeyEnvVarName = "GIT_REPO_SSH_KEY"
)

func Command() *cli.RootCommand {
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

	return &cli.RootCommand{
		Name: "container-builder",
		Command: &cli.Command{

			Short: "builds and optionally pushes a container for a lattice container build",
			Flags: cli.Flags{
				"container-build-id": &flags.String{
					Usage:    "ID of the container build",
					Required: true,
					Target:   &containerBuildID,
				},
				"namespace-prefix": &flags.String{
					Usage:    "namespace prefix of the lattice",
					Required: true,
					Target:   &namespacePrefix,
				},
				"system-id": &flags.String{
					Usage:    "id of the system",
					Required: true,
					Target:   &systemIDString,
				},
				"container-build-definition": &flags.String{
					Usage:    "JSON serialized version of the container build definition block",
					Required: true,
					Target:   &containerBuildDefinition,
				},
				"docker-registry": &flags.String{
					Usage:    "registry to tag the docker image artifact with",
					Required: true,
					Target:   &dockerRegistry,
				},
				"docker-registry-auth-type": &flags.String{
					Usage:  "registry to tag the docker image artifact with",
					Target: &dockerRegistryAuthType,
				},
				"docker-repository": &flags.String{
					Usage:    "repository to tag the docker image artifact with",
					Required: true,
					Target:   &dockerRepository,
				},
				"docker-tag": &flags.String{
					Usage:    "tag to tag the docker image artifact with",
					Required: true,
					Target:   &dockerTag,
				},
				"docker-push": &flags.Bool{
					Usage:   "whether or not the image should be pushed to the registry",
					Default: false,
					Target:  &dockerPush,
				},
				"kubeconfig": &flags.String{
					Usage:  "path to kubeconfig",
					Target: &kubeconfig,
				},
				"work-directory": &flags.String{
					Usage:   "path to use to store build artifacts",
					Default: "/tmp/container-build",
					Target:  &workDirectory,
				},
			},
			Run: func(args []string, flags cli.Flags) error {
				cb := &definitionv1.ContainerBuild{}
				if err := json.Unmarshal([]byte(containerBuildDefinition), cb); err != nil {
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

				return builder.Build(cb)
			},
		},
	}
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
