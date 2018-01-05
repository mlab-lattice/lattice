package constants

const (
	MasterNodeDNSSController = "local-dns-controller"
	MasterNodeDNSServer      = "local-dnsmasq-server"
	MasterNodeDNSService     = "local-dns-service"

	DockerImageLocalDNSController = "kubernetes-local-dns"
	DockerImageLocalDNSServer     = "gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.7"

	LocalDNSServerIP = "10.96.0.53"

	DNSSharedConfigDirectory = "/etc/dns-config/"
	DNSHostsFile             = "hosts"
	DnsmasqConfigFile        = "dnsmasq.conf"
)
