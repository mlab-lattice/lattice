package nodepool

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	awscloudprovider "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubetf "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/terraform/aws"
	tf "github.com/mlab-lattice/lattice/pkg/util/terraform"
	awstfprovider "github.com/mlab-lattice/lattice/pkg/util/terraform/provider/aws"

	"github.com/golang/glog"
)

const (
	finalizerName = "node-pool.aws.cloud-provider.lattice.mlab.com"

	terraformOutputAutoscalingGroupID              = "autoscaling_group_id"
	terraformOutputAutoscalingGroupName            = "autoscaling_group_name"
	terraformOutputAutoscalingGroupDesiredCapacity = "autoscaling_group_desired_capacity"
	terraformOutputSecurityGroupID                 = "security_group_id"
)

func (c *Controller) syncDeletedNodePool(nodePool *latticev1.NodePool) error {
	err := c.deprovisionNodePool(nodePool)
	if err != nil {
		return err
	}

	_, err = c.removeFinalizer(nodePool)
	return err
}

func (c *Controller) syncNodePoolState(nodePool *latticev1.NodePool) (*latticev1.NodePool, error) {
	info, err := c.currentNodePoolInfo(nodePool)

	if err != nil || info.NumInstances < nodePool.Spec.NumInstances {
		// For now, assume an error retrieving output values means that the node pool hasn't been provisioned
		// TODO: look into adding different exit codes to `terraform output`
		return c.updateNodePoolStatus(nodePool, latticev1.NodePoolStateScalingUp)
	}

	if info.NumInstances > nodePool.Spec.NumInstances {
		return c.updateNodePoolStatus(nodePool, latticev1.NodePoolStateScalingDown)
	}

	return c.updateNodePoolStatus(nodePool, latticev1.NodePoolStateStable)
}

func (c *Controller) provisionNodePool(nodePool *latticev1.NodePool) (*latticev1.NodePool, error) {
	config := c.nodePoolConfig(nodePool)
	_, err := tf.Apply(workDirectory(nodePool.IDLabelValue()), config)
	if err != nil {
		return nil, err
	}

	annotations, err := c.nodePoolAnnotations(nodePool)
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(nodePool.Annotations, annotations) {
		// Copy so the shared cache isn't mutated
		nodePool = nodePool.DeepCopy()
		nodePool.Annotations = annotations

		nodePool, err = c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
		if err != nil {
			return nil, err
		}
	}

	return c.syncNodePoolState(nodePool)
}

func (c *Controller) deprovisionNodePool(nodePool *latticev1.NodePool) error {
	config := c.nodePoolConfig(nodePool)
	_, err := tf.Destroy(workDirectory(nodePool.IDLabelValue()), config)
	return err
}

func (c *Controller) nodePoolConfig(nodePool *latticev1.NodePool) *tf.Config {
	nodePoolID := nodePool.IDLabelValue()

	nodePoolModule := kubetf.NewNodePoolModule(
		c.terraformModuleRoot,
		c.awsCloudProvider.AccountID(),
		c.awsCloudProvider.Region(),
		string(c.latticeID),
		c.awsCloudProvider.VPCID(),
		strings.Join(c.awsCloudProvider.SubnetIDs(), ","),
		c.awsCloudProvider.MasterNodeSecurityGroupID(),
		c.awsCloudProvider.BaseNodeAMIID(),
		c.awsCloudProvider.KeyName(),
		nodePoolID,
		nodePool.Spec.NumInstances,
		nodePool.Spec.InstanceType,
	)

	return &tf.Config{
		Provider: awstfprovider.Provider{
			Region: c.awsCloudProvider.Region(),
		},
		Backend: tf.S3BackendConfig{
			Region: c.awsCloudProvider.Region(),
			Bucket: c.terraformBackendOptions.S3.Bucket,
			Key: fmt.Sprintf(
				"%v/%v",
				kubetf.GetS3BackendNodePoolPathRoot(c.latticeID, nodePoolID),
				nodePoolID,
			),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"node-pool": nodePoolModule,
		},
		Output: map[string]tf.ConfigOutput{
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

func (c *Controller) currentNodePoolInfo(nodePool *latticev1.NodePool) (nodePoolInfo, error) {
	outputVars := []string{
		terraformOutputAutoscalingGroupID,
		terraformOutputAutoscalingGroupName,
		terraformOutputAutoscalingGroupDesiredCapacity,
		terraformOutputSecurityGroupID,
	}

	config := c.nodePoolConfig(nodePool)
	values, err := tf.Output(workDirectory(nodePool.IDLabelValue()), config, outputVars)
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

func (c *Controller) updateNodePoolStatus(
	nodePool *latticev1.NodePool,
	state latticev1.NodePoolState,
) (*latticev1.NodePool, error) {
	status := latticev1.NodePoolStatus{
		State: state,
	}

	if reflect.DeepEqual(nodePool.Status, status) {
		return nodePool, nil
	}

	// Copy the service so the shared cache isn't mutated
	nodePool = nodePool.DeepCopy()
	nodePool.Status = status

	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).UpdateStatus(nodePool)
}

func (c *Controller) nodePoolAnnotations(nodePool *latticev1.NodePool) (map[string]string, error) {
	info, err := c.currentNodePoolInfo(nodePool)
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{
		awscloudprovider.AnnotationKeyNodePoolAutoscalingGroupName: info.AutoScalingGroupName,
		awscloudprovider.AnnotationKeyNodePoolSecurityGroupID:      info.SecurityGroupID,
	}
	return annotations, nil
}

func (c *Controller) addFinalizer(nodePool *latticev1.NodePool) (*latticev1.NodePool, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range nodePool.Finalizers {
		if finalizer == finalizerName {
			glog.V(5).Infof("NodePool %v has %v finalizer", nodePool.Name, finalizerName)
			return nodePool, nil
		}
	}

	// Add the finalizer to the list and update.
	// If this fails due to a race the Endpoint should get requeued by the controller, so
	// not a big deal.
	nodePool.Finalizers = append(nodePool.Finalizers, finalizerName)
	glog.V(5).Infof("NodePool %v missing %v finalizer, adding it", nodePool.Name, finalizerName)

	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
}

func (c *Controller) removeFinalizer(nodePool *latticev1.NodePool) (*latticev1.NodePool, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range nodePool.Finalizers {
		if finalizer == finalizerName {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return nodePool, nil
	}

	// The finalizer was in the list, so we should remove it.
	nodePool.Finalizers = finalizers
	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
}

func workDirectory(nodePoolID string) string {
	return "/tmp/lattice-controller-manager/controllers/cloud/aws/node-pool/terraform/" + nodePoolID
}
