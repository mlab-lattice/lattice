package language

import (
	"fmt"
	"testing"
)

func TestVariables(t *testing.T) {

	fmt.Println("Testing variables")
	variables := map[string]interface{}{
		"x": 1,
		"y": 2,
		"z": map[string]interface{}{
			"foo": 3,
		},
	}

	result := evalStringExpression("x = ${x}", variables)

	if result != "x = 1" {
		t.Fatalf("Invalid result: %v", result)
	}

	fmt.Println(result)

	result = evalStringExpression("${x} ${y}", variables)

	if result != "1 2" {
		t.Fatalf("Invalid result: %v", result)
	}

	fmt.Println(result)

	result = evalStringExpression("z.foo = ${z.foo}", variables)

	if result != "z.foo = 3" {
		t.Fatalf("Invalid result: %v", result)
	}

	fmt.Println(result)

}
