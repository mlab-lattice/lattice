package v1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularTeardown  = "teardown"
	ResourcePluralTeardown    = "teardowns"
	ResourceShortNameTeardown = "ltdwn"
	ResourceScopeTeardown     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Teardown struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TeardownSpec   `json:"spec"`
	Status            TeardownStatus `json:"status"`
}

type TeardownSpec struct {
}

type TeardownStatus struct {
	State              TeardownState `json:"state"`
	ObservedGeneration int64         `json:"observedGeneration"`
	Message            string        `json:"message"`
}

type TeardownState string

const (
	TeardownStatePending    TeardownState = "pending"
	TeardownStateInProgress TeardownState = "in progress"
	TeardownStateSucceeded  TeardownState = "succeeded"
	TeardownStateFailed     TeardownState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TeardownList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Teardown `json:"items"`
}
