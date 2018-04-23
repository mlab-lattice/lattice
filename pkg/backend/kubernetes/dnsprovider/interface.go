package dnsprovider

import "github.com/mlab-lattice/lattice/pkg/api/v1"

type Interface interface {
	EnsureDNSARecord(latticeID v1.LatticeID, name, value string) error
	EnsureDNSCNAMERecord(latticeID v1.LatticeID, name, value string) error

	DestroyDNSARecord(latticeID v1.LatticeID, name string) error
	DestroyDNSCNAMERecord(latticeID v1.LatticeID, name string) error
}
