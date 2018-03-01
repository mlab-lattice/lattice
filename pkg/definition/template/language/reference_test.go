package language

import (
	"fmt"
	"testing"
)

// TestReference
func TestReference(t *testing.T) {

	// basic reference tests
	setupReferenceTest()
	defer teardownReferenceTest()

	simpleReferenceTest(t)
	templateReferenceTest(t)

}

func simpleReferenceTest(t *testing.T) {
	// testing int eval
	fmt.Println("Testing reference ")
	engine := NewEngine()

	result, err := engine.Eval(map[string]interface{}{
		"a": 1,
		"b": map[string]interface{}{
			"$reference": "a",
		},
	}, nil, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if resultMap, isMap := result.Value().(map[string]interface{}); isMap {
		if ref, isRef := resultMap["b"].(Reference); !isRef || ref["__reference"] != "a" {
			t.Fatal("Expected result[b][reference] to be 'a'")
		}

	} else {
		t.Fatal("Expected result to be of type map")
	}
}

func templateReferenceTest(t *testing.T) {
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

	prettyPrint(result.Value())

}

func setupReferenceTest() {
	fmt.Println("Setting up reference test")
	initTestRepo()

	commitTestFile("t1.json",
		`
{
  "a": {
    "x": 1
  },
  "i": {
    "$reference": "a.x"
  },
  "b": {
    "$include": {
      "url": "t2.json",
      "parameters": {
         "foo": {
           "$reference": "a.x"
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

  "bar": "${foo}",
  "z": 1,
  "c": {
    "$include": {
      "url": "t3.json",
      "parameters": {
         "baz": {
           "$reference": "z"
         },
         "foo": "${foo}"
      }
    }
  }
}
`)

	commitTestFile("t3.json",
		`
{
  "$parameters": {
    "baz": {
      "required": true
    },
    "foo": {
      "required": true
    }
  },

  "foo": "${foo}",
  "baz": ["${baz}"],
  "car": {
    "__reference": "a.x"
  }
}
`)

}

func teardownReferenceTest() {
	fmt.Println("Tearing down template engine test")
	deleteTestRepo()
}
