package definition

import (
	"fmt"
	"testing"

	//"github.com/mlab-lattice/system/pkg/definition/block"
	//"github.com/mlab-lattice/system/pkg/definition/block/mock"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func Test_NewSystemFromJSON(t *testing.T) {
	valid := []fromJSONTest{
		{
			description: "only type and name",
			bytes:       systemExpectedJSON(quoted(TypeSystem), quoted("system"), nil),
			additionalTests: []additionalFromJSONTest{
				{
					description: "check attributes",
					test: func(result interface{}) error {
						system := result.(System)

						subsystems := system.Subsystems()
						if subsystems != nil {
							return fmt.Errorf("expected Subsystems() to be nil but got %#v", subsystems)
						}

						return nil
					},
				},
			},
		},
		//{
		//	description: "type, name, and resources",
		//	bytes:       serviceExpectedJSON(quoted(TypeService), quoted("service"), nil, nil, mock.ResourcesExpectedJSON()),
		//	additionalTests: []additionalFromJSONTest{
		//		{
		//			description: "check attributes",
		//			test: func(result interface{}) error {
		//				service := result.(Service)
		//
		//				volumes := service.Volumes()
		//				if volumes != nil {
		//					return fmt.Errorf("expected Volumes() to be nil but got %#v", volumes)
		//				}
		//
		//				components := service.Components()
		//				if components != nil {
		//					return fmt.Errorf("expected Components() to be nil but got %#v", components)
		//				}
		//
		//				expectedResources := mock.Resources()
		//				resources := service.Resources()
		//
		//				if !reflect.DeepEqual(expectedResources.MinInstances, resources.MinInstances) {
		//					return fmt.Errorf("expected Resources().MinInstances to be %v but got %v", resources.MinInstances, expectedResources.MinInstances)
		//				}
		//
		//				if !reflect.DeepEqual(expectedResources.MaxInstances, resources.MaxInstances) {
		//					return fmt.Errorf("expected Resources().MaxInstances to be %v but got %v", resources.MaxInstances, expectedResources.MaxInstances)
		//				}
		//
		//				if !reflect.DeepEqual(expectedResources.NumInstances, resources.NumInstances) {
		//					return fmt.Errorf("expected Resources().InstanceType to be %v but got %v", resources.NumInstances, expectedResources.NumInstances)
		//				}
		//
		//				if !reflect.DeepEqual(expectedResources.InstanceType, resources.InstanceType) {
		//					return fmt.Errorf("expected Resources().InstanceType to be %v but got %v", resources.InstanceType, expectedResources.InstanceType)
		//				}
		//
		//				return nil
		//			},
		//		},
		//	},
		//},
		//{
		//	description: "type, name, resources, and components",
		//	bytes: serviceExpectedJSON(
		//		quoted(TypeService),
		//		quoted("service"),
		//		nil,
		//		jsonutil.GenerateArrayBytes([][]byte{
		//			mock.ComponentExpectedJSON(),
		//		}),
		//		mock.ResourcesExpectedJSON(),
		//	),
		//	additionalTests: []additionalFromJSONTest{
		//		{
		//			description: "check attributes",
		//			test: func(result interface{}) error {
		//				service := result.(Service)
		//
		//				volumes := service.Volumes()
		//				if volumes != nil {
		//					return fmt.Errorf("expected Volumes() to be nil but got %#v", volumes)
		//				}
		//
		//				expectedResources := mock.Resources()
		//				resources := service.Resources()
		//
		//				if !reflect.DeepEqual(*expectedResources, resources) {
		//					return fmt.Errorf("Resources() did not match the serialized version")
		//				}
		//
		//				expectedComponents := []*block.Component{
		//					mock.Component(),
		//				}
		//				components := service.Components()
		//
		//				if !reflect.DeepEqual(expectedComponents, components) {
		//					return fmt.Errorf("Components() did not match the serialized version")
		//				}
		//
		//				return nil
		//			},
		//		},
		//	},
		//},
	}

	expectFromJSONSuccesses(t, valid, func(data []byte) (interface{}, error) {
		system, err := NewSystemFromJSON(data)
		return interface{}(system), err
	})

	invalid := []fromJSONTest{
		{
			description: "invalid type",
			bytes:       serviceExpectedJSON(quoted(TypeService), nil, nil, nil, nil),
		},
		{
			description: "no type",
			bytes:       serviceExpectedJSON(nil, nil, nil, nil, nil),
		},
		{
			description: "emptystring type",
			bytes:       serviceExpectedJSON(quoted("\"\""), nil, nil, nil, nil),
		},
	}

	expectFromJSONFailures(t, invalid, func(data []byte) (interface{}, error) {
		system, err := NewSystemFromJSON(data)
		return interface{}(system), err
	})
}

func systemExpectedJSON(
	t,
	name,
	subsystems []byte,
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
			Name:      "subsystems",
			Bytes:     subsystems,
			OmitEmpty: true,
		},
	})
}
