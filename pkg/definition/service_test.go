package definition

import (
	"fmt"
	"testing"

	//"github.com/mlab-lattice/lattice/pkg/definition/block"
	"encoding/json"
	"github.com/mlab-lattice/lattice/pkg/definition/block/mock"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
	"reflect"
)

func Test_NewServiceFromJSON(t *testing.T) {
	valid := []fromJSONTest{
		{
			description: "only type and name",
			bytes:       serviceExpectedJSON(quoted(TypeService), quoted("service"), nil, nil),
			additionalTests: []additionalFromJSONTest{
				{
					description: "check attributes",
					test: func(result interface{}) error {
						service := result.(*Service)

						components := service.Components
						if components != nil {
							return fmt.Errorf("expected Components to be nil but got %#v", components)
						}

						return nil
					},
				},
			},
		},
		{
			description: "type, name, and resources",
			bytes:       serviceExpectedJSON(quoted(TypeService), quoted("service"), nil, mock.ResourcesExpectedJSON()),
			additionalTests: []additionalFromJSONTest{
				{
					description: "check attributes",
					test: func(result interface{}) error {
						service := result.(*Service)

						components := service.Components
						if components != nil {
							return fmt.Errorf("expected Components to be nil but got %#v", components)
						}

						expectedResources := mock.Resources()
						resources := service.Resources

						if !reflect.DeepEqual(expectedResources.MinInstances, resources.MinInstances) {
							return fmt.Errorf("expected Resources.MinInstances to be %v but got %v", resources.MinInstances, expectedResources.MinInstances)
						}

						if !reflect.DeepEqual(expectedResources.MaxInstances, resources.MaxInstances) {
							return fmt.Errorf("expected Resources.MaxInstances to be %v but got %v", resources.MaxInstances, expectedResources.MaxInstances)
						}

						if !reflect.DeepEqual(expectedResources.NumInstances, resources.NumInstances) {
							return fmt.Errorf("expected Resources.InstanceType to be %v but got %v", resources.NumInstances, expectedResources.NumInstances)
						}

						if !reflect.DeepEqual(expectedResources.InstanceType, resources.InstanceType) {
							return fmt.Errorf("expected Resources.InstanceType to be %v but got %v", resources.InstanceType, expectedResources.InstanceType)
						}

						return nil
					},
				},
			},
		},
		{
			description: "type, name, resources, and components",
			bytes: serviceExpectedJSON(
				quoted(TypeService),
				quoted("service"),
				jsonutil.GenerateArrayBytes([][]byte{
					mock.ComponentExpectedJSON(),
				}),
				mock.ResourcesExpectedJSON(),
			),
			additionalTests: []additionalFromJSONTest{
				{
					description: "check attributes",
					test: func(result interface{}) error {
						service := result.(*Service)

						expectedResources := mock.Resources()
						resources := service.Resources

						if !reflect.DeepEqual(*expectedResources, resources) {
							return fmt.Errorf("Resources did not match the serialized version")
						}

						// FIXME: this no longer works since component.exec.environment can have pointers now
						//expectedComponents := []*block.Component{
						//	mock.Component(),
						//}
						//components := service.Components()
						//
						//if !reflect.DeepEqual(expectedComponents, components) {
						//	return fmt.Errorf("Components() did not match the serialized version")
						//}

						return nil
					},
				},
			},
		},
	}

	expectFromJSONSuccesses(t, valid, func(data []byte) (interface{}, error) {
		var service *Service
		err := json.Unmarshal(data, &service)
		return interface{}(service), err
	})

	invalid := []fromJSONTest{
		{
			description: "invalid type",
			bytes:       serviceExpectedJSON(quoted(TypeSystem), nil, nil, nil),
		},
		{
			description: "no type",
			bytes:       serviceExpectedJSON(nil, nil, nil, nil),
		},
		{
			description: "emptystring type",
			bytes:       serviceExpectedJSON(quoted("\"\""), nil, nil, nil),
		},
	}

	expectFromJSONFailures(t, invalid, func(data []byte) (interface{}, error) {
		var service *Service
		err := json.Unmarshal(data, &service)
		return interface{}(service), err
	})
}

func serviceExpectedJSON(
	t,
	name,
	components,
	resources []byte,
) []byte {

	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:      "type",
			Bytes:     t,
			OmitEmpty: true,
		},
		{
			Name:      "name",
			Bytes:     name,
			OmitEmpty: true,
		},
		{
			Name:      "components",
			Bytes:     components,
			OmitEmpty: true,
		},
		{
			Name:      "resources",
			Bytes:     resources,
			OmitEmpty: true,
		},
	})
}
