package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularAddress = "address"
	ResourcePluralAddress   = "addresses"
	ResourceScopeAddress    = apiextensionsv1beta1.NamespaceScoped
)

var (
	AddressKind         = SchemeGroupVersion.WithKind("Address")
	AddressPathLabelKey = fmt.Sprintf("address.%v/path", GroupName)
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

func (a *Address) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, a.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", a.Namespace))
	}

	path, err := a.PathLabel()
	if err == nil {
		return fmt.Sprintf("address %v (%v in system %v)", a.Name, path, systemID)
	}

	return fmt.Sprintf("address %v (no path, system %v)", a.Name, systemID)
}

func (a *Address) PathLabel() (tree.NodePath, error) {
	path, ok := a.Labels[AddressPathLabelKey]
	if !ok {
		return "", fmt.Errorf("service did not contain service path label")
	}

	return tree.NodePathFromDomain(path)
}

type AddressSpec struct {
	Service      *tree.NodePath `json:"service,omitempty"`
	ExternalName *string        `json:"externalName,omitempty"`
}

type AddressStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`

	State       AddressState              `json:"state"`
	FailureInfo *AddressStatusFailureInfo `json:"failureInfo"`

	// Ports maps ports to their publicly accessible address
	Ports map[int32]string
}

type AddressState string

const (
	AddressStatePending AddressState = "pending"
	AddressStateStable  AddressState = "created"
	AddressStateFailed  AddressState = "failed"
)

type AddressStatusFailureInfo struct {
	Message string      `json:"message"`
	Time    metav1.Time `json:"time"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AddressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Address `json:"items"`
}
