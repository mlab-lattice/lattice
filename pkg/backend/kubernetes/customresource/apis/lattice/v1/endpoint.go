package v1

import (
	"github.com/mlab-lattice/system/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularEndpoint  = "endpoint"
	ResourcePluralEndpoint    = "endpoints"
	ResourceShortNameEndpoint = "lep"
	ResourceScopeEndpoint     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Endpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              EndpointSpec   `json:"spec"`
	Status            EndpointStatus `json:"status"`
}

type EndpointSpec struct {
	Path             tree.NodePath `json:"path"`
	ExternalEndpoint *string       `json:"externalEndpoint,omitempty"`
	IP               *string       `json:"ip,omitempty"`
}

type EndpointStatus struct {
	State EndpointState `json:"state"`
}

type EndpointState string

const (
	EndpointStatePending   EndpointState = "pending"
	EndpointStateSucceeded EndpointState = "created"
	EndpointStateFailed    EndpointState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Endpoint `json:"items"`
}
