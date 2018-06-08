package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/resource"
)

const ResourceTypeJob = "job"

var JobType = resource.Type{
	APIVersion: APIVersion,
	Type:       ResourceTypeJob,
}

type Job struct {
	Description string

	Container
	Sidecars map[string]Container

	// FIXME: remove these
	NodePool string `json:"node_pool"`
}

func (j *Job) Type() resource.Type {
	return JobType
}

func (j *Job) MarshalJSON() ([]byte, error) {
	e := jobEncoder{
		Type:        JobType,
		Description: j.Description,

		Container: j.Container,
		Sidecars:  j.Sidecars,
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

	if e.Type.Type != ResourceTypeJob {
		return fmt.Errorf("expected resource type %v but got %v", ResourceTypeJob, e.Type.Type)
	}

	job := &Job{
		Description: e.Description,

		Container: e.Container,
		Sidecars:  e.Sidecars,
	}
	*j = *job
	return nil
}

type jobEncoder struct {
	Type        resource.Type `json:"type"`
	Description string        `json:"description,omitempty"`

	Container
	Sidecars map[string]Container `json:"sidecars,omitempty"`
}
