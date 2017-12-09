package app

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func seedConfig(userSystemURL string) {
	fmt.Println("Seeding lattice config...")

	// Create config
	config := &crv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ConfigGlobal,
			Namespace: kubeconstants.NamespaceLatticeInternal,
		},
		Spec: crv1.ConfigSpec{
			SystemConfigs: map[types.LatticeNamespace]crv1.ConfigSystem{
				constants.UserSystemNamespace: {
					DefinitionURL: userSystemURL,
				},
			},
			Envoy: crv1.ConfigEnvoy{
				PrepareImage:      getContainerImageFQN(constants.DockerImageEnvoyPrepare),
				Image:             "envoyproxy/envoy-alpine",
				RedirectCidrBlock: "172.16.29.0/16",
				XDSAPIPort:        8080,
			},
			ComponentBuild: crv1.ConfigComponentBuild{
				DockerArtifact: crv1.ConfigComponentBuildDockerArtifact{
					RepositoryPerImage: false,
					Repository:         kubeconstants.DockerRegistryComponentBuildsDefault,
					Push:               true,
					Registry:           componentBuildRegistry,
				},
				Builder: crv1.ConfigComponentBuildBuilder{
					Image:            getContainerImageFQN(kubeconstants.DockerImageComponentBuilder),
					DockerAPIVersion: dockerAPIVersion,
				},
			},
			KubernetesNamespacePrefix: systemID,
		},
	}

	switch provider {
	case constants.ProviderLocal:
		config.Spec.ComponentBuild.DockerArtifact.Push = false

		localConfig, err := getLocalConfig()
		if err != nil {
			panic(err)
		}

		config.Spec.Provider.Local = localConfig
	case constants.ProviderAWS:
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
		return latticeClient.LatticeV1().Configs(kubeconstants.NamespaceLatticeInternal).Create(config)
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
		AccountID:                 expectedVars["account-id"].(string),
		VPCID:                     expectedVars["vpc-id"].(string),
		SubnetIDs:                 expectedVars["subnet-ids"].([]string),
		MasterNodeSecurityGroupID: expectedVars["master-node-security-group-id"].(string),
		BaseNodeAMIID:             expectedVars["base-node-ami-id"].(string),
		KeyName:                   expectedVars["key-name"].(string),
	}

	return awsConfig, nil
}

func getTerraformConfig() (*crv1.ConfigTerraform, error) {
	switch terraformBackend {
	case kubeconstants.TerraformBackendS3:
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
