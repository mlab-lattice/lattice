package v1

import (
	coretypes "github.com/mlab-lattice/core/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ConfigResourceSingular = "config"
	ConfigResourcePlural   = "configs"
	// TODO: should this be cluster scoped?
	ConfigResourceScope = apiextensionsv1beta1.NamespaceScoped
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ConfigSpec `json:"spec"`
}

type ConfigSpec struct {
	SystemId       string                                      `json:"systemId"`
	Provider       ConfigProvider                              `json:"providerConfig"`
	ComponentBuild ConfigComponentBuild                        `json:"componentBuild"`
	Envoy          ConfigEnvoy                                 `json:"envoy"`
	SystemConfigs  map[coretypes.LatticeNamespace]ConfigSystem `json:"userSystem"`
	Terraform      *ConfigTerraform                            `json:"terraform,omitempty"`
}

type ConfigProvider struct {
	Local *ConfigProviderLocal `json:"local,omitempty"`
	AWS   *ConfigProviderAWS   `json:"aws,omitempty"`
}

type ConfigProviderLocal struct {
	IP string `json:"ip"`
}

type ConfigProviderAWS struct {
	Region                    string   `json:"region"`
	AccountId                 string   `json:"accountId"`
	VPCId                     string   `json:"vpcId"`
	SubnetIds                 []string `json:"subnetIds"`
	MasterNodeSecurityGroupID string   `json:"masterNodeSecurityGroupId"`
	BaseNodeAMIId             string   `json:"baseNodeAmiId"`
	KeyName                   string   `json:"keyName"`
}

type ConfigSystem struct {
	Url string `json:"url"`
}

type ConfigComponentBuild struct {
	DockerConfig     ConfigBuildDocker `json:"dockerConfig"`
	BuildImage       string            `json:"buildImage"`
}

type ConfigBuildDocker struct {
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
	RedirectCidrBlock string `json:"redirectCidrBlock"`
	XdsApiPort        int32  `json:"xdsApiPort"`
}

type ConfigTerraform struct {
	S3Backend *ConfigTerraformBackendS3 `json:"s3Backend,omitempty"`
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

// Below is taken from: https://github.com/kubernetes/apiextensions-apiserver/blob/master/examples/client-go/apis/cr/v1/zz_generated.deepcopy.go
// It's needed because runtime.Scheme.AddKnownTypes requires the type to implement runtime.interfaces.Object,
// which includes DeepCopyObject
// TODO: figure out how to autogen this

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Config) DeepCopyInto(out *Config) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Example.
func (in *Config) DeepCopy() *Config {
	if in == nil {
		return nil
	}
	out := new(Config)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Config) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigList) DeepCopyInto(out *ConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Config, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleList.
func (in *ConfigList) DeepCopy() *ConfigList {
	if in == nil {
		return nil
	}
	out := new(ConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigSpec) DeepCopyInto(out *ConfigSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleSpec.
func (in *ConfigSpec) DeepCopy() *ConfigSpec {
	if in == nil {
		return nil
	}
	out := new(ConfigSpec)
	in.DeepCopyInto(out)
	return out
}
