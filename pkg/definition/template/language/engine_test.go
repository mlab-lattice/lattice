package language

import (
	"fmt"

	"strings"
	"testing"
)

const (
	t1File = "t1.json"
	t2File = "t2.json"
)

func TestEngine(t *testing.T) {

	fmt.Println("Running template engine tests...")
	// setup
	setupEngineTest()
	// defer teardown
	defer teardownEngineTest()
	t.Run("TestEngine", doTestEngine)
}

func setupEngineTest() {
	fmt.Println("Setting up test")
	initTestRepo()

	commitTestFile(t1File, t1JSON)
	commitTestFile(t2File, t2JSON)
	commitTestFile("t3.json", t3JSON)

}

func teardownEngineTest() {
	fmt.Println("Tearing down template engine test")
	deleteTestRepo()
}

func doTestEngine(t *testing.T) {
	testInclude(t)
	testValidateParams(t)
	testOperatorSiblings(t)
}

func testInclude(t *testing.T) {

	fmt.Println("Starting TemplateEngine test....")
	engine := NewEngine()
	options, err := CreateOptions(testWorkDir, nil)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	t1FileURL := getTestFileURL(t1File)

	fmt.Printf("calling EvalFromURL('%s')\n", t1FileURL)

	parameters := map[string]interface{}{
		"name": "joe",
	}

	result, err := engine.EvalFromURL(t1FileURL, parameters, options)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if result == nil {
		t.Fatal("Eval result is nil")
	}

	resultMap := result.ValueAsMap()

	fmt.Println("Evaluation Result: ")
	prettyPrint(resultMap)

	fmt.Println("Validating Eval result...")

	if resultMap["name"] != "joe" {
		t.Fatal("wrong name")
	}

	if resultMap["hi"] != "Hi joe" {
		t.Fatal("wrong hi value")
	}

	if int(resultMap["x"].(float64)) != 1 {
		t.Fatal("wrong x val")
	}

	if resultMap["y"] != 2.1 {
		t.Fatal("wrong y val")
	}

	fmt.Println("Validating array...")
	if resultMap["array"] == nil {
		t.Fatal("array is nil")
	}

	fmt.Println("Validating array is of type array...")
	if _, isArray := resultMap["array"].([]interface{}); !isArray {
		t.Fatal("array is not an array!")
	}

	fmt.Println("Validating array length...")

	if len(resultMap["array"].([]interface{})) != 3 {
		t.Fatal("wrong array length")
	}

	fmt.Println("Validating include...")
	if resultMap["address"] == nil {
		t.Fatal("address is nil")
	}

	fmt.Println("Validating address is of type map...")
	if _, isMap := resultMap["address"].(map[string]interface{}); !isMap {
		t.Fatal("address is not a map!")
	}

	fmt.Println("Validating address length...")

	address := resultMap["address"].(map[string]interface{})

	if len(address) != 2 {
		t.Fatal("wrong length of address object")
	}

	// validate parameter passing to $include
	fmt.Println("validate parameter passing to $include")
	if address["city"] != "San Francisco" {
		t.Fatal("invalid city")
	}

	// validate default parameters
	fmt.Println("validate default parameters")
	if address["state"] != "CA" {
		t.Fatal("invalid state")
	}

	fmt.Println("Testing metadata")

	metadata1 := result.GetPropertyMetadata("address")

	if metadata1 == nil {
		t.Fatalf("No metadata found for property address")
	}

	fmt.Println("Validating metadata template url")

	if metadata1.TemplateURL() != "file:///tmp/lattice-core/test/template-engine/my-repo/.git/t1.json" {
		t.Fatalf("invalid template file for address. Found '%s'", metadata1.TemplateURL())
	}

	fmt.Println("Validating metadata line number")

	if metadata1.LineNumber() != 29 {
		t.Fatalf("invalid line number for address.city. Expected 29 but found %v", metadata1.LineNumber())
	}

	fmt.Println("Testing metadata within includes")
	metadata2 := result.GetPropertyMetadata("address.city")

	if metadata2 == nil {
		t.Fatalf("No metadata found for property address.city")
	}

	fmt.Println("Validating metadata template url")

	if metadata2.TemplateURL() != "file:///tmp/lattice-core/test/template-engine/my-repo/.git/t2.json" {
		t.Fatalf("invalid template file for address.city. Found '%s'", metadata2.TemplateURL())
	}

	fmt.Println("Validating metadata line number")

	if metadata2.LineNumber() != 12 {
		t.Fatalf("invalid line number for address.city. Expected 12 but found %v", metadata2.LineNumber())
	}

	fmt.Println("Testing metadata for array elements")
	arrMetadata := result.GetPropertyMetadata("array.0")

	if arrMetadata == nil {
		t.Fatalf("No metadata found for property array.0")
	}

	if arrMetadata.LineNumber() != 24 {
		t.Fatalf("invalid line number for array.0. Expected 24 but found %v", arrMetadata.LineNumber())
	}
}

func testValidateParams(t *testing.T) {
	engine := NewEngine()

	t1FileURL := getTestFileURL(t1File)
	options, _ := CreateOptions(testWorkDir, nil)

	// ensure that some parameters are required
	fmt.Println("ensure that name parameter is required...")
	_, err := engine.EvalFromURL(t1FileURL, nil, options)

	if err == nil || !strings.Contains(fmt.Sprintf("%v", err), "parameter name is required") {
		t.Fatalf("Required parameter 'name' has not been validated")
	}

}

func testOperatorSiblings(t *testing.T) {
	engine := NewEngine()

	t3FileURL := getTestFileURL("t3.json")
	options, _ := CreateOptions(testWorkDir, nil)

	// ensure that some parameters are required
	fmt.Println("ensure that $include disallows other sibling keys")
	_, err := engine.EvalFromURL(t3FileURL, nil, options)
	if err == nil || !strings.Contains(fmt.Sprintf("%v", err), "sibling fields are not") {
		t.Fatalf("$include should disallow siblings")
	}

}

const t1JSON = `
{
  "$parameters": {
     "name": {
        "required": true
     }
  },

  "$variables": {
     "count": 1,
     "object": {
        "x": 1,
        "y": 2.1
     }
  },

  "name": "${name}",
  "hi": "Hi ${name}",

  "x": "${object.x}",
  "y": "${object.y}",

  "array": [
    "item1",
    "item2",
    "item3"
  ],

  "address": {
    "$include": {
      "url": "t2.json",
      "parameters": {
        "city": "San Francisco"
      }
    }
  },

  "int": 1,
  "bool": true

}
`

const t2JSON = `
{
  "$parameters": {
	 "city": {
		"required": true
	 },
	 "state": {
		"default": "CA"
	 }
   },

   "city": "${city}",
   "state": "${state}"
}
`

const t3JSON = `
{
  "address": {
    "bad": "bad sibling for include",
    "$include": {
      "url": "t2.json"
    }
  }
}
`
