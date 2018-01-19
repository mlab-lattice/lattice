package endpoint

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
)

func DNSName(domain string, systemID types.SystemID, clusterID types.ClusterID) string {
	return fmt.Sprintf("%v.local.%v.%v.local", domain, systemID, clusterID)
}
