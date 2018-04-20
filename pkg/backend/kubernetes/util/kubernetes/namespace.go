package kubernetes

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
)

func LatticeNamespace(namespacePrefix, namespace string) string {
	return fmt.Sprintf("%v-%v", namespacePrefix, namespace)
}

func InternalNamespace(namespacePrefix string) string {
	return LatticeNamespace(namespacePrefix, constants.NamespaceLatticeInternal)
}

func SystemNamespace(namespacePrefix string, systemID v1.SystemID) string {
	return LatticeNamespace(namespacePrefix, fmt.Sprintf("%v%v", constants.NamespacePrefixLatticeSystem, systemID))
}

func SystemID(namespacePrefix, namespace string) (v1.SystemID, error) {
	sep := SystemNamespace(namespacePrefix, "")
	parts := strings.Split(namespace, sep)
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected system namespace format: %v", namespace)
	}

	return v1.SystemID(strings.Join(parts[1:], sep)), nil
}
