package resolver

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

func NewComponent(m map[string]interface{}) (component.Interface, error) {
	t, err := component.TypeFromMap(m)
	if err != nil {
		return nil, err
	}

	switch t.APIVersion {
	case definitionv1.APIVersion:
		return definitionv1.NewComponent(m)

	default:
		return nil, fmt.Errorf("invalid type api %v", t.APIVersion)
	}
}
