package constants

const (
	DockerImageBootstrapKubernetes      = "kubernetes-bootstrap-lattice"
	DockerImageComponentBuilder         = "kubernetes-component-builder"
	DockerImageEnvoyXDSAPIRestPerNode   = "kubernetes-envoy-xds-api-rest-per-node"
	DockerImageLatticeControllerManager = "kubernetes-lattice-controller-manager"
	DockerImageManagerAPIRest           = "kubernetes-manager-api-rest"
	DockerImageLocalDNSController       = "lattice-local-dns"

	// TODO :: How to organize full name when not a lattice iamge. See lifecycle/provisioner/local/
	DockerImageLocalDNSServer = "gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.5"

	DockerRegistryComponentBuildsDefault = "component-builds"

	EnvVarNameDockerAPIVersion = "DOCKER_API_VERSION"
)
