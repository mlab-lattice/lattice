package v1

import (
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularConfig = "config"
	ResourcePluralConfig   = "configs"
	ResourceScopeConfig    = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ConfigSpec `json:"spec"`
}

type ConfigSpec struct {
	// FIXME: this shouldn't be dynamic config
	KubernetesNamespacePrefix string               `json:"systemID"`
	Provider                  ConfigProvider       `json:"providerConfig"`
	ComponentBuild            ConfigComponentBuild `json:"componentBuild"`
	Envoy                     ConfigEnvoy          `json:"envoy"`
	// FIXME: this shouldn't be dynamic config
	// FIXME: create empty System and add definition URL to system.Spec
	SystemConfigs map[types.LatticeNamespace]ConfigSystem `json:"userSystem"`
	Terraform     *ConfigTerraform                        `json:"terraform,omitempty"`
}

type ConfigProvider struct {
	Local *ConfigProviderLocal `json:"local,omitempty"`
	AWS   *ConfigProviderAWS   `json:"aws,omitempty"`
}

type ConfigProviderLocal struct {
	// FIXME: this shouldn't be dynamic config
	IP string `json:"ip"`
}

type ConfigProviderAWS struct {
	// FIXME: this shouldn't be dynamic config
	Region string `json:"region"`
	// FIXME: this shouldn't be dynamic config
	AccountID string `json:"accountID"`
	// FIXME: this shouldn't be dynamic config
	VPCID string `json:"vpcId"`
	// FIXME: this shouldn't be dynamic config
	SubnetIDs []string `json:"subnetIds"`
	// FIXME: this shouldn't be dynamic config
	MasterNodeSecurityGroupID string `json:"masterNodeSecurityGroupId"`
	BaseNodeAMIID             string `json:"baseNodeAmiId"`
	KeyName                   string `json:"keyName"`
}

type ConfigSystem struct {
	DefinitionURL string `json:"url"`
}

type ConfigComponentBuild struct {
	Builder        ConfigComponentBuildBuilder        `json:"builderConfig"`
	DockerArtifact ConfigComponentBuildDockerArtifact `json:"dockerConfig"`
}

type ConfigComponentBuildBuilder struct {
	Image string `json:"image"`

	// Version of the docker API used by the build node docker daemons
	DockerAPIVersion string `json:"dockerApiVersion"`
}

type ConfigComponentBuildDockerArtifact struct {
	// Registry used to tag images.
	Registry string `json:"registry"`

	// If true, make a new repository for the image.
	// If false, use Repository as the repository for the image and give it
	// a unique tag.
	RepositoryPerImage bool   `json:"repositoryPerImage"`
	Repository         string `json:"repository"`

	// If true push the image to the repository.
	// Set to false for the local case.
	Push bool `json:"push"`
}

type ConfigEnvoy struct {
	PrepareImage      string `json:"prepareImage"`
	Image             string `json:"image"`
	RedirectCIDRBlock string `json:"redirectCidrBlock"`
	XDSAPIPort        int32  `json:"xdsApiPort"`
}

type ConfigTerraform struct {
	Backend *ConfigTerraformBackend
}

type ConfigTerraformBackend struct {
	S3 *ConfigTerraformBackendS3 `json:"s3,omitempty"`
}

type ConfigTerraformBackendS3 struct {
	Bucket string `json:"bucket"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Config `json:"items"`
}
