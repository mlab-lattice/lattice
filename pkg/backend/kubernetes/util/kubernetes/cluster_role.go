package kubernetes

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
)

func ComponentBuilderClusterRoleName(latticeID v1.LatticeID) string {
	return fmt.Sprintf("%v-%v", latticeID, constants.ControlPlaneServiceComponentBuilder)
}
