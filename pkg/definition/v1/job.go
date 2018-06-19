package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
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
	NodePool string `json:"node_pool"`
}

func (j *Job) Type() component.Type {
	return JobType
}

func (j *Job) Containers() []Container {
	containers := []Container{j.Container}
	for _, sidecarContainer := range j.Sidecars {
		containers = append(containers, sidecarContainer)
	}

	return containers
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

	NodePool string `json:"node_pool"`
}
