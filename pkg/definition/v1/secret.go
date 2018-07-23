package v1

import (
	"encoding/json"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type SecretRef struct {
	Value tree.NodePathSubcomponent `json:"secret_ref"`
}

// Secret can either be a local named reference (i.e. just a string name without
// a path) or a NodePathSubcomponent.
// N.B.: in the future a secret may just be a local reference depending on how
//       we decide to deal with secret management
type Secret struct {
	Path  *tree.NodePathSubcomponent
	Local *string
}

func (s Secret) MarshalJSON() ([]byte, error) {
	if s.Path != nil {
		return json.Marshal(s.Path)
	}

	if s.Local != nil {
		json.Marshal(s.Local)
	}

	return json.Marshal(nil)
}

func (s *Secret) UnmarshalJSON(data []byte) error {
	var path tree.NodePathSubcomponent
	err := json.Unmarshal(data, &path)
	if err == nil {
		s.Path = &path
		return nil
	}

	var val string
	err = json.Unmarshal(data, &val)
	if err != nil {
		return err
	}

	s.Local = &val
	return nil
}

// ValueOrSecret contains either a value (i.e. just a string value), or a Secret.
type ValueOrSecret struct {
	Value  *string `json:"value"`
	Secret *Secret `json:"secret"`
}
