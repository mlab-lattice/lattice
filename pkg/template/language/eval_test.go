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

func TestParametersEval(t *testing.T) {

	fmt.Println("Testing eval with parameters")
	engine := NewEngine()

	m := map[string]interface{}{
		"$parameters": map[string]interface{}{
			"name": map[string]interface{}{
				"required": true,
			},
			"foo": map[string]interface{}{
				"required": false,
				"default":  1,
			},
		},
		"myName": "${name}",
		"myFoo":  "${foo}",
	}
	_, err := engine.Eval(m, nil, nil)

	if err == nil || fmt.Sprintf("%v", err) != "parameter name is required" {
		t.Fatal("Expected error 'parameter name is required'")
	}

	result, err := engine.Eval(m, map[string]interface{}{
		"name": "test",
	}, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if resultMap, isMap := result.(map[string]interface{}); isMap {
		if len(resultMap) != 2 {
			fmt.Println(resultMap)
			t.Fatal("Expected map size to be 1")
		}

		// test default values
		if resultMap["myFoo"] != 1 {
			t.Fatal("Expected result[myFoo] to be 2")
		}

	} else {
		t.Fatal("Expected result to be of type map")
	}

}

func TestVariablesEval(t *testing.T) {

	fmt.Println("Testing eval with variables")
	engine := NewEngine()

	m := map[string]interface{}{
		"$variables": map[string]interface{}{
			"name": "test",
			"foo":  1,
		},
		"myName": "${name}",
		"myFoo":  "${foo}",
	}
	_, err := engine.Eval(m, nil, nil)

	result, err := engine.Eval(m, nil, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if resultMap, isMap := result.(map[string]interface{}); isMap {
		if len(resultMap) != 2 {
			fmt.Println(resultMap)
			t.Fatal("Expected map size to be 2")
		}

		// test default values
		if resultMap["myFoo"] != 1 {
			t.Fatal("Expected result[myFoo] to be 1")
		}

	} else {
		t.Fatal("Expected result to be of type map")
	}

}
