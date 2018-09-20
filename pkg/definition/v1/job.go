package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

const ComponentTypeJob = "job"

var JobType = component.Type{
	APIVersion: APIVersion,
	Type:       ComponentTypeJob,
}

type Job struct {
	Description string

	Container
	Sidecars map[string]Container

	// FIXME: remove these
	NodePool tree.PathSubcomponent `json:"node_pool"`
}

func (j *Job) Type() component.Type {
	return JobType
}

func (j *Job) Containers() *WorkloadContainers {
	return &WorkloadContainers{
		Main:     j.Container,
		Sidecars: j.Sidecars,
	}
}

func (j *Job) MarshalJSON() ([]byte, error) {
	e := jobEncoder{
		Type:        JobType,
		Description: j.Description,

		Container: j.Container,
		Sidecars:  j.Sidecars,

		NodePool: j.NodePool,
	}
	return json.Marshal(&e)
}

func (j *Job) UnmarshalJSON(data []byte) error {
	var e *jobEncoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	if e.Type.APIVersion != APIVersion {
		return fmt.Errorf("expected api version %v but got %v", APIVersion, e.Type.APIVersion)
	}

	if e.Type.Type != ComponentTypeJob {
		return fmt.Errorf("expected resource type %v but got %v", ComponentTypeJob, e.Type.Type)
	}

	job := &Job{
		Description: e.Description,

		Container: e.Container,
		Sidecars:  e.Sidecars,

		NodePool: e.NodePool,
	}
	*j = *job
	return nil
}

type jobEncoder struct {
	Type        component.Type `json:"type"`
	Description string         `json:"description,omitempty"`

	Container
	Sidecars map[string]Container `json:"sidecars,omitempty"`

	NodePool tree.PathSubcomponent `json:"node_pool"`
}
