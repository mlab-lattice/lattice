package kubernetes

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
)

func ContainerBuilderClusterRoleName(namespacePrefix string) string {
	return fmt.Sprintf("%v-%v", namespacePrefix, constants.ControlPlaneServiceContainerBuilder)
}
