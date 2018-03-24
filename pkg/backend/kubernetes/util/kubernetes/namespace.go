package kubernetes

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
)

func LatticeNamespace(latticeID v1.LatticeID, namespace string) string {
	return fmt.Sprintf("%v-%v", latticeID, namespace)
}

func InternalNamespace(latticeID v1.LatticeID) string {
	return LatticeNamespace(latticeID, kubeconstants.NamespaceLatticeInternal)
}

func SystemNamespace(latticeID v1.LatticeID, systemID v1.SystemID) string {
	return LatticeNamespace(latticeID, fmt.Sprintf("%v%v", kubeconstants.NamespacePrefixLatticeSystem, systemID))
}
