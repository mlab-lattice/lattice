package kubernetes

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/system/pkg/types"
)

func SystemID(namespace string) (types.SystemID, error) {
	parts := strings.Split(namespace, "-")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected system namespace format: %v", namespace)
	}

	return types.SystemID(strings.Join(parts[2:], "-")), nil
}
