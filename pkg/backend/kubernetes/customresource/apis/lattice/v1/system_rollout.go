package v1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularSystemRollout  = "systemrollout"
	ResourcePluralSystemRollout    = "systemrollouts"
	ResourceShortNameSystemRollout = "lsysr"
	ResourceScopeSystemRollout     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemRollout struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemRolloutSpec   `json:"spec"`
	Status            SystemRolloutStatus `json:"status"`
}

// +k8s:deepcopy-gen=false
type SystemRolloutSpec struct {
	BuildName string `json:"buildName"`
}

type SystemRolloutStatus struct {
	State   SystemRolloutState `json:"state"`
	Message string             `json:"message"`
}

type SystemRolloutState string

const (
	SystemRolloutStatePending    SystemRolloutState = "pending"
	SystemRolloutStateAccepted   SystemRolloutState = "accepted"
	SystemRolloutStateInProgress SystemRolloutState = "in progress"
	SystemRolloutStateSucceeded  SystemRolloutState = "succeeded"
	SystemRolloutStateFailed     SystemRolloutState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemRolloutList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemRollout `json:"items"`
}
