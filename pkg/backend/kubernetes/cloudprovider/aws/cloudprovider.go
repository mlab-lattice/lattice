package aws

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	workDirectoryVolumeHostPathPrefix = "/var/lib/component-builder"

	AnnotationKeyLoadBalancerDNSName = "load-balancer.aws.cloud-provider.lattice.mlab.com/dns-name"
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

func NewCloudProvider(
	namespacePrefix string,
	kubeClient kubeclientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	latticeInformerFactory latticeinformers.SharedInformerFactory,
	options *Options,
) (*DefaultAWSCloudProvider, error) {
	s, err := session.NewSession(&aws.Config{Region: &options.Region})
	if err != nil {
		return nil, err
	}

	cp := &DefaultAWSCloudProvider{
		namespacePrefix: namespacePrefix,

		session: s,

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

		kubeClient:        kubeClient,
		kubeServiceLister: kubeInformerFactory.Core().V1().Services().Lister(),
		kubeNodeLister:    kubeInformerFactory.Core().V1().Nodes().Lister(),

		nodePoolLister: latticeInformerFactory.Lattice().V1().NodePools().Lister(),
	}

	// shared informer factories only start informers which have been referenced
	// by the time start is called
	// so we'll start the factories back up here in case the factories were started
	// before without having these listers referenced
	kubeInformerFactory.Start(nil)
	latticeInformerFactory.Start(nil)

	// wait for secondary caches to fill
	if !cache.WaitForCacheSync(
		nil,
		kubeInformerFactory.Core().V1().Nodes().Informer().HasSynced,
		kubeInformerFactory.Core().V1().Services().Informer().HasSynced,
		latticeInformerFactory.Lattice().V1().NodePools().Informer().HasSynced,
	) {
		return nil, fmt.Errorf("failed to sync caches for aws cloud provider")
	}

	return cp, nil
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
		&cli.StringFlag{
			Name:    "terraform-module-path",
			Default: "/etc/terraform/modules/aws",
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
	namespacePrefix string

	session *session.Session

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

	kubeClient kubeclientset.Interface

	kubeServiceLister corelisters.ServiceLister
	kubeNodeLister    corelisters.NodeLister

	nodePoolLister latticelisters.NodePoolLister
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

func (cp *DefaultAWSCloudProvider) TransformPodTemplateSpec(spec *corev1.PodTemplateSpec) *corev1.PodTemplateSpec {
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
