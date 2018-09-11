package resolver

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

// ResolutionInfo contains information about the resolution of a subcomponent.
type ResolutionInfo struct {
	// Component contains the hydrated but unresolved version of the component.
	// That is, if the component is a v1/system, it may contain unresolved references.
	Component component.Interface
	Commit    git.CommitReference
	// TODO(kevindrosendahl): probably want to move this out when we have a more
	// concrete theory on component resolution secrets.
	SSHKeySecret *tree.PathSubcomponent
}

func (i *ResolutionInfo) MarshalJSON() ([]byte, error) {
	componentData, err := json.Marshal(i.Component)
	if err != nil {
		return nil, err
	}

	d := &resolutionInfoDecoder{
		Component:    componentData,
		Commit:       i.Commit,
		SSHKeySecret: i.SSHKeySecret,
	}
	return json.Marshal(&d)
}

func (i *ResolutionInfo) UnmarshalJSON(data []byte) error {
	var d resolutionInfoDecoder
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	t, err := component.TypeFromJSON(d.Component)
	if err != nil {
		return err
	}

	var c component.Interface
	switch t.APIVersion {
	case definitionv1.APIVersion:
		c, err = definitionv1.NewComponentFromJSON(d.Component)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid type api %v", t.APIVersion)
	}

	(*i).Component = c
	(*i).Commit = d.Commit
	(*i).SSHKeySecret = d.SSHKeySecret

	return nil
}

type resolutionInfoDecoder struct {
	Component    json.RawMessage        `json:"component"`
	Commit       git.CommitReference    `json:"commit"`
	SSHKeySecret *tree.PathSubcomponent `json:"sshKeySecret"`
}

type resolutionContext struct {
	FileReference *git.FileReference
	SSHKeySecret  *tree.PathSubcomponent
	SSHKey        []byte
}
