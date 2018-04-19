package endpoint

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func DNSName(domain string, systemID v1.SystemID) string {
	// FIXME: make lattice.local configurable
	return fmt.Sprintf("%v.local.%v.lattice.local", domain, systemID)
}
