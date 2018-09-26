package component

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	TypeField = "type"
)

type Interface interface {
	Type() Type

	// see: https://github.com/kubernetes/gengo/blob/master/examples/deepcopy-gen/main.go#L24-L31
	DeepCopyInterface() Interface
}

func TypeFromMap(m map[string]interface{}) (Type, error) {
	i, ok := m[TypeField]
	if !ok {
		return Type{}, fmt.Errorf("component must contain type field")
	}

	s, ok := i.(string)
	if !ok {
		return Type{}, fmt.Errorf("type field must be a string")
	}

	return TypeFromString(s)
}

func TypeFromString(s string) (Type, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return Type{}, fmt.Errorf("imporperly formatted type: %v", s)
	}

	t := Type{
		APIVersion: parts[0],
		Type:       parts[1],
	}
	return t, nil
}

func TypeFromJSON(data []byte) (Type, error) {
	t := typeDecoder{}
	err := json.Unmarshal(data, &t)
	if err != nil {
		return Type{}, err
	}

	return t.Type, err
}

type typeDecoder struct {
	Type Type `json:"type"`
}

type Type struct {
	APIVersion string
	Type       string
}

func (t Type) String() string {
	return fmt.Sprintf("%v/%v", t.APIVersion, t.Type)
}

func (t *Type) MarshalJSON() ([]byte, error) {
	s := t.String()
	return json.Marshal(&s)
}

func (t *Type) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	tmp, err := TypeFromString(s)
	if err != nil {
		return err
	}

	*t = tmp
	return nil
}
