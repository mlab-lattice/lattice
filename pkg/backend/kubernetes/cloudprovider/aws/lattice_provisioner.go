package aws

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/mlab-lattice/system/pkg/api/client/rest"
	"github.com/mlab-lattice/system/pkg/api/v1"
	awsterraform "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	"github.com/mlab-lattice/system/pkg/util/terraform"
	awstfprovider "github.com/mlab-lattice/system/pkg/util/terraform/provider/aws"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	latticeModulePath = "aws/lattice"
	// FIXME: move to constants
	apiServerPort                   = 80
	terraformOutputAPIServerAddress = "api_server_address"
)

type LatticeProvisionerOptions struct {
	TerraformModulePath      string
	TerraformBackendS3Bucket string
	TerraformBackendS3Key    string

	LatticeAPIServerURL string

	AccountID         string
	Region            string
	AvailabilityZones []string
	KeyName           string

	MasterNodeInstanceType string
	MasterNodeAMIID        string
	BaseNodeAMIID          string
}

type DefaultAWSLatticeProvisioner struct {
	workDirectory string

	latticeContainerRegistry   string
	latticeContainerRepoPrefix string

	terraformModulePath      string
	terraformBackendS3Bucket string
	terraformBackendS3Key    string

	apiServerURL string

	accountID         string
	region            string
	availabilityZones []string
	keyName           string

	masterNodeInstanceType string
	masterNodeAMIID        string
	baseNodeAMIID          string
}

func NewLatticeProvisioner(latticeImageDockerRepository, latticeContainerRepoPrefix, workingDir string, options *LatticeProvisionerOptions) *DefaultAWSLatticeProvisioner {
	return &DefaultAWSLatticeProvisioner{
		workDirectory: workingDir,

		latticeContainerRegistry:   latticeImageDockerRepository,
		latticeContainerRepoPrefix: latticeContainerRepoPrefix,

		terraformModulePath:      options.TerraformModulePath,
		terraformBackendS3Bucket: options.TerraformBackendS3Bucket,
		terraformBackendS3Key:    options.TerraformBackendS3Key,

		apiServerURL: options.LatticeAPIServerURL,

		accountID:         options.AccountID,
		region:            options.Region,
		availabilityZones: options.AvailabilityZones,
		keyName:           options.KeyName,

		masterNodeInstanceType: options.MasterNodeInstanceType,
		masterNodeAMIID:        options.MasterNodeAMIID,
		baseNodeAMIID:          options.BaseNodeAMIID,
	}
}

func (p *DefaultAWSLatticeProvisioner) Provision(latticeID string, initialSystemDefinitionURL *string) (string, error) {
	fmt.Println("Provisioning lattice...")
	latticeModule := p.latticeModule(latticeID, initialSystemDefinitionURL)
	latticeConfig := p.latticeConfig(latticeModule)

	logfile, err := terraform.Apply(p.workDirectory, latticeConfig)
	if err != nil {
		if logfile != "" {
			fmt.Printf("error provisioning. logfile: %v", logfile)
		}
		return "", err
	}

	address, err := p.address(latticeID)
	if err != nil {
		return "", err
	}

	fmt.Println("Waiting for API server to be ready...")
	latticeClient := rest.NewClient(address)
	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		ok, _ := latticeClient.Status()
		return ok, nil
	})

	if err != nil {
		return "", err
	}

	return address, nil
}

func (p *DefaultAWSLatticeProvisioner) latticeConfig(latticeModule *awsterraform.Lattice) *terraform.Config {
	config := &terraform.Config{
		Provider: awstfprovider.Provider{
			Region: p.region,
		},
		Backend: terraform.S3BackendConfig{
			Region:  p.region,
			Bucket:  p.terraformBackendS3Bucket,
			Key:     p.terraformBackendS3Key,
			Encrypt: true,
		},
	}

	if latticeModule != nil {
		config.Modules = map[string]interface{}{
			"lattice": latticeModule,
		}

		config.Output = map[string]terraform.ConfigOutput{
			terraformOutputAPIServerAddress: {
				Value: fmt.Sprintf("${module.lattice.%v}", terraformOutputAPIServerAddress),
			},
		}
	}

	return config
}

func (p *DefaultAWSLatticeProvisioner) latticeModule(latticeID string, initialSystemDefinitionURL *string) *awsterraform.Lattice {
	url := ""
	if initialSystemDefinitionURL != nil {
		url = *initialSystemDefinitionURL
	}

	return &awsterraform.Lattice{
		Source: filepath.Join(p.terraformModulePath, latticeModulePath),

		AWSAccountID:      p.accountID,
		Region:            p.region,
		AvailabilityZones: p.availabilityZones,
		KeyName:           p.keyName,

		LatticeID:                    latticeID,
		ControlPlaneContainerChannel: fmt.Sprintf("%v/%v", p.latticeContainerRegistry, p.latticeContainerRepoPrefix),
		SystemDefinitionURL:          url,

		MasterNodeInstanceType: p.masterNodeInstanceType,
		MasterNodeAMIID:        p.masterNodeAMIID,
		APIServerPort:          apiServerPort,

		BaseNodeAMIID: p.baseNodeAMIID,
	}
}

func (p *DefaultAWSLatticeProvisioner) address(name string) (string, error) {
	tec, err := terraform.NewTerrafromExecContext(p.workDirectory, nil)
	if err != nil {
		return "", err
	}

	address, err := tec.Output(terraformOutputAPIServerAddress)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%v", address), nil
}

func (p *DefaultAWSLatticeProvisioner) Deprovision(latticeID string, force bool) error {
	fmt.Println("Deprovisioning lattice...")

	if !force {
		if err := p.tearDownSystems(latticeID); err != nil {
			return err
		}
	}

	logfile, err := terraform.Destroy(p.workDirectory, nil)
	if err != nil && logfile != "" {
		fmt.Printf("error destroying. logfile: %v", logfile)
	}
	return err
}

func (p *DefaultAWSLatticeProvisioner) tearDownSystems(latticeID string) error {
	if p.apiServerURL == "" {
		return fmt.Errorf("API server URL required to tear down systems")
	}

	latticeClient := rest.NewClient(p.apiServerURL)
	systems, err := latticeClient.Systems().List()
	if err != nil {
		return err
	}

	teardowns := map[v1.SystemID]v1.TeardownID{}
	for _, system := range systems {
		teardown, err := latticeClient.Systems().Teardowns(system.ID).Create()
		if err != nil {
			return err
		}

		teardowns[system.ID] = teardown.ID
	}

	err = wait.Poll(10*time.Second, 600*time.Second, func() (bool, error) {
		for systemID, teardownID := range teardowns {
			teardown, err := latticeClient.Systems().Teardowns(systemID).Get(teardownID)
			if err != nil {
				return false, err
			}

			if teardown.State == v1.TeardownStateFailed {
				return false, fmt.Errorf("teardown %v (system %v) failed", teardownID, systemID)
			}

			if teardown.State != v1.TeardownStateSucceeded {
				return false, nil
			}
		}

		return true, nil
	})

	return err
}
