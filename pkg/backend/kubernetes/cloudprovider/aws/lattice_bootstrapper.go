package aws

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"
)

type LatticeBootstrapperOptions struct {
	Region    string
	AccountID string
	VPCID     string

	Route53PrivateZoneID      string
	SubnetIDs                 []string
	MasterNodeSecurityGroupID string

	WorkerNodeAMIID string
	KeyName         string

	ControllerManagerOptions *LatticeBootstrapperControllerManagerOptions
}

type LatticeBootstrapperControllerManagerOptions struct {
	TerraformModulePath     string
	TerraformBackendOptions *terraform.BackendOptions
}

func NewLatticeBootstrapper(options *LatticeBootstrapperOptions) *DefaultAWSLatticeBootstrapper {
	return &DefaultAWSLatticeBootstrapper{
		region:    options.Region,
		accountID: options.AccountID,
		vpcID:     options.VPCID,

		route53PrivateZoneID:      options.Route53PrivateZoneID,
		subnetIDs:                 options.SubnetIDs,
		masterNodeSecurityGroupID: options.MasterNodeSecurityGroupID,

		workerNodeAMIID: options.WorkerNodeAMIID,
		keyName:         options.KeyName,

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
		"region": &flags.String{
			Required: true,
			Target:   &options.Region,
		},
		"account-id": &flags.String{
			Required: true,
			Target:   &options.AccountID,
		},
		"vpc-id": &flags.String{
			Required: true,
			Target:   &options.VPCID,
		},

		"route53-private-zone-id": &flags.String{
			Required: true,
			Target:   &options.Route53PrivateZoneID,
		},
		"subnet-ids": &flags.StringSlice{
			Required: true,
			Target:   &options.SubnetIDs,
		},
		"master-node-security-group-id": &flags.String{
			Required: true,
			Target:   &options.MasterNodeSecurityGroupID,
		},
		"worker-node-ami-id": &flags.String{
			Required: true,
			Target:   &options.WorkerNodeAMIID,
		},
		"key-name": &flags.String{
			Required: true,
			Target:   &options.KeyName,
		},

		"controller-manager-var": &flags.Embedded{
			Required: true,
			Flags: cli.Flags{
				"terraform-module-path": &flags.String{
					Default: "/etc/terraform/modules/aws",
					Target:  &options.ControllerManagerOptions.TerraformModulePath,
				},
				"terraform-backend": &flags.String{
					Required: true,
					Target:   &terraformBackend,
				},
				"terraform-backend-var": terraformBackendFlag,
			},
		},

		// FIXME(kevindrosendahl): think this can be removed but leaving it just in case things break
		//terraformBackendFlag,
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

	workerNodeAMIID string
	keyName         string

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
		)

		for _, flag := range cp.controllerManagerOptions.TerraformBackendOptions.AsFlags() {
			daemonSet.Spec.Template.Spec.Containers[0].Args = append(
				daemonSet.Spec.Template.Spec.Containers[0].Args,
				"--cloud-provider-var", flag,
			)
		}
	}
}
