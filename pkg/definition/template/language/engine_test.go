package language

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const (
	testRepoDir    = "/tmp/lattice-core/test/template-engine/my-repo"
	t1File         = "t1.json"
	t2File         = "t2.json"
	t1FileUrl      = "file:///tmp/lattice-core/test/template-engine/my-repo/.git/t1.json"
	t2FileUrl      = "file:///tmp/lattice-core/test/template-engine/my-repo/.git/t2.json"
	testGitWorkDir = "/tmp/lattice-core/test/test-git-file-repository"
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
	// ensure work directory
	os.Mkdir(testRepoDir, 0700)

	gogit.PlainInit(testRepoDir, false)

	commitTestFiles()

}

func teardownEngineTest() {
	fmt.Println("Tearing down template engine test")
	// remove the test repo
	os.RemoveAll(testRepoDir)
	// remove the work directory TODO replace with testGitWorkDir when we allow passing work directory in git options
	os.RemoveAll(gitWorkDirectory)
}

func doTestEngine(t *testing.T) {

	fmt.Println("Starting TemplateEngine test....")
	engine := NewEngine()

	fmt.Printf("calling EvalFromURL('%s')\n", t1FileUrl)

	parameters := map[string]interface{}{
		"name": "joe",
	}
	result, err := engine.EvalFromURL(t1FileUrl, parameters, &Options{})

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if result == nil {
		t.Fatal("Eval result is nil")
	}

	fmt.Println("Evaluation Result: ")
	prettyPrint(result)

	fmt.Println("Validating Eval result...")

	if result["name"] != "joe" {
		t.Fatal("wrong name")
	}

	if result["hi"] != "Hi joe" {
		t.Fatal("wrong hi value")
	}

	if int(result["x"].(float64)) != 1 {
		t.Fatal("wrong x val")
	}

	if result["y"] != 2.1 {
		t.Fatal("wrong y val")
	}

	fmt.Println("Validating array...")
	if result["array"] == nil {
		t.Fatal("array is nil")
	}

	fmt.Println("Validating array is of type array...")
	if _, isArray := result["array"].([]interface{}); !isArray {
		t.Fatal("array is not an array!")
	}

	fmt.Println("Validating array length...")

	if len(result["array"].([]interface{})) != 3 {
		t.Fatal("wrong array length")
	}

	// ensure that some parameters are required
	fmt.Println("ensure that name parameter is required...")
	_, err = engine.EvalFromURL(t1FileUrl, nil, &Options{})

	if err == nil || fmt.Sprintf("%v", err) != "parameter name is required" {
		t.Fatalf("Required parameter 'name' has not been validated")
	}

	fmt.Println("Validating include...")
	if result["address"] == nil {
		t.Fatal("address is nil")
	}

	fmt.Println("Validating address is of type map...")
	if _, isMap := result["address"].(map[string]interface{}); !isMap {
		t.Fatal("address is not a map!")
	}

	fmt.Println("Validating address length...")

	address := result["address"].(map[string]interface{})

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

	// validate $include to parent
	fmt.Println("validate include to parent")

	if result["city"] != "San Francisco" {
		t.Fatal("invalid city")
	}

	if result["state"] != "CA" {
		t.Fatal("invalid state")
	}

	// ensure that some parameters are required
	fmt.Println("ensure that name parameter is required...")
	_, err = engine.EvalFromURL(t1FileUrl, nil, &Options{})

	if err == nil || fmt.Sprintf("%v", err) != "parameter name is required" {
		t.Fatalf("Required parameter 'name' has not been validated")
	}

}

func commitTestFiles() {

	ioutil.WriteFile(path.Join(testRepoDir, t1File), []byte(t1JSON), 0644)

	ioutil.WriteFile(path.Join(testRepoDir, t2File), []byte(t2JSON), 0644)

	repo, _ := gogit.PlainOpen(testRepoDir)

	workTree, _ := repo.Worktree()

	workTree.Add(t1File)

	workTree.Add(t2File)

	// commit
	hash, _ := workTree.Commit("test", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@mlab-lattice.com",
			When:  time.Now(),
		},
	})

	// create the tag
	n := plumbing.ReferenceName("refs/tags/testv1")
	t := plumbing.NewHashReference(n, hash)
	repo.Storer.SetReference(t)

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

  "$include": {
    "url": "t2.json",
    "parameters": {
	  "city": "San Francisco"
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
