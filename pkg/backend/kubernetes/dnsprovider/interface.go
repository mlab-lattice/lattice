package dnsprovider

import "github.com/mlab-lattice/lattice/pkg/api/v1"

type DNS interface {
	ProvisionDNSARecord(latticeID v1.LatticeID, name, value string) error
	ProvisionDNSCNAMERecord(latticeID v1.LatticeID, name, value string) error

	DeprovisionDNSARecord(latticeID v1.LatticeID, name string) error
	DeprovisionDNSCNAMERecord(latticeID v1.LatticeID, name string) error
}
