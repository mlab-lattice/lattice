package template

import (
	"fmt"
	"testing"
)

func TestTemplateEvaluate(t *testing.T) {
	temp := Template{
		Parameters: Parameters{
			"foo": Parameter{
				Type:    ParameterTypeNumber,
				Default: float64(3),
			},
			"bar": Parameter{
				Type:    ParameterTypeObject,
				Default: map[string]interface{}{"hello": "world"},
			},
		},
		Fields: map[string]interface{}{
			"foo": "bar",
			"baz": []interface{}{
				"buzz",
				map[string]interface{}{
					"boo": "${bar} ${foo}",
				},
			},
			"bat": map[string]interface{}{
				"qux": []interface{}{
					"hello",
					map[string]interface{}{
						"goodnight": "moon",
						"hello":     "${foo}",
					},
				},
			},
		},
	}

	result, err := temp.Evaluate(make(map[string]interface{}))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result)
}
