package app

import (
	goflag "flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/api/server/backend"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	kuberesolver "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"

	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	kuberest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mlab-lattice/lattice/pkg/util/git"
	"github.com/spf13/pflag"
)

const (
	sshAuthSockEnvVarName = "SSH_AUTH_SOCK"
)

func Command() *cli.RootCommand {
	// https://flowerinthenight.com/blog/2017/12/01/golang-cobra-glog
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	var kubeconfig string
	var namespacePrefix string
	var workDirectory string
	var port int32
	var apiAuthKey string

	command := &cli.RootCommand{
		Name: "api-server",
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
				"port": &flags.Int32{
					Usage:   "port to bind to",
					Default: 8080,
					Target:  &port,
				},
				"api-auth-key": &flags.String{
					Usage:   "if supplied, the required value of the API_KEY header",
					Default: "",
					Target:  &apiAuthKey,
				},
			},
			Run: func(args []string, flags cli.Flags) error {
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
					return err
				}

				kubeClient, err := kubeclientset.NewForConfig(config)
				if err != nil {
					return err
				}

				latticeClient, err := latticeclientset.NewForConfig(config)
				if err != nil {
					return err
				}

				setupSSH()

				backend := backend.NewKubernetesBackend(namespacePrefix, kubeClient, latticeClient)

				latticeInformers := latticeinformers.NewSharedInformerFactory(latticeClient, time.Duration(12*time.Hour))
				kubeInformers := kubeinformers.NewSharedInformerFactory(kubeClient, time.Duration(12*time.Hour))
				templateStore := kuberesolver.NewKubernetesTemplateStore(namespacePrefix, latticeClient, latticeInformers, nil)
				secretStore := kuberesolver.NewKubernetesSecretStore(namespacePrefix, kubeInformers, nil)
				gitResolver, err := git.NewResolver(workDirectory, false)
				if err != nil {
					return err
				}

				r := resolver.NewComponentResolver(gitResolver, templateStore, secretStore)
				rest.RunNewRestServer(backend, r, port, apiAuthKey)
				return nil
			},
		},
	}

	return command
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
