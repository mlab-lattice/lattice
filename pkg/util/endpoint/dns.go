package endpoint

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func DNSName(domain string, systemID v1.SystemID, latticeID v1.LatticeID) string {
	return fmt.Sprintf("%v.local.%v.%v.local", domain, systemID, latticeID)
}
