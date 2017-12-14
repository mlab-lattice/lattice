package main

import (
	"flag"

	"github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app"
)

var (
	kubeconfig          string
	provider            string
	terraformModulePath string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&provider, "provider", "", "provider to use")
	flag.StringVar(&terraformModulePath, "terraform-module-path", "/etc/terraform/modules", "path to terraform modules")
	flag.Parse()
}

func main() {
	app.Run(kubeconfig, provider, terraformModulePath)
}