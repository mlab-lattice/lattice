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

var (
	ConfigKind     = SchemeGroupVersion.WithKind("Config")
	ConfigListKind = SchemeGroupVersion.WithKind("ConfigList")
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
	ContainerBuild ConfigContainerBuild `json:"containerBuild"`
	ServiceMesh    ConfigServiceMesh    `json:"serviceMesh"`
}

type ConfigCloudProvider struct {
	Local *ConfigCloudProviderLocal `json:"local,omitempty"`
	AWS   *ConfigCloudProviderAWS   `json:"aws,omitempty"`
}

type ConfigCloudProviderLocal struct {
}

type ConfigCloudProviderAWS struct {
	WorkerNodeAMIID string `json:"workerNodeAmiId"`
	KeyName         string `json:"keyName"`
}

type ConfigContainerBuild struct {
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

	// If auth is required, specify the auth type
	RegistryAuthType *string `json:"registryAuthType"`

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
	Envoy *ConfigServiceMeshEnvoy `json:"envoy"`
}

type ConfigServiceMeshEnvoy struct {
	PrepareImage string `json:"prepareImage"`
	Image        string `json:"image"`
	XDSAPIImage  string `json:"xdsApiImage"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Config `json:"items"`
}
