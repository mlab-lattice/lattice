package aws

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubetf "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws/terraform"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"
	awstfprovider "github.com/mlab-lattice/lattice/pkg/util/terraform/provider/aws"
)

const (
	aRecordType     = "A"
	cnameRecordType = "CNAME"
)

func (cp *DefaultAWSCloudProvider) DNSARecordNeedsUpdate(latticeID v1.LatticeID, name, value string) (bool, error) {
	module := cp.route53RecordTerraformModule(latticeID, cp.route53PrivateZoneID, aRecordType, name, value)
	config := cp.route53RecordTerraformConfig(latticeID, cp.route53PrivateZoneID, name, module)
	result, _, err := terraform.Plan(route53RecordWorkDirectory(cp.route53PrivateZoneID, name), config, false)
	if err != nil {
		return false, err
	}

	switch result {
	case terraform.PlanResultError:
		return false, fmt.Errorf("unknown error")

	case terraform.PlanResultEmpty:
		return false, nil

	case terraform.PlanResultNotEmpty:
		return true, nil

	default:
		return false, fmt.Errorf("unexpected terraform plan result: %v", result)
	}
}

func (cp *DefaultAWSCloudProvider) DNSARecordsNeedUpdate(latticeID v1.LatticeID, name string, value []string) (bool, error) {
	// GEB: implement me
	return false, nil
}

func (cp *DefaultAWSCloudProvider) EnsureDNSARecord(latticeID v1.LatticeID, name, value string) error {
	return cp.provisionRoute53Record(latticeID, aRecordType, name, value)
}

func (cp *DefaultAWSCloudProvider) EnsureDNSARecords(latticeID v1.LatticeID, name string, value []string) error {
	// GEB: implement me
	return nil
}

func (cp *DefaultAWSCloudProvider) EnsureDNSCNAMERecord(latticeID v1.LatticeID, name, value string) error {
	return cp.provisionRoute53Record(latticeID, cnameRecordType, name, value)
}

func (cp *DefaultAWSCloudProvider) provisionRoute53Record(latticeID v1.LatticeID, recordType, name, value string) error {
	module := cp.route53RecordTerraformModule(latticeID, cp.route53PrivateZoneID, recordType, name, value)
	config := cp.route53RecordTerraformConfig(latticeID, cp.route53PrivateZoneID, name, module)
	_, err := terraform.Apply(route53RecordWorkDirectory(cp.route53PrivateZoneID, name), config)
	return err
}

func (cp *DefaultAWSCloudProvider) DestroyDNSRecord(latticeID v1.LatticeID, name string) error {
	config := cp.route53RecordTerraformConfig(latticeID, cp.route53PrivateZoneID, name, nil)
	_, err := terraform.Destroy(route53RecordWorkDirectory(cp.route53PrivateZoneID, name), config)
	return err
}

func (cp *DefaultAWSCloudProvider) route53RecordTerraformConfig(
	latticeID v1.LatticeID,
	zoneID, name string,
	module *kubetf.Route53Record,
) *terraform.Config {
	config := &terraform.Config{
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
	}

	if module != nil {
		config.Modules = map[string]interface{}{
			"route53-record": module,
		}
	}
	return config
}

func (cp *DefaultAWSCloudProvider) route53RecordTerraformModule(latticeID v1.LatticeID, zoneID, recordType, name, value string) *kubetf.Route53Record {
	return &kubetf.Route53Record{
		Source: cp.terraformModulePath + kubetf.ModulePathRoute53Record,

		Region: cp.region,

		ZoneID: zoneID,
		Type:   recordType,
		Name:   name,
		Value:  value,
	}
}

func route53RecordWorkDirectory(zoneID, name string) string {
	return workDirectory("route53", fmt.Sprintf("%v/%v", zoneID, name))
}
