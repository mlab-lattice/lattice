package main

import (
	"flag"
)

var (
	kubeconfig          string
	clusterIDString		string
	provider            string
	terraformModulePath string
	resolvConfPath		string
	serverConfigPath	string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&clusterIDString, "cluster-id", "", "id of the cluster")
	flag.StringVar(&provider, "provider", "", "provider to use")
	flag.StringVar(&terraformModulePath, "terraform-module-path", "/etc/terraform/modules", "path to terraform modules")
	flag.StringVar(&serverConfigPath, "server-config-path", "/etc/dnsmasq.conf", "path to the additional dnsmasq server configuration")
	flag.Parse()
}

func main() {
	Run(clusterIDString, kubeconfig, provider, terraformModulePath, serverConfigPath, resolvConfPath)
}
