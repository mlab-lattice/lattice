package v1

import (
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceSingularSystemTeardown  = "systemteardown"
	ResourcePluralSystemTeardown    = "systemteardowns"
	ResourceShortNameSystemTeardown = "lsyst"
	ResourceScopeSystemTeardown     = apiextensionsv1beta1.NamespaceScoped
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SystemTeardown struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemTeardownSpec   `json:"spec"`
	Status            SystemTeardownStatus `json:"status,omitempty"`
}

type SystemTeardownSpec struct {
	LatticeNamespace types.LatticeNamespace
}

type SystemTeardownStatus struct {
	State   SystemTeardownState `json:"state,omitempty"`
	Message string              `json:"message,omitempty"`
}

type SystemTeardownState string

const (
	SystemTeardownStatePending    SystemTeardownState = "Pending"
	SystemTeardownStateInProgress SystemTeardownState = "InProgress"
	SystemTeardownStateSucceeded  SystemTeardownState = "Succeeded"
	SystemTeardownStateFailed     SystemTeardownState = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SystemTeardownList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemTeardown `json:"items"`
}

// Below is taken from: https://github.com/kubernetes/apiextensions-apiserver/blob/master/examples/client-go/apis/cr/v1/zz_generated.deepcopy.go
// It's needed because runtime.Scheme.AddKnownTypes requires the type to implement runtime.interfaces.Object,
// which includes DeepCopyObject
// TODO: figure out how to autogen this

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemTeardown) DeepCopyInto(out *SystemTeardown) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Example.
func (in *SystemTeardown) DeepCopy() *SystemTeardown {
	if in == nil {
		return nil
	}
	out := new(SystemTeardown)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SystemTeardown) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemTeardownList) DeepCopyInto(out *SystemTeardownList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SystemTeardown, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleList.
func (in *SystemTeardownList) DeepCopy() *SystemTeardownList {
	if in == nil {
		return nil
	}
	out := new(SystemTeardownList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SystemTeardownList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemTeardownSpec) DeepCopyInto(out *SystemTeardownSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleSpec.
func (in *SystemTeardownSpec) DeepCopy() *SystemTeardownSpec {
	if in == nil {
		return nil
	}
	out := new(SystemTeardownSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemTeardownStatus) DeepCopyInto(out *SystemTeardownStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleStatus.
func (in *SystemTeardownStatus) DeepCopy() *SystemTeardownStatus {
	if in == nil {
		return nil
	}
	out := new(SystemTeardownStatus)
	in.DeepCopyInto(out)
	return out
}
