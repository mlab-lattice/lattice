package main

import (
	"flag"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/managerapi/server/user/backend"
	"github.com/mlab-lattice/system/pkg/managerapi/server/rest"
	"github.com/mlab-lattice/system/pkg/types"
)

var (
	kubeconfig       string
	clusterIDString  string
	port             int
	workingDirectory string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&clusterIDString, "cluster-id", "", "id of the lattice cluster")
	flag.StringVar(&workingDirectory, "workingDirectory", "/tmp/lattice-manager-api", "working directory to use")
	flag.IntVar(&port, "port", 8080, "port to bind to")
	flag.Parse()
}

func main() {
	clusterID := types.ClusterID(clusterIDString)

	kubernetesBackend, err := backend.NewKubernetesBackend(clusterID, kubeconfig)
	if err != nil {
		panic(err)
	}

	rest.RunNewRestServer(kubernetesBackend, int32(port), workingDirectory)
}
