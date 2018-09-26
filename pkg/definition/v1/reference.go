package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

const ComponentTypeReference = "reference"

var ReferenceType = definition.Type{
	APIVersion: APIVersion,
	Type:       ComponentTypeReference,
}

// +k8s:deepcopy-gen:interfaces=github.com/mlab-lattice/lattice/pkg/definition.Component

type Reference struct {
	GitRepository *GitRepositoryReference
	File          *string

	Parameters ReferenceParameters
}

type ReferenceParameters map[string]interface{}

func (in *ReferenceParameters) DeepCopyInto(out *ReferenceParameters) {
	// not sure what to do about errors here
	// possible options:
	//   - do nothing (current impl, not great)
	//   - panic (not great)
	//   - have some sort of canary value in ReferenceParameter, so after you can check to see if
	//     the copy worked
	//     - not feasible for when ReferenceParameters are deeply nested
	data, _ := json.Marshal(&in)
	json.Unmarshal(data, &out)
	return
}

type GitRepositoryReference struct {
	File *string `json:"file"`
	*GitRepository
}

func (r *Reference) Type() definition.Type {
	return ReferenceType
}

func (r *Reference) MarshalJSON() ([]byte, error) {
	e := referenceEncoder{
		Type: ReferenceType,

		GitRepository: r.GitRepository,
		File:          r.File,

		Parameters: r.Parameters,
	}

	return json.Marshal(&e)
}

func (r *Reference) UnmarshalJSON(data []byte) error {
	var e referenceEncoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	// loop through the parameters and look for any that are secret references
	// and convert them into *SecretRefs
	// TODO(kevindrosendahl): consider security implications here
	for k, v := range e.Parameters {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		sv, ok := m["$secret_ref"]
		if !ok {
			continue
		}

		s, ok := sv.(string)
		if !ok {
			return fmt.Errorf("expected $secret_ref to be string")
		}

		p, err := tree.NewPathSubcomponent(s)
		if err != nil {
			return fmt.Errorf("expected $secret_ref to be path subcomponent")
		}

		sr := &SecretRef{Value: p}
		e.Parameters[k] = sr
	}

	r.GitRepository = e.GitRepository
	r.File = e.File
	r.Parameters = e.Parameters

	return nil
}

type referenceEncoder struct {
	Type definition.Type `json:"type"`

	GitRepository *GitRepositoryReference `json:"git_repository,omitempty"`
	File          *string                 `json:"file,omitempty"`

	Parameters ReferenceParameters `json:"parameters"`
}
