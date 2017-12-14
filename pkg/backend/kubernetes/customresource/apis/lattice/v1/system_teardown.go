package v1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularSystemTeardown  = "systemteardown"
	ResourcePluralSystemTeardown    = "systemteardowns"
	ResourceShortNameSystemTeardown = "lsyst"
	ResourceScopeSystemTeardown     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemTeardown struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemTeardownSpec   `json:"spec"`
	Status            SystemTeardownStatus `json:"status"`
}

type SystemTeardownSpec struct {
}

type SystemTeardownStatus struct {
	State              SystemTeardownState `json:"state"`
	ObservedGeneration int64               `json:"observedGeneration"`
	Message            string              `json:"message"`
}

type SystemTeardownState string

const (
	SystemTeardownStatePending    SystemTeardownState = "pending"
	SystemTeardownStateInProgress SystemTeardownState = "in progress"
	SystemTeardownStateSucceeded  SystemTeardownState = "succeeded"
	SystemTeardownStateFailed     SystemTeardownState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemTeardownList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemTeardown `json:"items"`
}
