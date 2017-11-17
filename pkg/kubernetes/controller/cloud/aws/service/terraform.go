package service

import (
	"encoding/json"

	"fmt"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
	kubetf "github.com/mlab-lattice/system/pkg/kubernetes/terraform/aws"
	tf "github.com/mlab-lattice/system/pkg/terraform"
	tfconfig "github.com/mlab-lattice/system/pkg/terraform/config"
	awstf "github.com/mlab-lattice/system/pkg/terraform/config/aws"
)

const (
	terraformStatePathService = "/services"
)

func (sc *ServiceController) provisionService(svc *crv1.Service) error {
	var svcTfConfig interface{}
	{
		// Need a consistent view of our config while generating the config
		sc.configLock.RLock()
		defer sc.configLock.RUnlock()

		svcTfConfig = sc.getServiceTerraformConfig(svc)
	}

	tec, err := tf.NewTerrafromExecContext("workingDir", nil)
	if err != nil {
		return err
	}

	svcTfConfigBytes, err := json.Marshal(svcTfConfig)
	if err != nil {
		return err
	}

	err = tec.AddFile("config.tf", svcTfConfigBytes)
	if err != nil {
		return err
	}

	result, _, err := tec.Apply(nil)
	if err != nil {
		return err
	}

	return result.Wait()
}

func (sc *ServiceController) getServiceTerraformConfig(svc *crv1.Service) interface{} {
	awsConfig := sc.config.Provider.AWS

	return tfconfig.Config{
		Provider: awstf.Provider{
			Region: awsConfig.Region,
		},
		Backend: awstf.S3Backend{
			Bucket: sc.config.Terraform.S3Backend.Bucket,
			Key: fmt.Sprintf("%v/%v/%v",
				kubetf.GetS3BackendStatePathRoot(sc.config.SystemId),
				terraformStatePathService,
				svc.Name),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"service": kubetf.ServiceDedicatedPrivate{
				Source: kubetf.ModulePathServiceDedicatedPrivate,

				AWSAccountId: awsConfig.AccountId,
				Region:       awsConfig.Region,

				VPCId:         awsConfig.VPCId,
				SubnetIds:     awsConfig.SubnetIds,
				BaseNodeAmiId: awsConfig.BaseNodeAMIId,
				KeyName:       awsConfig.KeyName,

				SystemId:  sc.config.SystemId,
				ServiceId: svc.Name,
				// FIXME: support min/max instances
				NumInstances: *svc.Spec.Definition.Resources.NumInstances,
				InstanceType: *svc.Spec.Definition.Resources.InstanceType,
			},
		},
	}
}
