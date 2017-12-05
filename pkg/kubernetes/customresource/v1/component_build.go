package v1

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceSingularComponentBuild  = "componentbuild"
	ResourcePluralComponentBuild    = "componentbuilds"
	ResourceShortNameComponentBuild = "lcb"
	ResourceScopeComponentBuild     = apiextensionsv1beta1.NamespaceScoped
)

type ComponentBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ComponentBuildSpec   `json:"spec"`
	Status            ComponentBuildStatus `json:"status"`
}

type ComponentBuildSpec struct {
	BuildDefinitionBlock block.ComponentBuild     `json:"definitionBlock"`
	Artifacts            *ComponentBuildArtifacts `json:"artifacts,omitempty"`
}

type ComponentBuildArtifacts struct {
	DockerImageFqn string `json:"dockerImageFqn"`
}

type ComponentBuildStatus struct {
	State             ComponentBuildState        `json:"state"`
	LastObservedPhase *types.ComponentBuildPhase `json:"lastObservedPhase,omitempty"`
	FailureInfo       *ComponentBuildFailureInfo `json:"failureInfo,omitempty"`
}

type ComponentBuildState string

const (
	ComponentBuildStatePending   ComponentBuildState = "Pending"
	ComponentBuildStateQueued    ComponentBuildState = "Queued"
	ComponentBuildStateRunning   ComponentBuildState = "Running"
	ComponentBuildStateSucceeded ComponentBuildState = "Succeeded"
	ComponentBuildStateFailed    ComponentBuildState = "Failed"
)

type ComponentBuildFailureInfo struct {
	Message  string `json:"message"`
	Internal bool   `json:"internal"`
}

type ComponentBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ComponentBuild `json:"items"`
}

// Below is taken from: https://github.com/kubernetes/apiextensions-apiserver/blob/master/examples/client-go/apis/cr/v1/zz_generated.deepcopy.go
// It's needed because runtime.Scheme.AddKnownTypes requires the type to implement runtime.interfaces.Object,
// which includes DeepCopyObject
// TODO: figure out how to autogen this

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentBuild) DeepCopyInto(out *ComponentBuild) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Example.
func (in *ComponentBuild) DeepCopy() *ComponentBuild {
	if in == nil {
		return nil
	}
	out := new(ComponentBuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ComponentBuild) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentBuildList) DeepCopyInto(out *ComponentBuildList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ComponentBuild, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleList.
func (in *ComponentBuildList) DeepCopy() *ComponentBuildList {
	if in == nil {
		return nil
	}
	out := new(ComponentBuildList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ComponentBuildList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentBuildSpec) DeepCopyInto(out *ComponentBuildSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleSpec.
func (in *ComponentBuildSpec) DeepCopy() *ComponentBuildSpec {
	if in == nil {
		return nil
	}
	out := new(ComponentBuildSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentBuildStatus) DeepCopyInto(out *ComponentBuildStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleStatus.
func (in *ComponentBuildStatus) DeepCopy() *ComponentBuildStatus {
	if in == nil {
		return nil
	}
	out := new(ComponentBuildStatus)
	in.DeepCopyInto(out)
	return out
}
