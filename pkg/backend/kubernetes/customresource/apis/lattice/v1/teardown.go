package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularTeardown = "teardown"
	ResourcePluralTeardown   = "teardowns"
	ResourceScopeTeardown    = apiextensionsv1beta1.NamespaceScoped
)

var (
	TeardownKind     = SchemeGroupVersion.WithKind("Teardown")
	TeardownListKind = SchemeGroupVersion.WithKind("TeardownList")
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Teardown struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TeardownSpec   `json:"spec"`
	Status            TeardownStatus `json:"status"`
}

func (t *Teardown) V1ID() v1.TeardownID {
	return v1.TeardownID(t.Name)
}

func (t *Teardown) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, t.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", t.Namespace))
	}

	return fmt.Sprintf("teardown %v (system %v)", t.Name, systemID)
}

type TeardownSpec struct {
}

type TeardownStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`

	State   TeardownState `json:"state"`
	Message string        `json:"message"`

	StartTimestamp      *metav1.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`
}

type TeardownState string

const (
	TeardownStatePending    TeardownState = ""
	TeardownStateInProgress TeardownState = "in progress"
	TeardownStateSucceeded  TeardownState = "succeeded"
	TeardownStateFailed     TeardownState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TeardownList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Teardown `json:"items"`
}
