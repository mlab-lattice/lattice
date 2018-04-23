package local

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func (cp *DefaultLocalCloudProvider) EnsureDNSARecord(latticeID v1.LatticeID, name, value string) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) EnsureDNSCNAMERecord(latticeID v1.LatticeID, name, value string) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) DestroyDNSARecord(latticeID v1.LatticeID, name string) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) DestroyDNSCNAMERecord(latticeID v1.LatticeID, name string) error {
	return nil
}
