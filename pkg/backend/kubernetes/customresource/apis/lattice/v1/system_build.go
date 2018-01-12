package v1

import (
	"encoding/json"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"

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

// N.B.: important: if you update the SystemBuildSpec or SystemBuildSpecServiceInfo you must also update
// the systemBuildSpecEncoder and SystemBuildSpec's UnmarshalJSON
// +k8s:deepcopy-gen=false
type SystemBuildSpec struct {
	DefinitionRoot tree.Node                                    `json:"definitionRoot"`
	Services       map[tree.NodePath]SystemBuildSpecServiceInfo `json:"services"`
}

// +k8s:deepcopy-gen=false
type SystemBuildSpecServiceInfo struct {
	Definition definition.Service `json:"definition"`
}

type systemBuildSpecEncoder struct {
	Services       map[tree.NodePath]systemBuildSpecServiceInfoEncoder `json:"services"`
	DefinitionRoot json.RawMessage                                     `json:"definitionRoot"`
}

type systemBuildSpecServiceInfoEncoder struct {
	Definition json.RawMessage `json:"definition"`
}

func (sbs *SystemBuildSpec) UnmarshalJSON(data []byte) error {
	var decoded systemBuildSpecEncoder
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	def, err := definition.NewFromJSON(decoded.DefinitionRoot)
	if err != nil {
		return err
	}

	rootNode, err := tree.NewNode(def, nil)
	if err != nil {
		return err
	}

	services := map[tree.NodePath]SystemBuildSpecServiceInfo{}
	for path, serviceInfo := range decoded.Services {
		service, err := definition.NewServiceFromJSON(serviceInfo.Definition)
		if err != nil {
			return err
		}

		services[path] = SystemBuildSpecServiceInfo{
			Definition: service,
		}
	}

	*sbs = SystemBuildSpec{
		DefinitionRoot: rootNode,
		Services:       services,
	}
	return nil
}

type SystemBuildStatus struct {
	State              SystemBuildState `json:"state"`
	ObservedGeneration int64            `json:"observedGeneration"`
	Message            string           `json:"message"`

	// Maps a service path to the ServiceBuild.Name responsible for it
	ServiceBuilds map[tree.NodePath]string `json:"serviceBuilds"`

	// Maps a ServiceBuild.Name to the ServiceBuild.Status
	ServiceBuildStatuses map[string]ServiceBuildStatus `json:"serviceBuildStatuses"`
}

type SystemBuildState string

const (
	SystemBuildStatePending   SystemBuildState = "pending"
	SystemBuildStateRunning   SystemBuildState = "running"
	SystemBuildStateSucceeded SystemBuildState = "succeeded"
	SystemBuildStateFailed    SystemBuildState = "failed"
)

type SystemBuildStatusServiceInfo struct {
	Name       string                                               `json:"name"`
	Status     ServiceBuildStatus                                   `json:"status"`
	Components map[string]SystemBuildStatusServiceInfoComponentInfo `json:"components"`
}

type SystemBuildStatusServiceInfoComponentInfo struct {
	Name   string               `json:"name"`
	Status ComponentBuildStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemBuild `json:"items"`
}
