package aws

import (
	"fmt"
	"strconv"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubetf "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws/terraform"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"
	awstfprovider "github.com/mlab-lattice/lattice/pkg/util/terraform/provider/aws"
)

const (
	AnnotationKeyNodePoolAutoscalingGroupName = "node-pool.aws.cloud-provider.lattice.mlab.com/autoscaling-group-name"
	AnnotationKeyNodePoolSecurityGroupID      = "node-pool.aws.cloud-provider.lattice.mlab.com/security-group-id"

	terraformOutputNodePoolAutoscalingGroupID              = "autoscaling_group_id"
	terraformOutputNodePoolAutoscalingGroupName            = "autoscaling_group_name"
	terraformOutputNodePoolAutoscalingGroupDesiredCapacity = "autoscaling_group_desired_capacity"
	terraformOutputNodePoolSecurityGroupID                 = "security_group_id"
)

func (cp *DefaultAWSCloudProvider) NodePoolNeedsNewEpoch(nodePool *latticev1.NodePool) (bool, error) {
	current, ok := nodePool.Status.Epochs.CurrentEpoch()

	// If the node pool doesn't have an epoch yet, it needs a new one
	if !ok {
		return true, nil
	}

	epoch, ok := nodePool.Status.Epochs.Epoch(current)
	if !ok {
		return false, fmt.Errorf("could not get epoch status for current epoch %v", current)
	}

	// If the node pool's instance type is not the instance type of the current epoch, we
	// need a new epoch.
	return nodePool.Spec.InstanceType != epoch.InstanceType, nil
}

func (cp *DefaultAWSCloudProvider) NodePoolCurrentEpochState(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
) (latticev1.NodePoolState, error) {
	current, ok := nodePool.Status.Epochs.CurrentEpoch()
	if !ok {
		return latticev1.NodePoolStatePending, fmt.Errorf("could not get current epoch for %v", nodePool.Description(cp.namespacePrefix))
	}

	epochInfo, ok := nodePool.Status.Epochs.Epoch(current)
	if !ok {
		return latticev1.NodePoolStatePending, fmt.Errorf("could not get epoch status for %v epoch %v", nodePool.Description(cp.namespacePrefix), current)
	}

	// Invariant: NodePoolCurrentEpochState will only be called if NodePoolNeedsNewEpoch returns false.
	// Therefore, we don't have to check the instance type since NodePoolNeedsNewEpoch would have returned
	// true if the they mismatched.

	if epochInfo.State == latticev1.NodePoolStatePending {
		return latticev1.NodePoolStatePending, nil
	}

	if nodePool.Spec.NumInstances != epochInfo.NumInstances {
		return latticev1.NodePoolStateScaling, nil
	}

	return latticev1.NodePoolStateStable, nil
}

func (cp *DefaultAWSCloudProvider) NodePoolAddAnnotations(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	annotations map[string]string,
	epoch latticev1.NodePoolEpoch,
) error {
	info, err := cp.nodePoolEpochInfo(latticeID, nodePool, epoch)
	if err != nil {
		return err
	}

	annotations[AnnotationKeyNodePoolAutoscalingGroupName] = info.AutoScalingGroupName
	annotations[AnnotationKeyNodePoolSecurityGroupID] = info.SecurityGroupID
	return nil
}

func (cp *DefaultAWSCloudProvider) ProvisionNodePoolEpoch(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) error {
	config := cp.nodePoolTerraformConfig(latticeID, nodePool, epoch)
	_, err := terraform.Apply(nodePoolWorkDirectory(nodePool.ID(epoch)), config)
	return err
}

func (cp *DefaultAWSCloudProvider) DeprovisionNodePoolEpoch(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) error {
	_, err := terraform.Destroy(nodePoolWorkDirectory(nodePool.ID(epoch)), nil)
	return err
}

func (cp *DefaultAWSCloudProvider) nodePoolEpochInfo(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) (nodePoolInfo, error) {
	outputVars := []string{
		terraformOutputNodePoolAutoscalingGroupID,
		terraformOutputNodePoolAutoscalingGroupName,
		terraformOutputNodePoolAutoscalingGroupDesiredCapacity,
		terraformOutputNodePoolSecurityGroupID,
	}

	config := cp.nodePoolTerraformConfig(latticeID, nodePool, epoch)
	values, err := terraform.Output(nodePoolWorkDirectory(nodePool.ID(epoch)), config, outputVars)
	if err != nil {
		return nodePoolInfo{}, err
	}

	numInstances, err := strconv.ParseInt(values[terraformOutputNodePoolAutoscalingGroupDesiredCapacity], 10, 32)
	if err != nil {
		return nodePoolInfo{}, err
	}

	info := nodePoolInfo{
		AutoScalingGroupID:   values[terraformOutputNodePoolAutoscalingGroupID],
		AutoScalingGroupName: values[terraformOutputNodePoolAutoscalingGroupName],
		NumInstances:         int32(numInstances),
		SecurityGroupID:      values[terraformOutputNodePoolSecurityGroupID],
	}
	return info, nil
}

func (cp *DefaultAWSCloudProvider) nodePoolTerraformConfig(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) *terraform.Config {
	nodePoolID := nodePool.ID(epoch)

	return &terraform.Config{
		Provider: awstfprovider.Provider{
			Region: cp.region,
		},
		Backend: terraform.S3BackendConfig{
			Region: cp.region,
			Bucket: cp.terraformBackendOptions.S3.Bucket,
			Key: fmt.Sprintf(
				"%v/%v",
				kubetf.GetS3BackendNodePoolPathRoot(latticeID, nodePool.Namespace, nodePoolID),
				nodePoolID,
			),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"node-pool": kubetf.NodePool{
				Source: cp.terraformModulePath + kubetf.ModulePathNodePool,

				AWSAccountID: cp.accountID,
				Region:       cp.region,

				LatticeID:                 latticeID,
				VPCID:                     cp.vpcID,
				SubnetIDs:                 cp.subnetIDs,
				MasterNodeSecurityGroupID: cp.masterNodeSecurityGroupID,
				WorkerNodeAMIID:           cp.workerNodeAMIID,
				KeyName:                   cp.keyName,

				Name:         nodePoolID,
				NumInstances: nodePool.Spec.NumInstances,
				InstanceType: nodePool.Spec.InstanceType,
			},
		},
		Output: map[string]terraform.ConfigOutput{
			terraformOutputNodePoolAutoscalingGroupID: {
				Value: fmt.Sprintf("${module.node-pool.%v}", terraformOutputNodePoolAutoscalingGroupID),
			},
			terraformOutputNodePoolAutoscalingGroupName: {
				Value: fmt.Sprintf("${module.node-pool.%v}", terraformOutputNodePoolAutoscalingGroupName),
			},
			terraformOutputNodePoolAutoscalingGroupDesiredCapacity: {
				Value: fmt.Sprintf("${module.node-pool.%v}", terraformOutputNodePoolAutoscalingGroupDesiredCapacity),
			},
			terraformOutputNodePoolSecurityGroupID: {
				Value: fmt.Sprintf("${module.node-pool.%v}", terraformOutputNodePoolSecurityGroupID),
			},
		},
	}
}

type nodePoolInfo struct {
	AutoScalingGroupID   string
	AutoScalingGroupName string
	NumInstances         int32
	SecurityGroupID      string
}

func nodePoolWorkDirectory(nodePoolID string) string {
	return workDirectory("node-pool", nodePoolID)
}
