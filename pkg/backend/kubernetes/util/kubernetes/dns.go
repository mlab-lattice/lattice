package kubernetes

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func InternalSystemSubdomain(systemID v1.SystemID, latticeID v1.LatticeID) string {
	return fmt.Sprintf("%v.%v", systemID, latticeID)
}

func InternalSubdomain(subdomain string, systemID v1.SystemID, latticeID v1.LatticeID) string {
	return fmt.Sprintf("%v.local.%v", subdomain, InternalSystemSubdomain(systemID, latticeID))
}

func FullyQualifiedInternalLatticeSubdomain(latticeID v1.LatticeID, internalDNSDomain string) string {
	return fmt.Sprintf("%v.%v", latticeID, internalDNSDomain)
}

func FullyQualifiedInternalSystemSubdomain(systemID v1.SystemID, latticeID v1.LatticeID, internalDNSDomain string) string {
	return fmt.Sprintf("%v.%v", systemID, FullyQualifiedInternalLatticeSubdomain(latticeID, internalDNSDomain))
}

func FullyQualifiedInternalSubdomain(subdomain string, systemID v1.SystemID, latticeID v1.LatticeID, internalDNSDomain string) string {
	return fmt.Sprintf("%v.%v", InternalSubdomain(subdomain, systemID, latticeID), internalDNSDomain)
}
