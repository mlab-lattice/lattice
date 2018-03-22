package aws

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/cli/command"
)

type LatticeBootstrapperOptions struct {
	Region    string
	AccountID string
	VPCID     string

	Route53PrivateZoneID      string
	SubnetIDs                 []string
	MasterNodeSecurityGroupID string

	BaseNodeAMIID string
	KeyName       string
}

func NewLatticeBootstrapper(options *LatticeBootstrapperOptions) *DefaultAWSLatticeBootstrapper {
	return &DefaultAWSLatticeBootstrapper{
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

func LatticeBootstrapperFlags() (command.Flags, *LatticeBootstrapperOptions) {
	options := &LatticeBootstrapperOptions{}
	flags := command.Flags{
		&command.StringFlag{
			Name:     "region",
			Required: true,
			Target:   &options.Region,
		},
		&command.StringFlag{
			Name:     "account-id",
			Required: true,
			Target:   &options.AccountID,
		},
		&command.StringFlag{
			Name:     "vpc-id",
			Required: true,
			Target:   &options.VPCID,
		},

		&command.StringFlag{
			Name:     "route-53-private-zone-id",
			Required: true,
			Target:   &options.Route53PrivateZoneID,
		},
		&command.StringSliceFlag{
			Name:     "subnet-ids",
			Required: true,
			Target:   &options.SubnetIDs,
		},
		&command.StringFlag{
			Name:     "master-node-security-group-id",
			Required: true,
			Target:   &options.MasterNodeSecurityGroupID,
		},
		&command.StringFlag{
			Name:     "base-node-ami-id",
			Required: true,
			Target:   &options.BaseNodeAMIID,
		},
		&command.StringFlag{
			Name:     "key-name",
			Required: true,
			Target:   &options.KeyName,
		},
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

	baseNodeAMIID string
	keyName       string
}

func (cp *DefaultAWSLatticeBootstrapper) BootstrapLatticeResources(resources *bootstrapper.Resources) {
	resources.Config.Spec.CloudProvider.AWS = &latticev1.ConfigCloudProviderAWS{
		BaseNodeAMIID: cp.baseNodeAMIID,
		KeyName:       cp.keyName,
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
		)
	}
}
