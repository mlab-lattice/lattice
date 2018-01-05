package aws

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	workDirectoryVolumeHostPathPrefix = "/var/lib/component-builder"
)

type Options struct {
	Region                    string
	AccountID                 string
	VPCID                     string
	SubnetIDs                 []string
	MasterNodeSecurityGroupID string
	BaseNodeAMIID             string
	KeyName                   string
}

func NewAWSCloudProvider(options *Options) *DefaultAWSCloudProvider {
	return &DefaultAWSCloudProvider{
		region:                    options.Region,
		accountID:                 options.AccountID,
		vpcID:                     options.VPCID,
		subnetIDs:                 options.SubnetIDs,
		masterNodeSecurityGroupID: options.MasterNodeSecurityGroupID,
		baseNodeAMIID:             options.BaseNodeAMIID,
		keyName:                   options.KeyName,
	}
}

type DefaultAWSCloudProvider struct {
	region                    string
	accountID                 string
	vpcID                     string
	subnetIDs                 []string
	masterNodeSecurityGroupID string
	baseNodeAMIID             string
	keyName                   string
}

func (cp *DefaultAWSCloudProvider) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	awsConfig := &crv1.ConfigCloudProviderAWS{
		BaseNodeAMIID: cp.baseNodeAMIID,
		KeyName:       cp.keyName,
	}
	resources.Config.Spec.CloudProvider.AWS = awsConfig
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
