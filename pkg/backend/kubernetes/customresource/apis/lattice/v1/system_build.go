package v1

import (
	"encoding/json"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularSystemBuild  = "systembuild"
	ResourcePluralSystemBuild    = "systembuilds"
	ResourceShortNameSystemBuild = "lsysb"
	ResourceScopeSystemBuild     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemBuildSpec   `json:"spec"`
	Status            SystemBuildStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=false
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

// +k8s:deepcopy-gen=false
type SystemBuildServicesInfo struct {
	Definition definition.Service                              `json:"definition"`
	Name       *string                                         `json:"name,omitempty"`
	Status     *ServiceBuildStatus                             `json:"status,omitempty"`
	Components map[string]SystemBuildServicesInfoComponentInfo `json:"components,omitempty"`
}

// +k8s:deepcopy-gen=false
type SystemBuildServicesInfoComponentInfo struct {
	Name   *string               `json:"name,omitempty"`
	Status *ComponentBuildStatus `json:"status,omitempty"`
}

type SystemBuildStatus struct {
	State   SystemBuildState `json:"state,omitempty"`
	Message string           `json:"message,omitempty"`
}

type SystemBuildState string

const (
	SystemBuildStatePending   SystemBuildState = "pending"
	SystemBuildStateRunning   SystemBuildState = "running"
	SystemBuildStateSucceeded SystemBuildState = "succeeded"
	SystemBuildStateFailed    SystemBuildState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemBuild `json:"items"`
}
