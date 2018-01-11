package language

import (
	"fmt"
	"testing"
)

func TestBasicEval(t *testing.T) {

	fmt.Println("Testing eval int")
	engine := NewEngine()

	result, err := engine.Eval(1, nil, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if result != 1 {
		t.Fatal("Expected result to be 1")
	}

	fmt.Println("Testing eval string")

	result, err = engine.Eval("abc", nil, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if result != "abc" {
		t.Fatal("Expected result to be 'abc'")
	}

	fmt.Println("Testing eval bool")

	result, err = engine.Eval(true, nil, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if result != true {
		t.Fatal("Expected result to be true")
	}

	fmt.Println("Testing eval map")

	result, err = engine.Eval(map[string]interface{}{"x": 1}, nil, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if resultMap, isMap := result.(map[string]interface{}); isMap {
		if resultMap["x"] != 1 {
			t.Fatal("Expected result[x] to be 1")
		}

	} else {
		t.Fatal("Expected result to be of type map")
	}

}
