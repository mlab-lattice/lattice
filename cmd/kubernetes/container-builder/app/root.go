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
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

const (
	sshAuthSockEnvVarName   = "SSH_AUTH_SOCK"
	gitRepoSSHKeyEnvVarName = "GIT_REPO_SSH_KEY"
)

func Command() *cli.Command {
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

	return &cli.Command{
		Name:  "container-builder",
		Short: "builds and optionally pushes a container for a lattice container build",
		Flags: cli.Flags{
			&flags.String{
				Name:     "container-build-id",
				Usage:    "ID of the container build",
				Required: true,
				Target:   &containerBuildID,
			},
			&flags.String{
				Name:     "namespace-prefix",
				Usage:    "namespace prefix of the lattice",
				Required: true,
				Target:   &namespacePrefix,
			},
			&flags.String{
				Name:     "system-id",
				Usage:    "id of the system",
				Required: true,
				Target:   &systemIDString,
			},
			&flags.String{
				Name:     "container-build-definition",
				Usage:    "JSON serialized version of the container build definition block",
				Required: true,
				Target:   &containerBuildDefinition,
			},
			&flags.String{
				Name:     "docker-registry",
				Usage:    "registry to tag the docker image artifact with",
				Required: true,
				Target:   &dockerRegistry,
			},
			&flags.String{
				Name:   "docker-registry-auth-type",
				Usage:  "registry to tag the docker image artifact with",
				Target: &dockerRegistryAuthType,
			},
			&flags.String{
				Name:     "docker-repository",
				Usage:    "repository to tag the docker image artifact with",
				Required: true,
				Target:   &dockerRepository,
			},
			&flags.String{
				Name:     "docker-tag",
				Usage:    "tag to tag the docker image artifact with",
				Required: true,
				Target:   &dockerTag,
			},
			&flags.Bool{
				Name:    "docker-push",
				Usage:   "whether or not the image should be pushed to the registry",
				Default: false,
				Target:  &dockerPush,
			},
			&flags.String{
				Name:   "kubeconfig",
				Usage:  "path to kubeconfig",
				Target: &kubeconfig,
			},
			&flags.String{
				Name:    "work-directory",
				Usage:   "path to use to store build artifacts",
				Default: "/tmp/container-build",
				Target:  &workDirectory,
			},
		},
		Run: func(args []string) {
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

			if err = builder.Build(cb); err != nil {
				os.Exit(1)
			}
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
