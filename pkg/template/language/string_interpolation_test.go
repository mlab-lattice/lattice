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

	result, err := evalStringExpression("x = ${x}", variables)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if result != "x = 1" {
		t.Fatalf("Invalid result: %v", result)
	}

	fmt.Println(result)

	result, err = evalStringExpression("${x} ${y}", variables)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if result != "1 2" {
		t.Fatalf("Invalid result: %v", result)
	}

	fmt.Println(result)

	result, err = evalStringExpression("z.foo = ${z.foo}", variables)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if result != "z.foo = 3" {
		t.Fatalf("Invalid result: %v", result)
	}

	fmt.Println(result)

}
