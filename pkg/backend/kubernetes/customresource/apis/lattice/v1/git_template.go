package v1

import (
	"fmt"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularGitTemplate = "gittemplate"
	ResourcePluralGitTemplate   = "gittemplates"
	ResourceScopeGitTemplate    = apiextensionsv1beta1.NamespaceScoped
)

var (
	GitTemplateKind     = SchemeGroupVersion.WithKind("GitTemplate")
	GitTemplateListKind = SchemeGroupVersion.WithKind("GitTemplateList")

	GitTemplateReferenceDigestLabelKey = fmt.Sprintf("gittemplate.%v/reference-digest", GroupName)
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
