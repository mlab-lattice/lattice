package kubernetes

import (
	"fmt"
	"strings"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	corev1 "k8s.io/api/core/v1"
)

const (
	kubeServiceServicePrefix = "lattice-service-"
)

func GetKubeServiceNameForService(service *crv1.Service) string {
	// This ensures that the kube Service name is a DNS-1035 label:
	// "a DNS-1035 label must consist of lower case alphanumeric characters or '-',
	//  and must start and end with an alphanumeric character (e.g. 'my-name',
	//  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')"
	return fmt.Sprintf("%v%v", kubeServiceServicePrefix, service.Name)
}

func GetServiceNameForKubeService(kubeService *corev1.Service) (string, error) {
	parts := strings.Split(kubeService.Name, kubeServiceServicePrefix)
	if len(parts) != 2 {
		return "", fmt.Errorf("kube service name did not match expected naming convention")
	}

	return parts[1], nil
}
