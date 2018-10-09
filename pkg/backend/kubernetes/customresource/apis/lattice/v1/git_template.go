package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	GitTemplateKind     = SchemeGroupVersion.WithKind("GitTemplate")
	GitTemplateListKind = SchemeGroupVersion.WithKind("GitTemplateList")

	GitTemplateRepoURLLabelKey = fmt.Sprintf("gittemplate.%v/repo-url", GroupName)
	GitTemplateCommitLabelKey  = fmt.Sprintf("gittemplate.%v/commit", GroupName)
	GitTemplateFileLabelKey    = fmt.Sprintf("gittemplate.%v/file", GroupName)
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GitTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              GitTemplateSpec `json:"spec"`
}

type GitTemplateSpec struct {
	TemplateDigest string `json:"templateDigest"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GitTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []GitTemplate `json:"items"`
}
