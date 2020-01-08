package v1

import (
	"encoding/json"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type SecretRef struct {
	Value tree.PathSubcomponent `json:"$secret_ref"`
}

// ValueOrSecret contains either a value (i.e. just a string value), or a Secret.
type ValueOrSecret struct {
	// Value
	Value *string
	// Secret Reference
	SecretRef *SecretRef
}

func (v ValueOrSecret) MarshalJSON() ([]byte, error) {
	if v.Value != nil {
		return json.Marshal(v.Value)
	}

	if v.SecretRef == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(v.SecretRef)
}

func (v *ValueOrSecret) UnmarshalJSON(data []byte) error {
	var s SecretRef
	if err := json.Unmarshal(data, &s); err == nil {
		v.SecretRef = &s
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	v.Value = &str
	return nil
}
