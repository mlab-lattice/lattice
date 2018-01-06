package aws

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	workDirectoryVolumeHostPathPrefix = "/var/lib/component-builder"

	AnnotationKeyNodePoolAutoscalingGroupName = "aws.cloud-provider.lattice.mlab.com/node-pool-autoscaling-group-name"
	AnnotationKeyNodePoolSecurityGroupID      = "aws.cloud-provider.lattice.mlab.com/node-pool-security-group-id"
)

type Options struct {
	Region    string
	AccountID string
	VPCID     string

	Route53PrivateZoneID      string
	SubnetIDs                 []string
	MasterNodeSecurityGroupID string

	BaseNodeAMIID string
	KeyName       string
}

type CloudProvider interface {
	Region() string
	AccountID() string
	VPCID() string

	Route53PrivateZoneID() string
	SubnetIDs() []string
	MasterNodeSecurityGroupID() string

	BaseNodeAMIID() string
	KeyName() string
}

func NewAWSCloudProvider(options *Options) *DefaultAWSCloudProvider {
	return &DefaultAWSCloudProvider{
		region:    options.Region,
		accountID: options.AccountID,
		vpcID:     options.VPCID,

		route53PrivateZoneID:      options.Route53PrivateZoneID,
		subnetIDs:                 options.SubnetIDs,
		masterNodeSecurityGroupID: options.MasterNodeSecurityGroupID,

		baseNodeAMIID: options.BaseNodeAMIID,
		keyName:       options.KeyName,
	}
}

type DefaultAWSCloudProvider struct {
	region    string
	accountID string
	vpcID     string

	route53PrivateZoneID      string
	subnetIDs                 []string
	masterNodeSecurityGroupID string

	baseNodeAMIID string
	keyName       string
}

func (cp *DefaultAWSCloudProvider) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	resources.Config.Spec.CloudProvider.AWS = &crv1.ConfigCloudProviderAWS{
		BaseNodeAMIID: cp.baseNodeAMIID,
		KeyName:       cp.keyName,
	}

	for _, daemonSet := range resources.DaemonSets {
		if daemonSet.Name != kubeconstants.MasterNodeComponentLatticeControllerManager {
			continue
		}

		daemonSet.Spec.Template.Spec.Containers[0].Args = append(
			daemonSet.Spec.Template.Spec.Containers[0].Args,
			"--cloud-provider-var", fmt.Sprintf("region=%v", cp.region),
			"--cloud-provider-var", fmt.Sprintf("account-id=%v", cp.accountID),
			"--cloud-provider-var", fmt.Sprintf("vpc-id=%v", cp.vpcID),
			"--cloud-provider-var", fmt.Sprintf("route53-private-zone-id=%v", cp.route53PrivateZoneID),
			"--cloud-provider-var", fmt.Sprintf("subnet-ids=%v", strings.Join(cp.subnetIDs, ",")),
			"--cloud-provider-var", fmt.Sprintf("master-node-security-group-id=%v", cp.masterNodeSecurityGroupID),
		)
	}
}

func (cp *DefaultAWSCloudProvider) BootstrapSystemResources(resources *systembootstrapper.SystemResources) {
	// nothing to do
}

func (cp *DefaultAWSCloudProvider) TransformComponentBuildJobSpec(spec *batchv1.JobSpec) *batchv1.JobSpec {
	// nothing to do
	return spec
}

func (cp *DefaultAWSCloudProvider) ComponentBuildWorkDirectoryVolumeSource(jobName string) corev1.VolumeSource {
	return corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: workDirectoryVolumeHostPathPrefix + "/" + jobName,
		},
	}
}

func (cp *DefaultAWSCloudProvider) TransformServiceDeploymentSpec(service *crv1.Service, spec *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	// nothing to do
	return spec
}

func (cp *DefaultAWSCloudProvider) IsDeploymentSpecUpdated(
	service *crv1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string, *appsv1.DeploymentSpec) {
	// nothing to do
	return true, "", current
}

func (cp *DefaultAWSCloudProvider) Region() string {
	return cp.region
}

func (cp *DefaultAWSCloudProvider) AccountID() string {
	return cp.accountID
}

func (cp *DefaultAWSCloudProvider) VPCID() string {
	return cp.vpcID
}

func (cp *DefaultAWSCloudProvider) Route53PrivateZoneID() string {
	return cp.route53PrivateZoneID
}

func (cp *DefaultAWSCloudProvider) SubnetIDs() []string {
	return cp.subnetIDs
}

func (cp *DefaultAWSCloudProvider) MasterNodeSecurityGroupID() string {
	return cp.masterNodeSecurityGroupID
}

func (cp *DefaultAWSCloudProvider) BaseNodeAMIID() string {
	return cp.baseNodeAMIID
}

func (cp *DefaultAWSCloudProvider) KeyName() string {
	return cp.keyName
}
