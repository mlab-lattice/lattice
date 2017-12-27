package v1

import (
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
	CloudProvider  ConfigCloudProvider  `json:"cloudProvider"`
	ComponentBuild ConfigComponentBuild `json:"componentBuild"`
	ServiceMesh    ConfigServiceMesh    `json:"serviceMesh"`
	Terraform      *ConfigTerraform     `json:"terraform,omitempty"`
}

type ConfigCloudProvider struct {
	Local *ConfigCloudProviderLocal `json:"local,omitempty"`
	AWS   *ConfigCloudProviderAWS   `json:"aws,omitempty"`
}

type ConfigCloudProviderLocal struct {
	// FIXME: this shouldn't be dynamic config
	IP                 string   `json:"ip"`
	DNSControllerIamge string   `json:"controller-image"`
	DNSServerImage     string   `json:"server-image"`
	DNSServerArgs      []string `json:"server-args"`
	DNSControllerArgs  []string `json:"controller-args"`
}

type ConfigCloudProviderAWS struct {
	// FIXME: this shouldn't be dynamic config
	Region string `json:"region"`
	// FIXME: this shouldn't be dynamic config
	AccountID string `json:"accountID"`
	// FIXME: this shouldn't be dynamic config
	VPCID string `json:"vpcId"`
	// FIXME: maybe this shouldn't be dynamic config
	SubnetIDs []string `json:"subnetIds"`
	// FIXME: this shouldn't be dynamic config
	MasterNodeSecurityGroupID string `json:"masterNodeSecurityGroupId"`
	BaseNodeAMIID             string `json:"baseNodeAmiId"`
	KeyName                   string `json:"keyName"`
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

type ConfigServiceMesh struct {
	Envoy *ConfigEnvoy `json:"envoy"`
}

type ConfigEnvoy struct {
	PrepareImage      string `json:"prepareImage"`
	Image             string `json:"image"`
	RedirectCIDRBlock string `json:"redirectCidrBlock"`
	XDSAPIImage       string `json:"xdsApiImage"`
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
