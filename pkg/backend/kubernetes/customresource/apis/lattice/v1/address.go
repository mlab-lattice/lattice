package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularServiceAddress = "address"
	ResourcePluralServiceAddress   = "addresses"
	ResourceScopeServiceAddress    = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Address struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AddressSpec   `json:"spec"`
	Status            AddressStatus `json:"status"`
}

func (a *Address) UpdateProcessed() bool {
	return a.Status.ObservedGeneration >= a.Generation
}

type AddressSpec struct {
	Path         tree.NodePath  `json:"path"`
	Service      *tree.NodePath `json:"service,omitempty"`
	ExternalName *string        `json:"externalName,omitempty"`
}

type AddressStatus struct {
	State              AddressState `json:"state"`
	ObservedGeneration int64        `json:"observedGeneration"`

	// Public maps ports to their publicly accessible address
	Public map[int32]string
}

type AddressState string

const (
	ServiceAddressStatePending AddressState = "pending"
	ServiceAddressStateCreated AddressState = "created"
	ServiceAddressStateFailed  AddressState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AddressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Address `json:"items"`
}
