package kubernetes

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
)

func ComponentBuilderClusterRoleName(latticeID v1.LatticeID) string {
	return fmt.Sprintf("%v-%v", latticeID, constants.ControlPlaneServiceComponentBuilder)
}
