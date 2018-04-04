package kubernetes

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func SystemID(namespace string) (v1.SystemID, error) {
	parts := strings.Split(namespace, "-")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected system namespace format: %v", namespace)
	}

	return v1.SystemID(strings.Join(parts[2:], "-")), nil
}
