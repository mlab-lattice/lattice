package kubernetes

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/system/pkg/types"
)

func ClusterNamespace(clusterID types.ClusterID, namespace string) string {
	return fmt.Sprintf("%v-%v", clusterID, namespace)
}

func InternalNamespace(clusterID types.ClusterID) string {
	return ClusterNamespace(clusterID, kubeconstants.NamespaceLatticeInternal)
}

func SystemNamespace(clusterID types.ClusterID, systemID types.SystemID) string {
	return ClusterNamespace(clusterID, fmt.Sprintf("%v%v", kubeconstants.NamespacePrefixLatticeSystem, systemID))
}
