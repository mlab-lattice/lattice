package v1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularDeploy = "deploy"
	// github.com/kubernetes/code-generator doesn't allow you to specify
	// the plural name of a struct it's generating clients for, and "deploys"
	// doesn't follow the plural rules here: https://github.com/kubernetes/gengo/blob/master/namer/plural_namer.go#L68-L74
	// need to find a way to customize this name, until then, use the misspelling
	ResourcePluralDeploy = "deploies"
	//ResourcePluralDeploy    = "deploys"
	ResourceShortNameDeploy = "ldply"
	ResourceScopeDeploy     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Deploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DeploySpec   `json:"spec"`
	Status            DeployStatus `json:"status"`
}

type DeploySpec struct {
	BuildName string `json:"buildName"`
}

type DeployStatus struct {
	State              DeployState `json:"state"`
	ObservedGeneration int64       `json:"observedGeneration"`
	Message            string      `json:"message"`
}

type DeployState string

const (
	DeployStatePending    DeployState = "pending"
	DeployStateAccepted   DeployState = "accepted"
	DeployStateInProgress DeployState = "in progress"
	DeployStateSucceeded  DeployState = "succeeded"
	DeployStateFailed     DeployState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DeployList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Deploy `json:"items"`
}
