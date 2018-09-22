package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Template *template.Template `json:"template"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Template `json:"items"`
}
