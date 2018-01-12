package aws

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	workDirectoryVolumeHostPathPrefix = "/var/lib/component-builder"

	AnnotationKeyLoadBalancerDNSName          = "load-balancer.aws.cloud-provider.lattice.mlab.com/dns-name"
	AnnotationKeyNodePoolAutoscalingGroupName = "node-pool.aws.cloud-provider.lattice.mlab.com/autoscaling-group-name"
	AnnotationKeyNodePoolSecurityGroupID      = "node-pool.aws.cloud-provider.lattice.mlab.com/security-group-id"
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

func NewCloudProvider(options *Options) *DefaultAWSCloudProvider {
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
