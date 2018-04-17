package aws

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	workDirectoryVolumeHostPathPrefix = "/var/lib/component-builder"

	AnnotationKeyLoadBalancerDNSName          = "load-balancer.aws.cloud-provider.lattice.mlab.com/dns-name"
	AnnotationKeyNodePoolAutoscalingGroupName = "node-pool.aws.cloud-provider.lattice.mlab.com/autoscaling-group-name"
	AnnotationKeyNodePoolSecurityGroupID      = "node-pool.aws.cloud-provider.lattice.mlab.com/security-group-id"

	terraformOutputAutoscalingGroupID              = "autoscaling_group_id"
	terraformOutputAutoscalingGroupName            = "autoscaling_group_name"
	terraformOutputAutoscalingGroupDesiredCapacity = "autoscaling_group_desired_capacity"
	terraformOutputSecurityGroupID                 = "security_group_id"
)

type Options struct {
	Region    string
	AccountID string
	VPCID     string

	Route53PrivateZoneID      string
	SubnetIDs                 []string
	MasterNodeSecurityGroupID string

	WorkerNodeAMIID string
	KeyName         string

	TerraformModulePath     string
	TerraformBackendOptions *terraform.BackendOptions
}

func NewOptions(staticOptions *Options, dynamicConfig *latticev1.ConfigCloudProviderAWS) (*Options, error) {
	options := &Options{
		Region:    staticOptions.Region,
		AccountID: staticOptions.AccountID,
		VPCID:     staticOptions.VPCID,

		Route53PrivateZoneID:      staticOptions.Route53PrivateZoneID,
		SubnetIDs:                 staticOptions.SubnetIDs,
		MasterNodeSecurityGroupID: staticOptions.MasterNodeSecurityGroupID,

		WorkerNodeAMIID: dynamicConfig.WorkerNodeAMIID,
		KeyName:         dynamicConfig.KeyName,

		TerraformModulePath:     staticOptions.TerraformModulePath,
		TerraformBackendOptions: staticOptions.TerraformBackendOptions,
	}
	return options, nil
}

func NewCloudProvider(options *Options) *DefaultAWSCloudProvider {
	return &DefaultAWSCloudProvider{
		region:    options.Region,
		accountID: options.AccountID,
		vpcID:     options.VPCID,

		route53PrivateZoneID:      options.Route53PrivateZoneID,
		subnetIDs:                 options.SubnetIDs,
		masterNodeSecurityGroupID: options.MasterNodeSecurityGroupID,

		workerNodeAMIID: options.WorkerNodeAMIID,
		keyName:         options.KeyName,

		terraformModulePath:     options.TerraformModulePath,
		terraformBackendOptions: options.TerraformBackendOptions,
	}
}

func Flags() (cli.Flags, *Options) {
	var terraformBackend string
	terraformBackendFlag, terraformBackendOptions := terraform.BackendFlags(&terraformBackend)
	options := &Options{
		TerraformBackendOptions: terraformBackendOptions,
	}

	flags := cli.Flags{
		&cli.StringFlag{
			Name:     "region",
			Required: true,
			Target:   &options.Region,
		},
		&cli.StringFlag{
			Name:     "account-id",
			Required: true,
			Target:   &options.AccountID,
		},
		&cli.StringFlag{
			Name:     "vpc-id",
			Required: true,
			Target:   &options.VPCID,
		},

		&cli.StringFlag{
			Name:     "route53-private-zone-id",
			Required: true,
			Target:   &options.Route53PrivateZoneID,
		},
		&cli.StringSliceFlag{
			Name:     "subnet-ids",
			Required: true,
			Target:   &options.SubnetIDs,
		},
		&cli.StringFlag{
			Name:     "master-node-security-group-id",
			Required: true,
			Target:   &options.MasterNodeSecurityGroupID,
		},
		// worker-node-ami-id and key-name should be set with dynamic config (i.e. custom resource)
		&cli.StringFlag{
			Name:    "terraform-module-path",
			Default: "/etc/terraform/modules/kubernetes/aws",
			Target:  &options.TerraformModulePath,
		},
		&cli.StringFlag{
			Name:     "terraform-backend",
			Required: true,
			Target:   &terraformBackend,
		},
		terraformBackendFlag,
	}
	return flags, options
}

type DefaultAWSCloudProvider struct {
	region    string
	accountID string
	vpcID     string

	route53PrivateZoneID      string
	subnetIDs                 []string
	masterNodeSecurityGroupID string

	workerNodeAMIID string
	keyName         string

	terraformModulePath     string
	terraformBackendOptions *terraform.BackendOptions
}

func (cp *DefaultAWSCloudProvider) BootstrapSystemResources(resources *bootstrapper.SystemResources) {
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

func (cp *DefaultAWSCloudProvider) TransformServiceDeploymentSpec(
	service *latticev1.Service,
	spec *appsv1.DeploymentSpec,
) *appsv1.DeploymentSpec {
	// nothing to do
	return spec
}

func (cp *DefaultAWSCloudProvider) IsDeploymentSpecUpdated(
	service *latticev1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string, *appsv1.DeploymentSpec) {
	// nothing to do
	return true, "", current
}

func workDirectory(resourceType, resourceID string) string {
	return fmt.Sprintf("/tmp/lattice/cloud-provider/aws/%v/terraform/%v", resourceType, resourceID)
}
