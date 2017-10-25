package kubernetes

import (
	"fmt"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

func GetKubeServiceNameForService(svc *crv1.Service) string {
	// This ensures that the kube Service name is a DNS-1035 label:
	// "a DNS-1035 label must consist of lower case alphanumeric characters or '-',
	//  and must start and end with an alphanumeric character (e.g. 'my-name',
	//  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')"
	return fmt.Sprintf("svc-%v-lattice", svc.Name)
}
