package language

import (
	"fmt"
	"testing"
)

// TestSecrets
func TestSecrets(t *testing.T) {

	simpleSecretTest(t)

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
