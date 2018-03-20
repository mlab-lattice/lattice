package v1

import (
	"encoding/json"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularBuild  = "build"
	ResourcePluralBuild    = "builds"
	ResourceShortNameBuild = "lbld"
	ResourceScopeBuild     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Build struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BuildSpec   `json:"spec"`
	Status            BuildStatus `json:"status,omitempty"`
}

// N.B.: important: if you update the BuildSpec or BuildSpecServiceInfo you must also update
// the buildSpecEncoder and BuildSpec's UnmarshalJSON
// +k8s:deepcopy-gen=false
type BuildSpec struct {
	DefinitionRoot tree.Node                              `json:"definitionRoot"`
	Services       map[tree.NodePath]BuildSpecServiceInfo `json:"services"`
}

// +k8s:deepcopy-gen=false
type BuildSpecServiceInfo struct {
	Definition definition.Service `json:"definition"`
}

type buildSpecEncoder struct {
	Services       map[tree.NodePath]buildSpecServiceInfoEncoder `json:"services"`
	DefinitionRoot json.RawMessage                               `json:"definitionRoot"`
}

type buildSpecServiceInfoEncoder struct {
	Definition json.RawMessage `json:"definition"`
}

func (sbs *BuildSpec) UnmarshalJSON(data []byte) error {
	var decoded buildSpecEncoder
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

	services := map[tree.NodePath]BuildSpecServiceInfo{}
	for path, serviceInfo := range decoded.Services {
		service, err := definition.NewServiceFromJSON(serviceInfo.Definition)
		if err != nil {
			return err
		}

		services[path] = BuildSpecServiceInfo{
			Definition: service,
		}
	}

	*sbs = BuildSpec{
		DefinitionRoot: rootNode,
		Services:       services,
	}
	return nil
}

type BuildStatus struct {
	State              BuildState `json:"state"`
	ObservedGeneration int64      `json:"observedGeneration"`
	Message            string     `json:"message"`

	// Maps a service path to the ServiceBuild.Name responsible for it
	ServiceBuilds map[tree.NodePath]string `json:"serviceBuilds"`

	// Maps a ServiceBuild.Name to the ServiceBuild.Status
	ServiceBuildStatuses map[string]ServiceBuildStatus `json:"serviceBuildStatuses"`
}

type BuildState string

const (
	BuildStatePending   BuildState = "pending"
	BuildStateRunning   BuildState = "running"
	BuildStateSucceeded BuildState = "succeeded"
	BuildStateFailed    BuildState = "failed"
)

type BuildStatusServiceInfo struct {
	Name       string                                         `json:"name"`
	Status     ServiceBuildStatus                             `json:"status"`
	Components map[string]BuildStatusServiceInfoComponentInfo `json:"components"`
}

type BuildStatusServiceInfoComponentInfo struct {
	Name   string               `json:"name"`
	Status ComponentBuildStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Build `json:"items"`
}
