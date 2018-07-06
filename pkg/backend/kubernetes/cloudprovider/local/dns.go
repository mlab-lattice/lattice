package local

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func (cp *DefaultLocalCloudProvider) DNSARecordNeedsUpdate(latticeID v1.LatticeID, name, value string) (bool, error) {
	return false, nil
}

func (cp *DefaultLocalCloudProvider) DNSARecordsNeedUpdate(latticeID v1.LatticeID, name string, value []string) (bool, error) {
	return false, nil
}

func (cp *DefaultLocalCloudProvider) EnsureDNSARecord(latticeID v1.LatticeID, name, value string) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) EnsureDNSARecords(latticeID v1.LatticeID, name string, value []string) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) EnsureDNSCNAMERecord(latticeID v1.LatticeID, name, value string) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) DestroyDNSRecord(latticeID v1.LatticeID, name string) error {
	return nil
}
