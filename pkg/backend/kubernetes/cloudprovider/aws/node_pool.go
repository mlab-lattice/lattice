package aws

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubetf "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws/terraform"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"
	awstfprovider "github.com/mlab-lattice/lattice/pkg/util/terraform/provider/aws"
)

func (cp *DefaultAWSCloudProvider) ProvisionNodePool(latticeID v1.LatticeID, nodePool *latticev1.NodePool) (*latticev1.NodePool, *time.Duration, error) {
	config := cp.nodePoolTerraformConfig(latticeID, nodePool)
	_, err := terraform.Apply(nodePoolWorkDirectory(nodePool.IDLabelValue()), config)
	if err != nil {
		return nil, nil, err
	}

	annotations, err := cp.nodePoolAnnotations(latticeID, nodePool)
	if err != nil {
		return nil, nil, err
	}

	if !reflect.DeepEqual(nodePool.Annotations, annotations) {
		// Copy so the shared cache isn't mutated
		nodePool = nodePool.DeepCopy()
		nodePool.Annotations = annotations
	}

	return nodePool, nil, nil
}

func (cp *DefaultAWSCloudProvider) DeprovisionNodePool(latticeID v1.LatticeID, nodePool *latticev1.NodePool) (*time.Duration, error) {
	config := cp.nodePoolTerraformConfig(latticeID, nodePool)
	_, err := terraform.Destroy(nodePoolWorkDirectory(nodePool.IDLabelValue()), config)
	return nil, err
}

func (cp *DefaultAWSCloudProvider) NodePoolState(latticeID v1.LatticeID, nodePool *latticev1.NodePool) (latticev1.NodePoolState, error) {
	info, err := cp.currentNodePoolInfo(latticeID, nodePool)

	if err != nil {
		// For now, assume an error retrieving output values means that the node pool hasn't been provisioned
		// TODO: look into adding different exit codes to `terraform output`
		return latticev1.NodePoolStatePending, nil
	}

	if info.NumInstances != nodePool.Spec.NumInstances {
		return latticev1.NodePoolStateScaling, nil
	}

	return latticev1.NodePoolStateStable, nil
}

func (cp *DefaultAWSCloudProvider) nodePoolTerraformConfig(latticeID v1.LatticeID, nodePool *latticev1.NodePool) *terraform.Config {
	nodePoolID := nodePool.IDLabelValue()

	nodePoolModule := kubetf.NewNodePoolModule(
		cp.terraformModulePath,
		cp.accountID,
		cp.region,
		latticeID,
		cp.vpcID,
		cp.subnetIDs,
		cp.masterNodeSecurityGroupID,
		cp.workerNodeAMIID,
		cp.keyName,
		nodePoolID,
		nodePool.Spec.NumInstances,
		nodePool.Spec.InstanceType,
	)

	return &terraform.Config{
		Provider: awstfprovider.Provider{
			Region: cp.region,
		},
		Backend: terraform.S3BackendConfig{
			Region: cp.region,
			Bucket: cp.terraformBackendOptions.S3.Bucket,
			Key: fmt.Sprintf(
				"%v/%v",
				kubetf.GetS3BackendNodePoolPathRoot(latticeID, nodePoolID),
				nodePoolID,
			),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"node-pool": nodePoolModule,
		},
		Output: map[string]terraform.ConfigOutput{
			terraformOutputAutoscalingGroupID: {
				Value: fmt.Sprintf("${module.node-pool.%v}", terraformOutputAutoscalingGroupID),
			},
			terraformOutputAutoscalingGroupName: {
				Value: fmt.Sprintf("${module.node-pool.%v}", terraformOutputAutoscalingGroupName),
			},
			terraformOutputAutoscalingGroupDesiredCapacity: {
				Value: fmt.Sprintf("${module.node-pool.%v}", terraformOutputAutoscalingGroupDesiredCapacity),
			},
			terraformOutputSecurityGroupID: {
				Value: fmt.Sprintf("${module.node-pool.%v}", terraformOutputSecurityGroupID),
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

func (cp *DefaultAWSCloudProvider) currentNodePoolInfo(latticeID v1.LatticeID, nodePool *latticev1.NodePool) (nodePoolInfo, error) {
	outputVars := []string{
		terraformOutputAutoscalingGroupID,
		terraformOutputAutoscalingGroupName,
		terraformOutputAutoscalingGroupDesiredCapacity,
		terraformOutputSecurityGroupID,
	}

	config := cp.nodePoolTerraformConfig(latticeID, nodePool)
	values, err := terraform.Output(nodePoolWorkDirectory(nodePool.IDLabelValue()), config, outputVars)
	if err != nil {
		return nodePoolInfo{}, err
	}

	numInstances, err := strconv.ParseInt(values[terraformOutputAutoscalingGroupDesiredCapacity], 10, 32)
	if err != nil {
		return nodePoolInfo{}, err
	}

	info := nodePoolInfo{
		AutoScalingGroupID:   values[terraformOutputAutoscalingGroupID],
		AutoScalingGroupName: values[terraformOutputAutoscalingGroupName],
		NumInstances:         int32(numInstances),
		SecurityGroupID:      values[terraformOutputSecurityGroupID],
	}
	return info, nil
}

func (cp *DefaultAWSCloudProvider) nodePoolAnnotations(latticeID v1.LatticeID, nodePool *latticev1.NodePool) (map[string]string, error) {
	info, err := cp.currentNodePoolInfo(latticeID, nodePool)
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{
		AnnotationKeyNodePoolAutoscalingGroupName: info.AutoScalingGroupName,
		AnnotationKeyNodePoolSecurityGroupID:      info.SecurityGroupID,
	}
	return annotations, nil
}

func nodePoolWorkDirectory(nodePoolID string) string {
	return workDirectory("node-pool", nodePoolID)
}
