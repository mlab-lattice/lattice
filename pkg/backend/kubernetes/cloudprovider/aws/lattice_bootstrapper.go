package aws

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/util/cli"
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

func LatticeBootstrapperFlags() (cli.Flags, *LatticeBootstrapperOptions) {
	options := &LatticeBootstrapperOptions{}
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
			Name:     "route-53-private-zone-id",
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
			Name:     "base-node-ami-id",
			Required: true,
			Target:   &options.BaseNodeAMIID,
		},
		&cli.StringFlag{
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
