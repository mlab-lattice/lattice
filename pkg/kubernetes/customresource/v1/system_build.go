package v1

import (
	"encoding/json"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceSingularSystemBuild  = "systembuild"
	ResourcePluralSystemBuild    = "systembuilds"
	ResourceShortNameSystemBuild = "lsysb"
	ResourceScopeSystemBuild     = apiextensionsv1beta1.NamespaceScoped
)

type SystemBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemBuildSpec   `json:"spec"`
	Status            SystemBuildStatus `json:"status,omitempty"`
}

type SystemBuildSpec struct {
	LatticeNamespace types.LatticeNamespace                    `json:"latticeNamespace"`
	DefinitionRoot   tree.Node                                 `json:"definition"`
	Services         map[tree.NodePath]SystemBuildServicesInfo `json:"services"`
}

// Some JSON (un)marshalling trickiness needed to deal with the fact that we have an interface
// type in our SystemBuildSpec (DefinitionRoot)
type systemBuildSpecRaw struct {
	LatticeNamespace types.LatticeNamespace                    `json:"latticeNamespace"`
	Services         map[tree.NodePath]SystemBuildServicesInfo `json:"services"`
	Definition       json.RawMessage
}

func (sbs *SystemBuildSpec) MarshalJSON() ([]byte, error) {
	jsonMap := map[string]interface{}{
		"latticeNamespace": sbs.LatticeNamespace,
		"definition":       sbs.DefinitionRoot.Definition(),
		"services":         sbs.Services,
	}
	return json.Marshal(jsonMap)
}

func (sbs *SystemBuildSpec) UnmarshalJSON(data []byte) error {
	var raw systemBuildSpecRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	def, err := definition.UnmarshalJSON(raw.Definition)
	if err != nil {
		return err
	}

	rootNode, err := tree.NewNode(def, nil)
	if err != nil {
		return err
	}

	*sbs = SystemBuildSpec{
		LatticeNamespace: raw.LatticeNamespace,
		DefinitionRoot:   rootNode,
		Services:         raw.Services,
	}
	return nil
}

type SystemBuildServicesInfo struct {
	Definition definition.Service                              `json:"definition"`
	BuildName  *string                                         `json:"buildName,omitempty"`
	BuildState *ServiceBuildState                              `json:"buildState"`
	Components map[string]SystemBuildServicesInfoComponentInfo `json:"components"`
}

type SystemBuildServicesInfoComponentInfo struct {
	BuildName         *string                    `json:"buildName,omitempty"`
	BuildState        *ComponentBuildState       `json:"buildState"`
	LastObservedPhase *types.ComponentBuildPhase `json:"lastObservedPhase,omitempty"`
	FailureInfo       *ComponentBuildFailureInfo `json:"failureInfo,omitempty"`
}

type SystemBuildStatus struct {
	State   SystemBuildState `json:"state,omitempty"`
	Message string           `json:"message,omitempty"`
}

type SystemBuildState string

const (
	SystemBuildStatePending   SystemBuildState = "Pending"
	SystemBuildStateRunning   SystemBuildState = "Running"
	SystemBuildStateSucceeded SystemBuildState = "Succeeded"
	SystemBuildStateFailed    SystemBuildState = "Failed"
)

type SystemBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemBuild `json:"items"`
}

// Below is taken from: https://github.com/kubernetes/apiextensions-apiserver/blob/master/examples/client-go/apis/cr/v1/zz_generated.deepcopy.go
// It's needed because runtime.Scheme.AddKnownTypes requires the type to implement runtime.interfaces.Object,
// which includes DeepCopyObject
// TODO: figure out how to autogen this

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemBuild) DeepCopyInto(out *SystemBuild) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Example.
func (in *SystemBuild) DeepCopy() *SystemBuild {
	if in == nil {
		return nil
	}
	out := new(SystemBuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SystemBuild) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemBuildList) DeepCopyInto(out *SystemBuildList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SystemBuild, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleList.
func (in *SystemBuildList) DeepCopy() *SystemBuildList {
	if in == nil {
		return nil
	}
	out := new(SystemBuildList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SystemBuildList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (sbs *SystemBuildSpec) DeepCopyInto(out *SystemBuildSpec) {
	*out = *sbs
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleSpec.
func (sbs *SystemBuildSpec) DeepCopy() *SystemBuildSpec {
	if sbs == nil {
		return nil
	}
	out := new(SystemBuildSpec)
	sbs.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemBuildStatus) DeepCopyInto(out *SystemBuildStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleStatus.
func (in *SystemBuildStatus) DeepCopy() *SystemBuildStatus {
	if in == nil {
		return nil
	}
	out := new(SystemBuildStatus)
	in.DeepCopyInto(out)
	return out
}
