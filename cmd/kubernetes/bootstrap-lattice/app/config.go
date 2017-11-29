package app

import (
	"fmt"
	"strings"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crdclient "github.com/mlab-lattice/system/pkg/kubernetes/customresource"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

func seedConfig(kubeconfig *rest.Config, userSystemUrl string) {
	fmt.Println("Seeding lattice config...")
	crClient, _, err := crdclient.NewClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	// Create config
	config := &crv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ConfigGlobal,
			Namespace: constants.NamespaceLatticeInternal,
		},
		Spec: crv1.ConfigSpec{
			SystemConfigs: map[coretypes.LatticeNamespace]crv1.ConfigSystem{
				coreconstants.UserSystemNamespace: {
					Url: userSystemUrl,
				},
			},
			Envoy: crv1.ConfigEnvoy{
				PrepareImage:      latticeContainerRegistry + "/envoy-prepare-envoy",
				Image:             "envoyproxy/envoy-alpine",
				RedirectCidrBlock: "172.16.29.0/16",
				XdsApiPort:        8080,
			},
			ComponentBuild: crv1.ConfigComponentBuild{
				DockerConfig: crv1.ConfigBuildDocker{
					RepositoryPerImage: false,
					Repository:         constants.DockerRegistryComponentBuildsDefault,
					Push:               true,
					Registry:           componentBuildRegistry,
				},
				BuildImage: latticeContainerRegistry + "/kubernetes-component-builder",
			},
			SystemId: systemId,
		},
	}

	switch provider {
	case coreconstants.ProviderLocal:
		config.Spec.ComponentBuild.DockerConfig.Push = false

		localConfig, err := getLocalConfig()
		if err != nil {
			panic(err)
		}

		config.Spec.Provider.Local = localConfig
	case coreconstants.ProviderAWS:
		awsConfig, err := getAwsConfig()
		if err != nil {
			panic(err)
		}
		config.Spec.Provider.AWS = awsConfig

		terraformConfig, err := getTerraformConfig()
		if err != nil {
			panic(err)
		}
		config.Spec.Terraform = terraformConfig
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return nil, crClient.Post().
			Namespace(constants.NamespaceLatticeInternal).
			Resource(crv1.ConfigResourcePlural).
			Body(config).
			Do().Into(nil)
	})
}

func getLocalConfig() (*crv1.ConfigProviderLocal, error) {
	// TODO: find a better way to do the parsing of the provider variables
	expectedVars := map[string]interface{}{
		"system-ip": nil,
	}

	for _, providerVar := range *providerVars {
		split := strings.Split(providerVar, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid provider variable " + providerVar)
		}

		key := split[0]
		var value interface{} = split[1]

		existingVal, ok := expectedVars[key]
		if !ok {
			return nil, fmt.Errorf("unexpected provider variable " + key)
		}
		if existingVal != nil {
			return nil, fmt.Errorf("provider variable " + key + " set multiple times")
		}

		expectedVars[key] = value
	}

	for k, v := range expectedVars {
		if v == nil {
			return nil, fmt.Errorf("missing required provider variable " + k)
		}
	}

	localConfig := &crv1.ConfigProviderLocal{
		IP: expectedVars["system-ip"].(string),
	}

	return localConfig, nil
}

func getAwsConfig() (*crv1.ConfigProviderAWS, error) {
	// TODO: find a better way to do the parsing of the provider variables
	expectedVars := map[string]interface{}{
		"account-id":                    nil,
		"region":                        nil,
		"vpc-id":                        nil,
		"subnet-ids":                    nil,
		"master-node-security-group-id": nil,
		"base-node-ami-id":              nil,
		"key-name":                      nil,
	}

	for _, providerVar := range *providerVars {
		split := strings.Split(providerVar, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid provider variable " + providerVar)
		}

		key := split[0]
		var value interface{} = split[1]

		existingVal, ok := expectedVars[key]
		if !ok {
			return nil, fmt.Errorf("unexpected provider variable " + key)
		}
		if existingVal != nil {
			return nil, fmt.Errorf("provider variable " + key + " set multiple times")
		}

		if key == "subnet-ids" {
			value = strings.Split(value.(string), ",")
		}

		expectedVars[key] = value
	}

	for k, v := range expectedVars {
		if v == nil {
			return nil, fmt.Errorf("missing required provider variable " + k)
		}
	}

	awsConfig := &crv1.ConfigProviderAWS{
		Region:                    expectedVars["region"].(string),
		AccountId:                 expectedVars["account-id"].(string),
		VPCId:                     expectedVars["vpc-id"].(string),
		SubnetIds:                 expectedVars["subnet-ids"].([]string),
		MasterNodeSecurityGroupID: expectedVars["master-node-security-group-id"].(string),
		BaseNodeAMIId:             expectedVars["base-node-ami-id"].(string),
		KeyName:                   expectedVars["key-name"].(string),
	}

	return awsConfig, nil
}

func getTerraformConfig() (*crv1.ConfigTerraform, error) {
	switch terraformBackend {
	case constants.TerraformBackendS3:
		backendConfigS3, err := getTerraformBackendConfigS3()
		if err != nil {
			return nil, err
		}

		terraformConfig := &crv1.ConfigTerraform{
			S3Backend: backendConfigS3,
		}
		return terraformConfig, nil
	default:
		return nil, fmt.Errorf("unrecognized terraform backend " + terraformBackend)
	}
}

func getTerraformBackendConfigS3() (*crv1.ConfigTerraformBackendS3, error) {
	// TODO: find a better way to do the parsing of the provider variables
	expectedVars := map[string]interface{}{
		"bucket": nil,
	}

	for _, terraformBackendVar := range *terraformBackendVars {
		split := strings.Split(terraformBackendVar, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid terraform backend variable " + terraformBackendVar)
		}

		key := split[0]
		var value interface{} = split[1]

		existingVal, ok := expectedVars[key]
		if !ok {
			return nil, fmt.Errorf("unexpected terraform backend variable " + key)
		}
		if existingVal != nil {
			return nil, fmt.Errorf("terraform backend variable " + key + " set multiple times")
		}

		expectedVars[key] = value
	}

	for k, v := range expectedVars {
		if v == nil {
			return nil, fmt.Errorf("missing required terraform backend variable " + k)
		}
	}

	terraformBackendConfigS3 := &crv1.ConfigTerraformBackendS3{
		Bucket: expectedVars["bucket"].(string),
	}

	return terraformBackendConfigS3, nil
}
