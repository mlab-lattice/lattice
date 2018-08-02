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
	AddressKind     = SchemeGroupVersion.WithKind("Address")
	AddressListKind = SchemeGroupVersion.WithKind("AddressList")

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

func (a *Address) PathLabel() (tree.Path, error) {
	path, ok := a.Labels[AddressPathLabelKey]
	if !ok {
		return "", fmt.Errorf("service did not contain service path label")
	}

	return tree.NewPathFromDomain(path)
}

func (a *Address) Stable() bool {
	return a.UpdateProcessed() && a.Status.State == AddressStateStable
}

func (a *Address) Failed() bool {
	return a.UpdateProcessed() && a.Status.State == AddressStateFailed
}

func (a *Address) UpdateProcessed() bool {
	return a.Status.ObservedGeneration >= a.Generation
}

func (a *Address) Reason() string {
	if !a.UpdateProcessed() {
		return "waiting for update to be processed"
	}

	switch a.Status.State {
	case AddressStateStable:
		return ""
	case AddressStateUpdating:
		return "updating"
	case AddressStateFailed:
		failureReason := "unknown reason"
		if a.Status.FailureInfo != nil {
			failureReason = fmt.Sprintf("%v at %v", a.Status.FailureInfo.Message, a.Status.FailureInfo.Time.String())
		}

		return fmt.Sprintf("failed: %v", failureReason)
	default:
		return fmt.Sprintf("in unknown state: %v", a.Status.State)
	}
}

type AddressSpec struct {
	Service      *tree.Path `json:"service,omitempty"`
	ExternalName *string    `json:"externalName,omitempty"`
}

type AddressStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`

	State       AddressState              `json:"state"`
	Message     *string                   `json:"message"`
	FailureInfo *AddressStatusFailureInfo `json:"failureInfo"`

	// Ports maps ports to their publicly accessible address
	Ports map[int32]string `json:"ports"`
}

type AddressState string

const (
	AddressStatePending  AddressState = ""
	AddressStateUpdating AddressState = "updating"
	AddressStateStable   AddressState = "stable"
	AddressStateFailed   AddressState = "failed"
	AddressStateDeleting AddressState = "deleting"
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
