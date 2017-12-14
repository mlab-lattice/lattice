package v1

import (
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularService  = "service"
	ResourcePluralService    = "services"
	ResourceShortNameService = "lsvc"
	ResourceScopeService     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ServiceSpec   `json:"spec"`
	Status            ServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=false
type ServiceSpec struct {
	Path       tree.NodePath      `json:"path"`
	Definition definition.Service `json:"definition"`

	// ComponentBuildArtifacts maps Component names to the artifacts created by their build
	ComponentBuildArtifacts map[string]ComponentBuildArtifacts `json:"componentBuildArtifacts"`

	// Ports maps Component names to a list of information about its ports
	Ports map[string][]ComponentPort `json:"ports"`

	// EnvoyAdminPort is the port assigned to this service to use for the Envoy admin interface
	EnvoyAdminPort int32 `json:"envoyAdminPort"`
	// EnvoyEgressPort is the port assigned to this service to use for the Envoy egress listener
	EnvoyEgressPort int32 `json:"envoyEgressPort"`

	NumInstances int32 `json:"numInstances"`
}

// +k8s:deepcopy-gen=false
type ComponentPort struct {
	Name string `json:"name"`
	Port int32  `json:"port"`
	// EnvoyPort is the port assigned to this service to use for the Envoy ingress listener for
	// this component port
	EnvoyPort int32  `json:"envoyPort"`
	Protocol  string `json:"protocol"`
	Public    bool   `json:"public"`
}

type ServiceStatus struct {
	State            ServiceState        `json:"state"`
	UpdatedInstances int32               `json:"updatedInstances"`
	StaleInstances   int32               `json:"staleInstances"`
	FailureInfo      *ServiceFailureInfo `json:"failureInfo,omitempty"`
}

type ServiceState string

const (
	ServiceStatePending     ServiceState = "pending"
	ServiceStateScalingDown ServiceState = "scaling down"
	ServiceStateScalingUp   ServiceState = "scaling up"
	ServiceStateUpdating    ServiceState = "updating"
	ServiceStateStable      ServiceState = "stable"
	ServiceStateFailed      ServiceState = "failed"
)

type ServiceFailureInfo struct {
	Message  string      `json:"message"`
	Internal bool        `json:"internal"`
	Time     metav1.Time `json:"time"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Service `json:"items"`
}
