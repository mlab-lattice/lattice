package v1

import (
	"encoding/json"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type SecretRef struct {
	Value tree.PathSubcomponent `json:"secret_ref"`
}

// Secret can either be a local named reference (i.e. just a string name without
// a path) or a PathSubcomponent.
// N.B.: in the future a secret may just be a local reference depending on how
//       we decide to deal with secret management
type Secret struct {
	Path  *tree.PathSubcomponent
	Local *string
}

func (s Secret) MarshalJSON() ([]byte, error) {
	if s.Path != nil {
		return json.Marshal(s.Path)
	}

	if s.Local != nil {
		return json.Marshal(s.Local)
	}

	return json.Marshal(nil)
}

func (s *Secret) UnmarshalJSON(data []byte) error {
	var path tree.PathSubcomponent
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
	Value  *string
	Secret *Secret
}

func (v ValueOrSecret) MarshalJSON() ([]byte, error) {
	if v.Value != nil {
		return json.Marshal(v.Value)
	}

	if v.Secret == nil {
		return json.Marshal(nil)
	}
	s := valueOrSecretSecretDecoder{*v.Secret}
	return json.Marshal(&s)
}

func (v *ValueOrSecret) UnmarshalJSON(data []byte) error {
	var s valueOrSecretSecretDecoder
	if err := json.Unmarshal(data, &s); err == nil {
		v.Secret = &s.Secret
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	v.Value = &str
	return nil
}

type valueOrSecretSecretDecoder struct {
	Secret Secret `json:"$secret"`
}
