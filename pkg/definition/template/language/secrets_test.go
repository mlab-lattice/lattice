package language

import (
	"fmt"
	"reflect"
	"testing"
)

// TestSecrets
func TestSecrets(t *testing.T) {

	setupSecretsTest()
	defer teardownSecretsTest()
	simpleSecretTest(t)
	templateSecretsTest(t)

}

func simpleSecretTest(t *testing.T) {
	// testing int eval
	fmt.Println("Testing secrets ")
	engine := NewEngine()

	result, err := engine.Eval(map[string]interface{}{
		"secrets": []string{"my-secret"},
		"b": map[string]interface{}{
			"$secret": "my-secret",
		},
	}, nil, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if resultMap, isMap := result.Value().(map[string]interface{}); isMap {
		if ref, isRef := resultMap["b"].(Reference); !isRef || ref["__reference"] != "secrets.my-secret" {
			t.Fatal("Expected result[b][__reference] to be 'secrets.my-secret'")
		}

	} else {
		t.Fatal("Expected result to be of type map")
	}

}

func templateSecretsTest(t *testing.T) {
	engine := NewEngine()
	options, err := CreateOptions(testWorkDir, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	t1FileURL := getTestFileURL("t1.json")

	fmt.Printf("calling EvalFromURL('%s')\n", t1FileURL)

	result, err := engine.EvalFromURL(t1FileURL, nil, options)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	fmt.Println("Evaluation result")
	// validate result

	resultMap := result.ValueAsMap()

	ref1 := resultMap["__references"]
	if ref1 == nil {
		t.Fatalf("No __references populated in results")
	}

	references := ref1.([]interface{})
	if len(references) != 2 {
		t.Fatalf("invalid length for __references")
	}

	expected := []interface{}{
		map[string]interface{}{
			"recipient": "i",
			"target":    "secrets.my_secret",
		},
		map[string]interface{}{
			"recipient": "b.bar",
			"target":    "secrets.my_secret",
		},
	}

	matches := 0
	for _, ref := range references {
		for _, exp := range expected {
			if reflect.DeepEqual(ref, exp) {
				matches++
			}
		}

	}

	if matches != 2 {
		t.Fatalf("invalid __references. Found %v, Expected %v", references, expected)
	}

}

func teardownSecretsTest() {
	fmt.Println("Tearing down secrets test")
	deleteTestRepo()
}

func setupSecretsTest() {
	fmt.Println("Setting up secrets test")
	initTestRepo()

	commitTestFile("t1.json",
		`
{
  "secrets": {
      "my_secret": {}
   },
  "i": {
    "$secret": "my_secret"
  },
  "b": {
    "$include": {
      "url": "t2.json",
      "parameters": {
         "foo": {
           "$secret": "my_secret"
         }
      }
    }
  }
}`)

	commitTestFile("t2.json",
		`
{
  "$parameters": {
    "foo": {
      "required": true
    }
  },

  "bar": "${foo}"
}
`)

}
