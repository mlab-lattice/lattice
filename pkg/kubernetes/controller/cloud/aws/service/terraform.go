package service

import (
	"encoding/json"
	"fmt"
	"strings"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
	kubetf "github.com/mlab-lattice/system/pkg/kubernetes/terraform/aws"
	tf "github.com/mlab-lattice/system/pkg/terraform"
	tfconfig "github.com/mlab-lattice/system/pkg/terraform/config"
	awstf "github.com/mlab-lattice/system/pkg/terraform/config/aws"

	corev1 "k8s.io/api/core/v1"
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

		svcTf, err := sc.getServiceTerraformConfig(svc)
		if err != nil {
			return err
		}

		svcTfConfig = svcTf
	}

	tec, err := tf.NewTerrafromExecContext(getWorkingDirectory(svc), nil)
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

	result, _, err := tec.Init()
	if err != nil {
		return err
	}

	err = result.Wait()
	if err != nil {
		return err
	}

	result, _, err = tec.Apply(nil)
	if err != nil {
		return err
	}

	return result.Wait()
}

func (sc *ServiceController) deprovisionService(svc *crv1.Service) error {
	var svcTfConfig interface{}
	{
		// Need a consistent view of our config while generating the config
		sc.configLock.RLock()
		defer sc.configLock.RUnlock()

		svcTf, err := sc.getServiceTerraformConfig(svc)
		if err != nil {
			return err
		}

		svcTfConfig = svcTf
	}

	tec, err := tf.NewTerrafromExecContext(getWorkingDirectory(svc), nil)
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

	result, _, err := tec.Init()
	if err != nil {
		return err
	}

	err = result.Wait()
	if err != nil {
		return err
	}

	result, _, err = tec.Destroy(nil)
	if err != nil {
		return err
	}

	err = result.Wait()
	if err != nil {
		return err
	}

	return sc.removeFinalizer(svc)
}

func (sc *ServiceController) getServiceTerraformConfig(svc *crv1.Service) (interface{}, error) {
	kubeSvc, necessary, err := sc.getKubeServiceForService(svc)
	if err != nil {
		return nil, err
	}

	var serviceModule interface{}
	if necessary {
		if kubeSvc == nil {
			return nil, fmt.Errorf("Service %v requires kubeSvc but it does not exist")
		}

		serviceModule = sc.getServiceDedicatedPublicHttpTerraformModule(svc, kubeSvc)
	} else {
		serviceModule = sc.getServiceDedicatedPrivateTerraformModule(svc)
	}

	awsConfig := sc.config.Provider.AWS

	config := tfconfig.Config{
		Provider: awstf.Provider{
			Region: awsConfig.Region,
		},
		Backend: awstf.S3Backend{
			Region: awsConfig.Region,
			Bucket: sc.config.Terraform.S3Backend.Bucket,
			Key: fmt.Sprintf("%v%v/%v",
				kubetf.GetS3BackendStatePathRoot(sc.config.SystemId),
				terraformStatePathService,
				svc.Name),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"service": serviceModule,
		},
	}

	return config, nil
}

func (sc *ServiceController) getServiceDedicatedPrivateTerraformModule(svc *crv1.Service) interface{} {
	awsConfig := sc.config.Provider.AWS

	return kubetf.ServiceDedicatedPrivate{
		Source: sc.terraformModulePath + kubetf.ModulePathServiceDedicatedPrivate,

		AWSAccountId: awsConfig.AccountId,
		Region:       awsConfig.Region,

		VPCId:         awsConfig.VPCId,
		SubnetIds:     strings.Join(awsConfig.SubnetIds, ","),
		BaseNodeAmiId: awsConfig.BaseNodeAMIId,
		KeyName:       awsConfig.KeyName,

		SystemId:  sc.config.SystemId,
		ServiceId: svc.Name,
		// FIXME: support min/max instances
		NumInstances: *svc.Spec.Definition.Resources.NumInstances,
		InstanceType: *svc.Spec.Definition.Resources.InstanceType,
	}
}

func (sc *ServiceController) getServiceDedicatedPublicHttpTerraformModule(svc *crv1.Service, kubeSvc *corev1.Service) interface{} {
	awsConfig := sc.config.Provider.AWS

	publicComponentPorts := map[int32]bool{}
	for _, component := range svc.Spec.Definition.Components {
		for _, port := range component.Ports {
			if port.ExternalAccess != nil && port.ExternalAccess.Public {
				publicComponentPorts[port.Port] = true
			}
		}
	}

	ports := map[int32]int32{}
	for _, port := range kubeSvc.Spec.Ports {
		if _, ok := publicComponentPorts[port.Port]; ok {
			ports[port.Port] = port.NodePort
		}
	}

	return kubetf.ServiceDedicatedPublicHttp{
		Source: sc.terraformModulePath + kubetf.ModulePathServiceDedicatedPublicHttp,

		AWSAccountId: awsConfig.AccountId,
		Region:       awsConfig.Region,

		VPCId:         awsConfig.VPCId,
		SubnetIds:     strings.Join(awsConfig.SubnetIds, ","),
		BaseNodeAmiId: awsConfig.BaseNodeAMIId,
		KeyName:       awsConfig.KeyName,

		SystemId:  sc.config.SystemId,
		ServiceId: svc.Name,
		// FIXME: support min/max instances
		NumInstances: *svc.Spec.Definition.Resources.NumInstances,
		InstanceType: *svc.Spec.Definition.Resources.InstanceType,

		Ports: ports,
	}
}

func getWorkingDirectory(svc *crv1.Service) string {
	return "/tmp/lattice-controller-manager/controllers/cloud/aws/service/terraform/" + svc.Name
}
