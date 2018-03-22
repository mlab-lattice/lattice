package aws

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
)

func GetS3BackendSystemStatePathRoot(latticeID types.LatticeID, systemID types.SystemID) string {
	return fmt.Sprintf("%v/system/%v/terraform/state", GetS3BackendStatePathRoot(latticeID), systemID)
}

func GetS3BackendNodePoolPathRoot(latticeID types.LatticeID, nodePoolID string) string {
	return fmt.Sprintf("%v/node-pool/terraform/state", GetS3BackendStatePathRoot(latticeID))
}

func GetS3BackendStatePathRoot(latticeID types.LatticeID) string {
	return fmt.Sprintf("lattice/%v", latticeID)
}
