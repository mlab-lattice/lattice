package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/resource"
)

const ResourceTypeService = "service"

var ServiceType = resource.Type{
	APIVersion: APIVersion,
	Type:       ResourceTypeService,
}

type Service struct {
	Description string

	Container
	Sidecars map[string]Container

	// FIXME: remove these
	NumInstances int32
	NodePool     *NodePoolOrReference
	InstanceType *string
}

func (s *Service) Type() resource.Type {
	return ServiceType
}

func (s *Service) MarshalJSON() ([]byte, error) {
	e := serviceEncoder{
		Type:        ServiceType,
		Description: s.Description,

		Container: s.Container,
		Sidecars:  s.Sidecars,

		NumInstances: s.NumInstances,
	}
	return json.Marshal(&e)
}

func (s *Service) UnmarshalJSON(data []byte) error {
	var e *serviceEncoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	if e.Type.APIVersion != APIVersion {
		return fmt.Errorf("expected api version %v but got %v", APIVersion, e.Type.APIVersion)
	}

	if e.Type.Type != ResourceTypeService {
		return fmt.Errorf("expected resource type %v but got %v", ResourceTypeService, e.Type.Type)
	}

	service := &Service{
		Description: e.Description,

		Container: e.Container,
		Sidecars:  e.Sidecars,

		NumInstances: e.NumInstances,
		NodePool:     e.NodePool,
		InstanceType: e.InstanceType,
	}
	*s = *service
	return nil
}

type serviceEncoder struct {
	Type        resource.Type `json:"type"`
	Description string        `json:"description,omitempty"`

	Container
	Sidecars map[string]Container `json:"sidecars,omitempty"`

	NumInstances int32                `json:"num_instances"`
	NodePool     *NodePoolOrReference `json:"node_pool"`
	InstanceType *string              `json:"instance_type"`
}
