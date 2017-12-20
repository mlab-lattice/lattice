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
	// TODO :: Should make this clear that this is required to be additional config whose purpose is jusst to contain multiple cname=... fields.
	flag.StringVar(&serverConfigPath, "server-config-path", "/etc/dnsmasq.conf", "path to the additional dnsmasq server configuration")
	flag.StringVar(&resolvConfPath, "resolv-path", "/etc/dnsmasq.resolv.conf", "path to the dnsmasq nameserver file")
	flag.Parse()
}

func main() {
	Run(clusterIDString, kubeconfig, provider, terraformModulePath, serverConfigPath, resolvConfPath)
}
