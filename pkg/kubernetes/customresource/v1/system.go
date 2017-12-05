package v1

import (
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceSingularSystem  = "system"
	ResourcePluralSystem    = "systems"
	ResourceShortNameSystem = "lsys"
	ResourceScopeSystem     = apiextensionsv1beta1.NamespaceScoped
)

type System struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemSpec   `json:"spec"`
	Status            SystemStatus `json:"status,omitempty"`
}

type SystemSpec struct {
	Services map[tree.NodePath]SystemServicesInfo `json:"services"`
}

type SystemServicesInfo struct {
	Definition definition.Service `json:"definition"`

	// ComponentBuildArtifacts maps Component names to the artifacts created by their build
	ComponentBuildArtifacts map[string]ComponentBuildArtifacts `json:"componentBuildArtifacts"`

	// ServiceName is the name of the Service CustomResource that is created by the lattice-system-controller
	ServiceName *string `json:"serviceName,omitempty"`
	// ServiceState is the last observed state of the Service CustomResource
	ServiceState *ServiceState `json:"serviceState"`
}

type SystemStatus struct {
	State   SystemState `json:"state,omitempty"`
	Message string      `json:"message,omitempty"`
}

type SystemState string

const (
	SystemStateRollingOut       SystemState = "RollingOut"
	SystemStateRolloutSucceeded SystemState = "RolloutSucceeded"
	SystemStateRolloutFailed    SystemState = "RolloutFailed"
)

type SystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []System `json:"items"`
}

// Below is taken from: https://github.com/kubernetes/apiextensions-apiserver/blob/master/examples/client-go/apis/cr/v1/zz_generated.deepcopy.go
// It's needed because runtime.Scheme.AddKnownTypes requires the type to implement runtime.interfaces.Object,
// which includes DeepCopyObject
// TODO: figure out how to autogen this

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *System) DeepCopyInto(out *System) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Example.
func (in *System) DeepCopy() *System {
	if in == nil {
		return nil
	}
	out := new(System)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *System) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemList) DeepCopyInto(out *SystemList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]System, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleList.
func (in *SystemList) DeepCopy() *SystemList {
	if in == nil {
		return nil
	}
	out := new(SystemList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SystemList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemSpec) DeepCopyInto(out *SystemSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleSpec.
func (in *SystemSpec) DeepCopy() *SystemSpec {
	if in == nil {
		return nil
	}
	out := new(SystemSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemStatus) DeepCopyInto(out *SystemStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleStatus.
func (in *SystemStatus) DeepCopy() *SystemStatus {
	if in == nil {
		return nil
	}
	out := new(SystemStatus)
	in.DeepCopyInto(out)
	return out
}
