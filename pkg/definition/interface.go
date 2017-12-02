package definition

import (
	"encoding/json"
	"errors"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

// IMPORTANT: it is assumed that all system/definition.Interface implementers
// also implement system/definition/block.Interface.

type Interface interface {
	Metadata() *block.Metadata
}

// Have to jump through some hoops to properly unmarshal json into
// a definition.Interface. All definition json unmarshaling must go
// through NewFromJSON.

func UnmarshalJSON(bytes []byte) (Interface, error) {
	var decoded decoder
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return nil, err
	}

	if err := decoded.definition.(block.Interface).Validate(nil); err != nil {
		return nil, err
	}

	return decoded.definition, nil
}

type decoder struct {
	definition Interface
}

type metadataDecoder struct {
	Metadata block.Metadata `json:"$"`
}

type subsystemsDecoder struct {
	Subsystems []interface{} `json:"subsystems"`
}

func (u *decoder) UnmarshalJSON(data []byte) error {
	var dm metadataDecoder
	if err := json.Unmarshal(data, &dm); err != nil {
		return err
	}

	switch dm.Metadata.Type {
	case SystemType:
		system := System{
			Meta: dm.Metadata,
		}

		var ds subsystemsDecoder
		if err := json.Unmarshal(data, &ds); err != nil {
			return err
		}

		subsystems := []Interface{}
		for _, subsystem := range ds.Subsystems {
			subsystemBytes, err := json.Marshal(subsystem)
			if err != nil {
				return err
			}

			var unpacked decoder
			if err := json.Unmarshal(subsystemBytes, &unpacked); err != nil {
				return nil
			}

			subsystems = append(subsystems, unpacked.definition)
		}

		system.Subsystems = subsystems
		u.definition = Interface(&system)
	case ServiceType:
		var service Service
		if err := json.Unmarshal(data, &service); err != nil {
			return err
		}

		u.definition = Interface(&service)
	default:
		// TODO: maybe process template files here?
		return errors.New("unrecognized type " + dm.Metadata.Type)
	}

	return nil
}
