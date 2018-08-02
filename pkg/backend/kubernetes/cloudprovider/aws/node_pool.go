package aws

import (
	"fmt"
	"strconv"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubetf "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws/terraform"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"
	awstfprovider "github.com/mlab-lattice/lattice/pkg/util/terraform/provider/aws"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	// note - will be GenerateBootstrapToken from "k8s.io/client-go/tools/bootstrap/token/util" at 1.10.4
	tokenutil "k8s.io/kubernetes/cmd/kubeadm/app/util/token"
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
	return nodePool.Spec.InstanceType != epoch.Spec.InstanceType, nil
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

func (cp *DefaultAWSCloudProvider) NodePoolEpochStatus(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
	epochSpec *latticev1.NodePoolSpec,
) (*latticev1.NodePoolStatusEpochStatus, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.NodePoolIDLabelKey, selection.Equals, []string{nodePool.ID(epoch)})
	if err != nil {
		return nil, fmt.Errorf("error making requirement for %v node lookup: %v", nodePool.Description(cp.namespacePrefix), err)
	}

	selector = selector.Add(*requirement)
	nodes, err := cp.kubeNodeLister.List(selector)
	if err != nil {
		return nil, fmt.Errorf("error getting nodes for %v: %v", nodePool.Description(cp.namespacePrefix), err)
	}

	var n []corev1.Node
	for _, node := range nodes {
		n = append(n, *node)
	}

	ready := kubernetes.NumReadyNodes(n)
	status := &latticev1.NodePoolStatusEpochStatus{
		NumInstances: ready,
		InstanceType: epochSpec.InstanceType,
		State:        latticev1.NodePoolStateScaling,
	}

	if ready == epochSpec.NumInstances {
		status.State = latticev1.NodePoolStateStable
	}

	return status, nil
}

func (cp *DefaultAWSCloudProvider) EnsureNodePoolEpoch(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) error {
	// TODO get lattice apiserver address + port and push into TF module here too (or downstream in the module fcn)
	bootstrapSecret, err := cp.CreateBootstrapToken(latticeID)
	if err != nil {
		return fmt.Errorf(
			"error making secret for %v epoch %v: %v",
			nodePool.Description(cp.namespacePrefix),
			epoch,
			err,
		)
	}

	module := cp.nodePoolTerraformModule(latticeID, nodePool, epoch, bootstrapSecret)
	config := cp.nodePoolTerraformConfig(latticeID, nodePool, epoch, module)

	result, _, err := terraform.Plan(nodePoolWorkDirectory(nodePool.ID(epoch)), config, false)
	if err != nil {
		return fmt.Errorf(
			"error getting terraform plan for %v epoch %v: %v",
			nodePool.Description(cp.namespacePrefix),
			epoch,
			err,
		)
	}

	switch result {
	case terraform.PlanResultNotEmpty:
		_, err = terraform.Apply(nodePoolWorkDirectory(nodePool.ID(epoch)), config)
		if err != nil {
			return fmt.Errorf(
				"error applying terraform for %v epoch %v: %v",
				nodePool.Description(cp.namespacePrefix),
				epoch,
				err,
			)
		}

		return nil

	case terraform.PlanResultEmpty:
		return nil

	default:
		return fmt.Errorf(
			"unknown error getting terraform plan for %v epoch %v",
			nodePool.Description(cp.namespacePrefix),
			epoch,
		)
	}
}

// TODO push this functionality elsewhere, it shouldn't be here
func (cp *DefaultAWSCloudProvider) CreateBootstrapToken(
	latticeId v1.LatticeID,
) (string, error) {
	bootstrapToken, err := tokenutil.GenerateToken()
	if err != nil {
		return "", fmt.Errorf(
			"error generating bootstrap token: %v",
			err,
		)
	}
	tokenId, tokenSecret, err := tokenutil.ParseToken(bootstrapToken)
	if err != nil {
		return "", fmt.Errorf(
			"error splitting bootstrap token into component parts: %v",
			err,
		)
	}

	/*
		apiVersion: v1
		kind: Secret
		metadata:
		  # Name MUST be of form "bootstrap-token-<token id>"
		  name: bootstrap-token-07401b
		  namespace: kube-system

		# Type MUST be 'bootstrap.kubernetes.io/token'
		type: bootstrap.kubernetes.io/token
		stringData:
		  # Human readable description. Optional.
		  description: "The default bootstrap token generated by 'kubeadm init'."

		  # Token ID and secret. Required.
		  token-id: 07401b
		  token-secret: f395accd246ae52d

		  # Expiration. Optional.
		  expiration: 2017-03-10T03:22:11Z

		  # Allowed usages.
		  usage-bootstrap-authentication: "true"
		  usage-bootstrap-signing: "true"

		  # Extra groups to authenticate the token as. Must start with "system:bootstrappers:"
		  auth-extra-groups: system:bootstrappers:worker,system:bootstrappers:ingress
	*/
	// TODO add expiration date to token (1h? 12h?)
	secretMap := map[string]string{
		"description":                    "Bootstrap token for node-pool on lattice " + string(latticeId),
		"token-id":                       tokenId,
		"token-secret":                   tokenSecret,
		"usage-bootstrap-authentication": "true",
		"usage-bootstrap-signing":        "true",
		"auth-extra-groups":              "system:bootstrappers:kubeadm:default-node-token",
	}
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bootstrap-token-" + tokenId,
			Namespace: "kube-system",
		},
		Type:       "bootstrap.kubernetes.io/token",
		StringData: secretMap,
	}
	_, errr := cp.kubeClient.CoreV1().Secrets("kube-system").Create(secret)
	return bootstrapToken, errr
}

func (cp *DefaultAWSCloudProvider) DestroyNodePoolEpoch(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) error {
	config := cp.nodePoolTerraformConfig(latticeID, nodePool, epoch, nil)
	_, err := terraform.Destroy(nodePoolWorkDirectory(nodePool.ID(epoch)), config)
	if err != nil {
		return fmt.Errorf(
			"error destroying terraform for %v epoch %v: %v",
			nodePool.Description(cp.namespacePrefix),
			epoch,
			err,
		)
	}

	return nil
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

	module := cp.nodePoolTerraformModule(latticeID, nodePool, epoch, "")
	config := cp.nodePoolTerraformConfig(latticeID, nodePool, epoch, module)
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
	module *kubetf.NodePool,
) *terraform.Config {
	nodePoolID := nodePool.ID(epoch)

	config := &terraform.Config{
		Provider: awstfprovider.Provider{
			Region: cp.region,
		},
		Backend: terraform.S3BackendConfig{
			Region:  cp.region,
			Bucket:  cp.terraformBackendOptions.S3.Bucket,
			Key:     kubetf.GetS3BackendNodePoolPathRoot(latticeID, nodePool.Namespace, nodePoolID),
			Encrypt: true,
		},
	}

	if module != nil {
		config.Modules = map[string]interface{}{
			"node-pool": module,
		}

		config.Output = map[string]terraform.ConfigOutput{
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
		}
	}

	return config
}

func (cp *DefaultAWSCloudProvider) nodePoolTerraformModule(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
	bootstrapSecret string,
) *kubetf.NodePool {
	nodePoolID := nodePool.ID(epoch)

	//bootstrapTokenId := bootstrapSecret.StringData["token-id"]
	//bootstrapTokenSecret := bootstrapSecret.StringData["token-secret"]
	//fmt.Println("full bootstrap token string data: %v", bootstrapSecret.StringData)
	//fmt.Println("bootstrap token ID: %v", bootstrapTokenId)
	fmt.Println("bootstrap token secret: %v", bootstrapSecret)
	//bootstrapSecretString := bootstrapTokenId + "." + bootstrapTokenSecret

	//// TODO bubble up error
	apiServerPort, err := strconv.ParseInt(cp.ApiServerPort, 10, 64)
	if err != nil {
		fmt.Sprintf("error parsing apiserver port: %v", err)
		return &kubetf.NodePool{}
	}
	fmt.Printf("aws:node_pool#nodePoolTerraformModule: controller manager addr and port: %v:%v", cp.ApiServerAddress, apiServerPort)

	return &kubetf.NodePool{
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

		KubeBootstrapToken:      bootstrapSecret,
		LatticeApiServerAddress: cp.ApiServerAddress,
		LatticeApiServerPort:    apiServerPort,
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
