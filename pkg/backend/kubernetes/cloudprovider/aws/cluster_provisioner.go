package aws

import (
	"fmt"
	"path/filepath"
	"time"

	awsterraform "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
	"github.com/mlab-lattice/system/pkg/terraform"
	awstfprovider "github.com/mlab-lattice/system/pkg/terraform/provider/aws"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	clusterModulePath = "aws/cluster"
	// FIXME: move to constants
	clusterManagerAPIPort                = 80
	terraformOutputclusterManagerAddress = "cluster_manager_address"
)

type ClusterProvisionerOptions struct {
	TerraformModulePath      string
	TerraformBackendS3Bucket string
	TerraformBackendS3Key    string

	ClusterManagerURL string

	AccountID         string
	Region            string
	AvailabilityZones []string
	KeyName           string

	MasterNodeInstanceType string
	MasterNodeAMIID        string
	BaseNodeAMIID          string
}

type DefaultAWSClusterProvisioner struct {
	workDirectory string

	latticeContainerRegistry   string
	latticeContainerRepoPrefix string

	terraformModulePath      string
	terraformBackendS3Bucket string
	terraformBackendS3Key    string

	clusterManagerURL string

	accountID         string
	region            string
	availabilityZones []string
	keyName           string

	masterNodeInstanceType string
	masterNodeAMIID        string
	baseNodeAMIID          string
}

func NewClusterProvisioner(latticeImageDockerRepository, latticeContainerRepoPrefix, workingDir string, options *ClusterProvisionerOptions) *DefaultAWSClusterProvisioner {
	return &DefaultAWSClusterProvisioner{
		workDirectory: workingDir,

		latticeContainerRegistry:   latticeImageDockerRepository,
		latticeContainerRepoPrefix: latticeContainerRepoPrefix,

		terraformModulePath:      options.TerraformModulePath,
		terraformBackendS3Bucket: options.TerraformBackendS3Bucket,
		terraformBackendS3Key:    options.TerraformBackendS3Key,

		accountID:         options.AccountID,
		region:            options.Region,
		availabilityZones: options.AvailabilityZones,
		keyName:           options.KeyName,

		masterNodeInstanceType: options.MasterNodeInstanceType,
		masterNodeAMIID:        options.MasterNodeAMIID,
		baseNodeAMIID:          options.BaseNodeAMIID,
	}
}

func (p *DefaultAWSClusterProvisioner) Provision(clusterID, url string) (string, error) {
	fmt.Println("Provisioning cluster...")
	clusterModule := p.clusterModule(clusterID, url)
	clusterConfig := p.clusterConfig(clusterModule)

	logfile, err := terraform.Apply(p.workDirectory, clusterConfig)
	if err != nil {
		if logfile != "" {
			fmt.Printf("error provisioning. logfile: %v", logfile)
		}
		return "", err
	}

	address, err := p.address(clusterID)
	if err != nil {
		return "", err
	}

	fmt.Println("Waiting for Cluster Manager to be ready...")
	clusterClient := rest.NewClient(fmt.Sprintf("http://%v", address))
	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		ok, _ := clusterClient.Status()
		return ok, nil
	})

	if err != nil {
		return "", err
	}

	return address, nil
}

func (p *DefaultAWSClusterProvisioner) clusterConfig(clusterModule *awsterraform.Cluster) *terraform.Config {
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

	if clusterModule != nil {
		config.Modules = map[string]interface{}{
			"cluster": clusterModule,
		}

		config.Output = map[string]terraform.ConfigOutput{
			terraformOutputclusterManagerAddress: {
				Value: fmt.Sprintf("${module.cluster.%v}", terraformOutputclusterManagerAddress),
			},
		}
	}

	return config
}

func (p *DefaultAWSClusterProvisioner) clusterModule(clusterID, url string) *awsterraform.Cluster {
	return &awsterraform.Cluster{
		Source: filepath.Join(p.terraformModulePath, clusterModulePath),

		AWSAccountID:      p.accountID,
		Region:            p.region,
		AvailabilityZones: p.availabilityZones,
		KeyName:           p.keyName,

		ClusterID:                    clusterID,
		ControlPlaneContainerChannel: fmt.Sprintf("%v/%v", p.latticeContainerRegistry, p.latticeContainerRepoPrefix),
		SystemDefinitionURL:          url,

		MasterNodeInstanceType: p.masterNodeInstanceType,
		MasterNodeAMIID:        p.masterNodeAMIID,
		ClusterManagerAPIPort:  clusterManagerAPIPort,

		BaseNodeAMIID: p.baseNodeAMIID,
	}
}

func (p *DefaultAWSClusterProvisioner) address(name string) (string, error) {
	tec, err := terraform.NewTerrafromExecContext(p.workDirectory, nil)
	if err != nil {
		return "", err
	}

	return tec.Output(terraformOutputclusterManagerAddress)
}

func (p *DefaultAWSClusterProvisioner) Deprovision(clusterID string, force bool) error {
	fmt.Println("Deprovisioning cluster...")

	if !force {
		if err := p.tearDownSystems(clusterID); err != nil {
			return err
		}
	}

	logfile, err := terraform.Destroy(p.workDirectory, nil)
	if err != nil && logfile != "" {
		fmt.Printf("error destroying. logfile: %v", logfile)
	}
	return err
}

func (p *DefaultAWSClusterProvisioner) tearDownSystems(clusterID string) error {
	if p.clusterManagerURL == "" {
		return fmt.Errorf("cluster manager URL required to tear down systems")
	}

	clusterClient := rest.NewClient(p.clusterManagerURL)
	systems, err := clusterClient.Systems().List()
	if err != nil {
		return err
	}

	teardowns := map[types.SystemID]types.SystemTeardownID{}
	for _, system := range systems {
		teardownID, err := clusterClient.Systems().Teardowns(system.ID).Create()
		if err != nil {
			return err
		}

		teardowns[system.ID] = teardownID
	}

	err = wait.Poll(10*time.Second, 600*time.Second, func() (bool, error) {
		for systemID, teardownID := range teardowns {
			teardown, err := clusterClient.Systems().Teardowns(systemID).Get(teardownID)
			if err != nil {
				return false, err
			}

			if teardown.State == types.SystemTeardownStateFailed {
				return false, fmt.Errorf("teardown %v (system %v) failed", teardownID, systemID)
			}

			if teardown.State != types.SystemTeardownStateSucceeded {
				return false, nil
			}
		}

		return true, nil
	})

	return err
}
