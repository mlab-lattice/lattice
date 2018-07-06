package dnsprovider

import "github.com/mlab-lattice/lattice/pkg/api/v1"

type Interface interface {
	DNSARecordNeedsUpdate(latticeID v1.LatticeID, name, value string) (bool, error)
	DNSARecordsNeedUpdate(latticeID v1.LatticeID, name string, value []string) (bool, error)

	EnsureDNSARecord(latticeID v1.LatticeID, name, value string) error
	EnsureDNSARecords(latticeID v1.LatticeID, name string, value []string) error

	EnsureDNSCNAMERecord(latticeID v1.LatticeID, name, value string) error

	DestroyDNSRecord(latticeID v1.LatticeID, name string) error
}
