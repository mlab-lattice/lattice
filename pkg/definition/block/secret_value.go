package block

import (
	"encoding/json"
)

type SecretValue struct {
	Name      *string
	Reference *Reference
}

func (s *SecretValue) UnmarshalJSON(data []byte) error {
	se := &secretValueEncoder{}
	if err := json.Unmarshal(data, &se); err != nil {
		return err
	}

	if se.Secret.Name != nil {
		s.Name = se.Secret.Name
	}

	if se.Secret.Reference != nil {
		s.Reference = se.Secret.Reference
	}

	return nil
}

func (s *SecretValue) MarshalJSON() ([]byte, error) {
	se := &secretValueEncoder{
		Secret: &secretValueValueEncoder{
			Name:      s.Name,
			Reference: s.Reference,
		},
	}
	return json.Marshal(se)
}

type secretValueValueEncoder struct {
	Name      *string
	Reference *Reference
}

type secretValueEncoder struct {
	Secret *secretValueValueEncoder `json:"secret"`
}

func (sve *secretValueValueEncoder) UnmarshalJSON(data []byte) error {
	originalName := sve.Name
	// First, try to unmarshal it into Name to see if the value
	// is just a string (aka the name of a secret)
	err := json.Unmarshal(data, &sve.Name)
	if err != nil {
		// If Unmarshalling failed due to a type error, that means that
		// we were trying to unmarshal something that was not a string.
		// So we handle this error and keep going.
		if _, ok := err.(*json.UnmarshalTypeError); !ok {
			return err
		}

		// A failed Unmarshal can leave some weird data leftover, so
		// if it failed, reset sve.Name to whatever it was before
		// the attempt.
		sve.Name = originalName

		// Then, try to Unmarshal the value into the reference field.
		err = json.Unmarshal(data, &sve.Reference)
	}

	return err
}

func (sve *secretValueValueEncoder) MarshalJSON() ([]byte, error) {
	if sve.Name != nil {
		return json.Marshal(*sve.Name)
	}

	return json.Marshal(sve.Reference)
}
