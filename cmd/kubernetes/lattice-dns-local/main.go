package main

import (
	"flag"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
)

var (
	kubeconfig          string
	clusterIDString     string
	provider            string
	terraformModulePath string
	hostConfigPath      string
	serverConfigPath    string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&clusterIDString, "cluster-id", "", "id of the cluster")
	flag.StringVar(&provider, "provider", "", "provider to use")
	flag.StringVar(&terraformModulePath, "terraform-module-path", "/etc/terraform/modules", "path to terraform modules")
	flag.StringVar(&serverConfigPath, "server-config-path", kubeconstants.DNSSharedConfigDirectory + kubeconstants.DNSConfigFile, "path to the additional dnsmasq cname file")
	flag.StringVar(&hostConfigPath, "host-config-path", kubeconstants.DNSSharedConfigDirectory + kubeconstants.DNSHostsFile, "path to the additional dnsmasq hosts")
	flag.Parse()
}

func main() {
	Run(clusterIDString, kubeconfig, provider, terraformModulePath, serverConfigPath, hostConfigPath)
}
