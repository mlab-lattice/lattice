package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
)

const ComponentTypeService = "service"

var ServiceType = component.Type{
	APIVersion: APIVersion,
	Type:       ComponentTypeService,
}

//  NOTE: if you update the Service struct, you _must_ update the
//        serviceEncoder struct as well as the serviceIR struct

type Service struct {
	Description string

	Container
	Sidecars map[string]Container

	// FIXME: remove these
	NumInstances int32
	NodePool     *NodePoolOrReference
	InstanceType *string
}

func (s *Service) Type() component.Type {
	return ServiceType
}

func (s *Service) Containers() *WorkloadContainers {
	return &WorkloadContainers{
		Main:     s.Container,
		Sidecars: s.Sidecars,
	}
}

func (s *Service) containers() []Container {
	containers := []Container{s.Container}
	for _, sidecarContainer := range s.Sidecars {
		containers = append(containers, sidecarContainer)
	}

	return containers
}

func (s *Service) ContainerPorts() map[int32]ContainerPort {
	ports := make(map[int32]ContainerPort)
	for _, container := range s.containers() {
		for portNum, port := range container.Ports {
			ports[portNum] = port
		}
	}

	return ports
}

func (s *Service) MarshalJSON() ([]byte, error) {
	e := serviceEncoder{
		Type:        ServiceType,
		Description: s.Description,

		Container: s.Container,
		Sidecars:  s.Sidecars,

		NumInstances: s.NumInstances,
		NodePool:     s.NodePool,
		InstanceType: s.InstanceType,
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

	if e.Type.Type != ComponentTypeService {
		return fmt.Errorf("expected resource type %v but got %v", ComponentTypeService, e.Type.Type)
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
	Type        component.Type `json:"type"`
	Description string         `json:"description,omitempty"`

	Container
	Sidecars map[string]Container `json:"sidecars,omitempty"`

	NumInstances int32                `json:"num_instances,omitempty"`
	NodePool     *NodePoolOrReference `json:"node_pool,omitempty"`
	InstanceType *string              `json:"instance_type,omitempty"`
}
