package template

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

func TestParameterBind(t *testing.T) {
	tests := []struct {
		description string
		params      Parameters
		bindings    map[string]interface{}
		path        tree.Path
		result      map[string]interface{}
		err         error
	}{
		// boolean parameter
		{
			description: "default boolean",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeBool,
					Default: true,
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": true},
			err:      nil,
		},
		{
			description: "wrong default type boolean",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeBool,
					Default: "hello",
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeBool),
				Actual:   "string",
			},
		},
		{
			description: "bound boolean",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeBool,
				},
			},
			bindings: map[string]interface{}{"foo": true},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": true},
			err:      nil,
		},
		{
			description: "overridden boolean",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeBool,
					Default: false,
				},
			},
			bindings: map[string]interface{}{"foo": true},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": true},
			err:      nil,
		},
		{
			description: "wrong type boolean",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeBool,
				},
			},
			bindings: map[string]interface{}{"foo": "hello"},
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeBool),
				Actual:   "string",
			},
		},

		// number parameter
		{
			description: "default number",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeNumber,
					Default: float64(5),
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": float64(5)},
			err:      nil,
		},
		{
			description: "wrong default type number",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeNumber,
					Default: "hello",
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeNumber),
				Actual:   "string",
			},
		},
		{
			description: "bound number",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeNumber,
				},
			},
			bindings: map[string]interface{}{"foo": float64(5)},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": float64(5)},
			err:      nil,
		},
		{
			description: "overridden number",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeNumber,
					Default: float64(5),
				},
			},
			bindings: map[string]interface{}{"foo": float64(6)},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": float64(6)},
			err:      nil,
		},
		{
			description: "wrong type number",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeNumber,
				},
			},
			bindings: map[string]interface{}{"foo": "hello"},
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeNumber),
				Actual:   "string",
			},
		},

		// string parameter
		{
			description: "default string",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeString,
					Default: "bar",
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": "bar"},
			err:      nil,
		},
		{
			description: "wrong default type string",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeString,
					Default: false,
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeString),
				Actual:   "bool",
			},
		},
		{
			description: "bound string",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeString,
				},
			},
			bindings: map[string]interface{}{"foo": "bar"},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": "bar"},
			err:      nil,
		},
		{
			description: "overridden string",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeString,
					Default: "bar",
				},
			},
			bindings: map[string]interface{}{"foo": "baz"},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": "baz"},
			err:      nil,
		},
		{
			description: "wrong type string",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeString,
				},
			},
			bindings: map[string]interface{}{"foo": false},
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeString),
				Actual:   "bool",
			},
		},

		// array parameter
		{
			description: "default array",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeArray,
					Default: []interface{}{"bar"},
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": []interface{}{"bar"}},
			err:      nil,
		},
		{
			description: "wrong default type array",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeArray,
					Default: false,
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeArray),
				Actual:   "bool",
			},
		},
		{
			description: "bound array",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeArray,
				},
			},
			bindings: map[string]interface{}{"foo": []interface{}{"bar"}},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": []interface{}{"bar"}},
			err:      nil,
		},
		{
			description: "overridden array",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeArray,
					Default: []interface{}{"bar"},
				},
			},
			bindings: map[string]interface{}{"foo": []interface{}{"baz"}},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": []interface{}{"baz"}},
			err:      nil,
		},
		{
			description: "wrong type array",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeArray,
				},
			},
			bindings: map[string]interface{}{"foo": false},
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeArray),
				Actual:   "bool",
			},
		},

		// object parameter
		{
			description: "default object",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeObject,
					Default: map[string]interface{}{"bar": "baz"},
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}},
			err:      nil,
		},
		{
			description: "wrong default type object",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeObject,
					Default: false,
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeObject),
				Actual:   "bool",
			},
		},
		{
			description: "bound object",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeObject,
				},
			},
			bindings: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}},
			err:      nil,
		},
		{
			description: "overridden object",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeObject,
					Default: map[string]interface{}{"bar": "baz"},
				},
			},
			bindings: map[string]interface{}{"foo": map[string]interface{}{"buzz": "qux"}},
			path:     tree.RootPath(),
			result:   map[string]interface{}{"foo": map[string]interface{}{"buzz": "qux"}},
			err:      nil,
		},
		{
			description: "wrong type object",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeObject,
				},
			},
			bindings: map[string]interface{}{"foo": false},
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeObject),
				Actual:   "bool",
			},
		},

		// secret parameter
		{
			description: "default secret",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeSecret,
					Default: map[string]interface{}{"$secret": "baz"},
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.Path("/foo/bar"),
			result:   map[string]interface{}{"foo": &definitionv1.SecretRef{Value: tree.PathSubcomponent("/foo/bar:baz")}},
			err:      nil,
		},
		{
			description: "wrong default type secret",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeSecret,
					Default: false,
				},
			},
			bindings: make(map[string]interface{}),
			path:     tree.Path("/foo/bar"),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeSecret),
				Actual:   "bool",
			},
		},
		{
			description: "bound secret",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeSecret,
				},
			},
			bindings: map[string]interface{}{"foo": &definitionv1.SecretRef{Value: tree.PathSubcomponent("/foo/bar:baz")}},
			path:     tree.Path("/foo/bar"),
			result:   map[string]interface{}{"foo": &definitionv1.SecretRef{Value: tree.PathSubcomponent("/foo/bar:baz")}},
			err:      nil,
		},
		{
			description: "overridden secret",
			params: Parameters{
				"foo": Parameter{
					Type:    ParameterTypeSecret,
					Default: map[string]interface{}{"$secret": "baz"},
				},
			},
			bindings: map[string]interface{}{"foo": &definitionv1.SecretRef{Value: tree.PathSubcomponent("/foo/bar:qux")}},
			path:     tree.Path("/foo/bar"),
			result:   map[string]interface{}{"foo": &definitionv1.SecretRef{Value: tree.PathSubcomponent("/foo/bar:qux")}},
			err:      nil,
		},
		{
			description: "wrong type secret",
			params: Parameters{
				"foo": Parameter{
					Type: ParameterTypeSecret,
				},
			},
			bindings: map[string]interface{}{"foo": map[string]interface{}{"$secret": "foo"}},
			path:     tree.RootPath(),
			result:   nil,
			err: &ParameterTypeError{
				Expected: string(ParameterTypeSecret),
				Actual:   "map[string]interface {}",
			},
		},
	}

	for _, test := range tests {
		result, err := test.params.Bind(test.path, test.bindings)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("for %v expected err %v but got %v", test.description, test.err, err)
			continue
		}

		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("for %v expected result %v but got %v", test.description, test.result, result)
			continue
		}
	}
}
