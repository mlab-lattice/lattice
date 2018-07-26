package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/template"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularTemplate = "template"
	ResourcePluralTemplate   = "templates"
	ResourceScopeTemplate    = apiextensionsv1beta1.NamespaceScoped
)

var (
	TemplateKind     = SchemeGroupVersion.WithKind("Template")
	TemplateListKind = SchemeGroupVersion.WithKind("TemplateList")
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TemplateSpec `json:"spec"`
}

// +k8s:deepcopy-gen=false

type TemplateSpec struct {
	*template.Template
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Template `json:"items"`
}
