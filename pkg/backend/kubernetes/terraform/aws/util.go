package aws

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
)

func GetS3BackendStatePathRoot(clusterID types.ClusterID, systemID types.SystemID) string {
	return fmt.Sprintf("lattice/%v/terraform/state/%v", clusterID, systemID)
}
