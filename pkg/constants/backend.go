package constants

const (
	BackendTypeKubernetes = "kubernetes"
)

var SupportedBackendTypes = map[string]struct{}{
	BackendTypeKubernetes: {},
}
