package kubernetes

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/system/pkg/types"
)

func LatticeNamespace(latticeID types.LatticeID, namespace string) string {
	return fmt.Sprintf("%v-%v", latticeID, namespace)
}

func InternalNamespace(latticeID types.LatticeID) string {
	return LatticeNamespace(latticeID, kubeconstants.NamespaceLatticeInternal)
}

func SystemNamespace(latticeID types.LatticeID, systemID types.SystemID) string {
	return LatticeNamespace(latticeID, fmt.Sprintf("%v%v", kubeconstants.NamespacePrefixLatticeSystem, systemID))
}
