package aws

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubetf "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws/terraform"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"
	awstfprovider "github.com/mlab-lattice/lattice/pkg/util/terraform/provider/aws"
)

func (cp *DefaultAWSCloudProvider) EnsureDNSARecord(latticeID v1.LatticeID, name, value string) error {
	return cp.provisionRoute53Record(latticeID, "A", name, value)
}

func (cp *DefaultAWSCloudProvider) EnsureDNSCNAMERecord(latticeID v1.LatticeID, name, value string) error {
	return cp.provisionRoute53Record(latticeID, "CNAME", name, value)
}

func (cp *DefaultAWSCloudProvider) provisionRoute53Record(latticeID v1.LatticeID, recordType, name, value string) error {
	config := cp.route53RecordTerraformConfig(latticeID, cp.route53PrivateZoneID, recordType, name, value)
	_, err := terraform.Apply(route53RecordWorkDirectory(cp.route53PrivateZoneID, name), config)
	return err
}

func (cp *DefaultAWSCloudProvider) DestroyDNSARecord(latticeID v1.LatticeID, name string) error {
	return cp.deprovisionRoute53Record(latticeID, name)
}

func (cp *DefaultAWSCloudProvider) DestroyDNSCNAMERecord(latticeID v1.LatticeID, name string) error {
	return cp.deprovisionRoute53Record(latticeID, name)
}

func (cp *DefaultAWSCloudProvider) deprovisionRoute53Record(latticeID v1.LatticeID, name string) error {
	_, err := terraform.Destroy(route53RecordWorkDirectory(cp.route53PrivateZoneID, name), nil)
	return err
}

func (cp *DefaultAWSCloudProvider) route53RecordTerraformConfig(latticeID v1.LatticeID, zoneID, recordType, name, value string) *terraform.Config {
	return &terraform.Config{
		Provider: awstfprovider.Provider{
			Region: cp.region,
		},
		Backend: terraform.S3BackendConfig{
			Region: cp.region,
			Bucket: cp.terraformBackendOptions.S3.Bucket,
			Key: fmt.Sprintf(
				"%v/%v",
				kubetf.GetS3BackendRoute53PathRoot(latticeID, zoneID),
				name,
			),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"route53-record": kubetf.Route53Record{
				Source: cp.terraformModulePath + kubetf.ModulePathRoute53Record,

				Region: cp.region,

				ZoneID: zoneID,
				Type:   recordType,
				Name:   name,
				Value:  value,
			},
		},
	}
}

func route53RecordWorkDirectory(zoneID, name string) string {
	return workDirectory("route53", fmt.Sprintf("%v/%v", zoneID, name))
}
