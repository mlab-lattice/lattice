package endpoint

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
)

func DNSName(domain string, systemID types.SystemID, latticeID types.LatticeID) string {
	return fmt.Sprintf("%v.local.%v.%v.local", domain, systemID, latticeID)
}
