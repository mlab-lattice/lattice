package provisioner

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	awsterraform "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	"github.com/mlab-lattice/system/pkg/terraform"
)

type AWSProvisioner struct {
	latticeContainerRegistry   string
	latticeContainerRepoPrefix string
	tec                        *terraform.ExecContext
	config                     AWSProvisionerConfig
}

type AWSProvisionerConfig struct {
	TerraformModulePath string

	AccountID         string
	Region            string
	AvailabilityZones []string
	KeyName           string

	MasterNodeInstanceType string
	MasterNodeAMIID        string
	BaseNodeAMIID          string
}

const (
	clusterModulePath = "aws/cluster"
	// FIXME: move to constants
	clusterManagerAPIPort               = 80
	clusterManagerAddressOutputVariable = "cluster_manager_address"
)

func NewAWSProvisioner(latticeImageDockerRepository, latticeContainerRepoPrefix, workingDir string, config AWSProvisionerConfig) (*AWSProvisioner, error) {
	tec, err := terraform.NewTerrafromExecContext(workingDir, nil)
	if err != nil {
		return nil, err
	}

	ap := &AWSProvisioner{
		latticeContainerRegistry:   latticeImageDockerRepository,
		latticeContainerRepoPrefix: latticeContainerRepoPrefix,
		tec:    tec,
		config: config,
	}
	return ap, nil
}

func (ap *AWSProvisioner) Provision(clusterID, url string) error {
	// Add system json to working dir
	sysJSON, err := ap.getClusterTerraformJSON(clusterID, url)
	if err != nil {
		return err
	}

	err = ap.tec.AddFile("config.tf.json", sysJSON)
	if err != nil {
		return err
	}

	// Run `terraform init`
	result, logFilename, err := ap.tec.Init()
	if err != nil {
		return err
	}

	fmt.Printf("Running terraform init (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*ap.tec.LogPath, logFilename))

	err = result.Wait()
	if err != nil {
		return err
	}

	// Run `terraform apply`
	result, logFilename, err = ap.tec.Apply(nil)
	if err != nil {
		return err
	}

	fmt.Printf("Running terraform apply (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*ap.tec.LogPath, logFilename))

	err = result.Wait()
	if err != nil {
		return err
	}

	address, err := ap.Address(clusterID)
	if err != nil {
		return err
	}

	fmt.Println("Waiting for Cluster Environment Manager to be ready...")
	return pollForClusterReadiness(address)
}

func (ap *AWSProvisioner) getClusterTerraformJSON(clusterID, url string) ([]byte, error) {
	sysModule := awsterraform.Cluster{
		Source: filepath.Join(ap.config.TerraformModulePath, clusterModulePath),

		AWSAccountID:      ap.config.AccountID,
		Region:            ap.config.Region,
		AvailabilityZones: ap.config.AvailabilityZones,
		KeyName:           ap.config.KeyName,

		ClusterID:           clusterID,
		SystemDefinitionURL: url,

		MasterNodeInstanceType: ap.config.MasterNodeInstanceType,
		MasterNodeAMIID:        ap.config.MasterNodeAMIID,
		ClusterManagerAPIPort:  clusterManagerAPIPort,

		BaseNodeAMIID: ap.config.BaseNodeAMIID,
	}

	jsonMap := map[string]interface{}{
		"provider": map[string]interface{}{
			"aws": map[string]interface{}{
				"region": sysModule.Region,
			},
		},
		"module": map[string]interface{}{
			"cluster": sysModule,
		},
		"output": map[string]interface{}{
			clusterManagerAddressOutputVariable: map[string]interface{}{
				"value": fmt.Sprintf("${module.cluster.%v}", clusterManagerAddressOutputVariable),
			},
		},
	}

	return json.MarshalIndent(jsonMap, "", "  ")
}

func (ap *AWSProvisioner) Address(name string) (string, error) {
	return ap.tec.Output(clusterManagerAddressOutputVariable)
}

func (ap *AWSProvisioner) Deprovision(name string) error {
	address, err := ap.Address(name)
	if err != nil {
		return err
	}

	err = tearDownAndWaitForSuccess(address)
	if err != nil {
		return err
	}

	// Run `terraform destroy`
	result, logFilename, err := ap.tec.Destroy(nil)
	if err != nil {
		return err
	}

	fmt.Printf("Running terraform destroy (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*ap.tec.LogPath, logFilename))
	return result.Wait()
}
