package resource

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Interface interface {
	Type() Type
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

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return fmt.Errorf("imporperly formatted type: %v", s)
	}

	tmp := &Type{
		APIVersion: parts[0],
		Type:       parts[1],
	}
	*t = *tmp
	return nil
}
