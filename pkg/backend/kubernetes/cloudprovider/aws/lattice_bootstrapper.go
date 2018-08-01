package aws

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"
)

type LatticeBootstrapperOptions struct {
	Region    string
	AccountID string
	VPCID     string

	Route53PrivateZoneID      string
	SubnetIDs                 []string
	MasterNodeSecurityGroupID string

	WorkerNodeAMIID  string
	KeyName          string
	ApiServerAddress string
	ApiServerPort    string

	ControllerManagerOptions *LatticeBootstrapperControllerManagerOptions
}

type LatticeBootstrapperControllerManagerOptions struct {
	TerraformModulePath     string
	TerraformBackendOptions *terraform.BackendOptions
}

func NewLatticeBootstrapper(internalDnsDomain string, options *LatticeBootstrapperOptions) *DefaultAWSLatticeBootstrapper {
	return &DefaultAWSLatticeBootstrapper{
		region:    options.Region,
		accountID: options.AccountID,
		vpcID:     options.VPCID,

		route53PrivateZoneID:      options.Route53PrivateZoneID,
		subnetIDs:                 options.SubnetIDs,
		masterNodeSecurityGroupID: options.MasterNodeSecurityGroupID,

		workerNodeAMIID:   options.WorkerNodeAMIID,
		keyName:           options.KeyName,
		apiServerAddress:  options.ApiServerAddress,
		apiServerPort:     options.ApiServerPort,
		internalDnsDomain: internalDnsDomain,

		controllerManagerOptions: options.ControllerManagerOptions,
	}
}

func LatticeBootstrapperFlags() (cli.Flags, *LatticeBootstrapperOptions) {
	var terraformBackend string
	terraformBackendFlag, terraformBackendOptions := terraform.BackendFlags(&terraformBackend)

	options := &LatticeBootstrapperOptions{
		ControllerManagerOptions: &LatticeBootstrapperControllerManagerOptions{
			TerraformBackendOptions: terraformBackendOptions,
		},
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
			Name:     "worker-node-ami-id",
			Required: true,
			Target:   &options.WorkerNodeAMIID,
		},
		&cli.StringFlag{
			Name:     "key-name",
			Required: true,
			Target:   &options.KeyName,
		},
		&cli.StringFlag{
			Name:     "api-server-address",
			Required: false,
			Target:   &options.ApiServerAddress,
		},
		&cli.StringFlag{
			Name:     "api-server-port",
			Required: false,
			Target:   &options.ApiServerPort,
		},

		&cli.EmbeddedFlag{
			Name:     "controller-manager-var",
			Required: true,
			Flags: cli.Flags{
				&cli.StringFlag{
					Name:    "terraform-module-path",
					Default: "/etc/terraform/modules/aws",
					Target:  &options.ControllerManagerOptions.TerraformModulePath,
				},
				&cli.StringFlag{
					Name:     "terraform-backend",
					Required: true,
					Target:   &terraformBackend,
				},
				terraformBackendFlag,
			},
		},

		terraformBackendFlag,
	}
	return flags, options
}

type DefaultAWSLatticeBootstrapper struct {
	region    string
	accountID string
	vpcID     string

	route53PrivateZoneID      string
	subnetIDs                 []string
	masterNodeSecurityGroupID string

	workerNodeAMIID   string
	keyName           string
	apiServerAddress  string
	apiServerPort     string
	internalDnsDomain string

	controllerManagerOptions *LatticeBootstrapperControllerManagerOptions
}

func (cp *DefaultAWSLatticeBootstrapper) BootstrapLatticeResources(resources *bootstrapper.Resources) {
	resources.Config.Spec.CloudProvider.AWS = &latticev1.ConfigCloudProviderAWS{
		WorkerNodeAMIID: cp.workerNodeAMIID,
		KeyName:         cp.keyName,
	}

	for _, daemonSet := range resources.DaemonSets {
		if daemonSet.Name != kubeconstants.ControlPlaneServiceLatticeControllerManager {
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
			"--cloud-provider-var", fmt.Sprintf("terraform-module-path=%v", cp.controllerManagerOptions.TerraformModulePath),
			"--cloud-provider-var", fmt.Sprintf("api-server-address=%v.%v", cp.apiServerAddress, cp.internalDnsDomain),
			"--cloud-provider-var", fmt.Sprintf("api-server-port=%v", cp.apiServerPort),
		)

		for _, flag := range cp.controllerManagerOptions.TerraformBackendOptions.AsFlags() {
			daemonSet.Spec.Template.Spec.Containers[0].Args = append(
				daemonSet.Spec.Template.Spec.Containers[0].Args,
				"--cloud-provider-var", flag,
			)
		}
	}
}
