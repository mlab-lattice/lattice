package main

import (
	"flag"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/managerapi/server/user/backend"
	"github.com/mlab-lattice/system/pkg/managerapi/server/rest"
)

var (
	kubeconfig       string
	port             int
	workingDirectory string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&workingDirectory, "workingDirectory", "/tmp/lattice-manager-api", "working directory to use")
	flag.IntVar(&port, "port", 8080, "port to bind to")
	flag.Parse()
}

func main() {
	kubernetesBackend, err := backend.NewKubernetesBackend(kubeconfig)
	if err != nil {
		panic(err)
	}

	rest.RunNewRestServer(kubernetesBackend, int32(port), workingDirectory)
}
