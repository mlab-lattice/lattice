package terraform

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func GetS3BackendSystemStatePathRoot(latticeID v1.LatticeID, systemID v1.SystemID) string {
	return fmt.Sprintf("%v/system/%v/terraform/state", GetS3BackendStatePathRoot(latticeID), systemID)
}

func GetS3BackendNodePoolPathRoot(latticeID v1.LatticeID, nodePoolID string) string {
	return fmt.Sprintf("%v/node-pool/terraform/state", GetS3BackendStatePathRoot(latticeID))
}

func GetS3BackendStatePathRoot(latticeID v1.LatticeID) string {
	return fmt.Sprintf("lattice/%v", latticeID)
}
