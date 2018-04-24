package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularService = "service"
	ResourcePluralService   = "services"
	ResourceScopeService    = apiextensionsv1beta1.NamespaceScoped
)

var (
	ServiceKind = SchemeGroupVersion.WithKind("Service")

	// ServiceID label is the key that should be used in a label referencing a service's ID.
	ServiceIDLabelKey = fmt.Sprintf("service.%v/id", GroupName)

	// ServiceID label is the key that should be used for the path of the service.
	ServicePathLabelKey = fmt.Sprintf("service.%v/path", GroupName)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ServiceSpec   `json:"spec"`
	Status            ServiceStatus `json:"status,omitempty"`
}

func (s *Service) Stable() bool {
	return s.UpdateProcessed() && s.Status.State == ServiceStateStable
}

func (s *Service) UpdateProcessed() bool {
	return s.Status.ObservedGeneration >= s.Generation
}

func (s *Service) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, s.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", s.Namespace))
	}

	path, err := s.PathLabel()
	if err == nil {
		return fmt.Sprintf("service %v (%v in system %v)", s.Name, path, systemID)
	}

	return fmt.Sprintf("service %v (no path, system %v)", s.Name, systemID)
}

func (s *Service) PathLabel() (tree.NodePath, error) {
	path, ok := s.Labels[ServicePathLabelKey]
	if !ok {
		return "", fmt.Errorf("service did not contain service path label")
	}

	return tree.NodePathFromDomain(path)
}

func (s *Service) NodePoolAnnotation() (NodePoolAnnotationValue, error) {
	annotation := make(NodePoolAnnotationValue)
	existingAnnotationString, ok := s.Annotations[NodePoolWorkloadAnnotationKey]
	if ok {
		err := json.Unmarshal([]byte(existingAnnotationString), &annotation)
		if err != nil {
			return nil, err
		}
	}

	return annotation, nil
}

// N.B.: important: if you update the ServiceSpec you must also update
// the serviceSpecEncoder and ServiceSpec's UnmarshalJSON
// +k8s:deepcopy-gen=false
type ServiceSpec struct {
	Definition definition.Service `json:"definition"`

	// ComponentBuildArtifacts maps Component names to the artifacts created by their build
	ComponentBuildArtifacts map[string]ComponentBuildArtifacts `json:"componentBuildArtifacts"`

	// Ports maps Component names to a list of information about its ports
	Ports map[string][]ComponentPort `json:"ports"`

	NumInstances int32 `json:"numInstances"`
}

// +k8s:deepcopy-gen=false
type ComponentPort struct {
	Name     string `json:"name"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
	Public   bool   `json:"public"`
}

type serviceSpecEncoder struct {
	Definition              json.RawMessage                    `json:"definition"`
	ComponentBuildArtifacts map[string]ComponentBuildArtifacts `json:"componentBuildArtifacts"`
	Ports                   map[string][]ComponentPort         `json:"ports"`
	NumInstances            int32                              `json:"numInstances"`
}

func (s *ServiceSpec) UnmarshalJSON(data []byte) error {
	var decoded serviceSpecEncoder
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	service, err := definition.NewServiceFromJSON(decoded.Definition)
	if err != nil {
		return err
	}

	*s = ServiceSpec{
		Definition:              service,
		ComponentBuildArtifacts: decoded.ComponentBuildArtifacts,
		Ports:        decoded.Ports,
		NumInstances: decoded.NumInstances,
	}
	return nil
}

type ServiceStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`

	State       ServiceState              `json:"state"`
	Reason      *string                   `json:"reason"`
	FailureInfo *ServiceStatusFailureInfo `json:"failureInfo,omitempty"`

	UpdatedInstances int32 `json:"updatedInstances"`
	StaleInstances   int32 `json:"staleInstances"`

	Ports map[int32]string `json:"ports"`
}

type ServiceState string

const (
	ServiceStatePending  ServiceState = "pending"
	ServiceStateScaling  ServiceState = "scaling"
	ServiceStateUpdating ServiceState = "updating"
	ServiceStateStable   ServiceState = "stable"
	ServiceStateFailed   ServiceState = "failed"
)

type ServiceStatusFailureInfo struct {
	Message  string      `json:"message"`
	Internal bool        `json:"internal"`
	Time     metav1.Time `json:"time"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Service `json:"items"`
}
