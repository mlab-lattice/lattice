package main

import (
	"flag"

	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app"
	"github.com/mlab-lattice/system/pkg/types"
)

var (
	kubeconfig          string
	clusterIDString     string
	provider            string
	terraformModulePath string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&clusterIDString, "cluster-id", "", "id of the cluster")
	flag.StringVar(&provider, "provider", "", "provider to use")
	flag.StringVar(&terraformModulePath, "terraform-module-path", "/etc/terraform/modules", "path to terraform modules")
	flag.Parse()
}

func main() {
	clusterID := types.ClusterID(clusterIDString)
	app.Run(clusterID, kubeconfig, provider, terraformModulePath)
}
