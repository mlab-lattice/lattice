package v1

import (
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
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Endpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              EndpointSpec `json:"spec"`
}

type EndpointSpec struct {
	Alias            *string `json:"alias,omitempty"`
	ExternalEndpoint *string `json:"externalEndpoint,omitempty"`
	IP               *string `json:"ip,omitempty"`
	Internal         *bool   `json:"internal,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemTeardown `json:"items"`
}
