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
		"b00l": true,
	}

	// basic var ref tests

	x := evalStringExpression("${x}", variables)

	if _, isInt := x.(int); !(isInt && x == 1) {
		t.Fatal("Expected x to be an integer = 1")
	}

	b00l := evalStringExpression("${b00l}", variables)

	if _, isBool := b00l.(bool); !(isBool && b00l == true) {
		t.Fatal("Expected b00l to be a boolean and is true")
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
