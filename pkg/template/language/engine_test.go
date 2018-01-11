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

const testRepoDir = "/tmp/lattice-core/test/template-engine/my-repo"
const systemFileName = "system.json"
const serviceFileName = "service.json"
const systemFileUrl = "file:///tmp/lattice-core/test/template-engine/my-repo/.git/system.json"
const serviceFileUrl = "file:///tmp/lattice-core/test/template-engine/my-repo/.git/service.json"
const testGitWorkDir = "/tmp/lattice-core/test/test-git-file-repository"

func TestEngine(t *testing.T) {

	fmt.Println("Running template engine tests...")
	setupEngineTest()
	t.Run("TestEngine", doTestEngine)

	teardownEngineTest()
}

func setupEngineTest() {
	fmt.Println("Setting up test")
	// ensure work directory
	os.Mkdir(testRepoDir, 0700)

	gogit.PlainInit(testRepoDir, false)

	commitTestFiles(systemJSON, serviceJSON, "v1")

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

	fmt.Printf("calling EvalFromURL('%s')\n", systemFileUrl)

	parameters := map[string]interface{}{
		"systemName": "mySystem",
	}
	result, err := engine.EvalFromURL(systemFileUrl, parameters, &Options{})

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	if result == nil {
		t.Fatal("Eval result is nil")
	}

	fmt.Println("Evaluation Result: ")
	prettyPrint(result)

	fmt.Println("Validating Eval result...")

	fmt.Println("Validating subsystems...")
	if result["subsystems"] == nil {
		t.Fatal("subsystems is nil")
	}

	fmt.Println("Validating subsystems is of type array...")
	if _, isArray := result["subsystems"].([]interface{}); !isArray {
		t.Fatal("subsystems is not an array!")
	}

	fmt.Println("Validating subsystems length...")

	if len(result["subsystems"].([]interface{})) != 2 {
		t.Fatal("wrong length for subsystems")
	}

	// ensure that some parameters are required
	fmt.Println("ensure that systemName parameter is required...")
	_, err = engine.EvalFromURL(systemFileUrl, nil, &Options{})

	if err == nil || fmt.Sprintf("%v", err) != "parameter systemName is required" {
		t.Fatalf("Required parameter 'systemName' has not been validated")
	}

}

func commitTestFiles(systemJson string, serviceJson string, tag string) {

	systemFileContents := []byte(systemJson)
	ioutil.WriteFile(path.Join(testRepoDir, systemFileName), systemFileContents, 0644)

	serviceFileContents := []byte(serviceJson)
	ioutil.WriteFile(path.Join(testRepoDir, serviceFileName), serviceFileContents, 0644)

	repo, _ := gogit.PlainOpen(testRepoDir)

	workTree, _ := repo.Worktree()

	workTree.Add(systemFileName)

	workTree.Add(serviceFileName)

	// commit
	hash, _ := workTree.Commit("test", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@mlab-lattice.com",
			When:  time.Now(),
		},
	})

	// create the tag
	n := plumbing.ReferenceName("refs/tags/" + tag)
	t := plumbing.NewHashReference(n, hash)
	repo.Storer.SetReference(t)

}

const systemJSON = `
{
  "$parameters": {
     "systemName": {
        "required": true
     }
  },
  "$": {
    "name": "${systemName}",
    "type": "system",
    "description": "This is my system"
  },
  "subsystems": [
    { "$include": {
         "url": "service.json",
         "parameters": {
           "init": false
         }
       }
    },
    {
      "$": {
        "name": "my-service",
        "type": "service",
        "description": "This is my service"
      },
      "components": [
        {
          "name": "service",
          "init": false,
          "build": {
            "docker_image": {
              "registry": "registry.company.com",
              "repository": "foobar",
              "tag": "v1.0.0"
            }
          },
          "exec": {
            "command": [
              "./start",
              "--my-app"
            ],
            "environment": {
              "biz": "baz",
              "foo": "bar"
            }
          }
        }
      ],
      "resources": {
        "min_instances": 1,
        "max_instances": 1,
        "instance_type": "mock.instance.type"
      }
    }
  ]
}
`

const serviceJSON = `
{
  "$": {
    "name": "my-service-2",
    "type": "service",
    "description": "This is my service 2"
  },

  "$parameters": {
      "init": {
         "required": true,
         "default": false
      }
  },

  "components": [
    {
      "name": "service",
      "init": "${init}",
      "build": {
        "docker_image": {
          "registry": "registry.company.com",
          "repository": "foobar",
          "tag": "v1.0.0"
        }
      },
      "exec": {
        "command": [
          "./start",
          "--my-app"
        ],
        "environment": {
          "biz": "baz",
          "foo": "bar"
        }
      }
    }
  ],
  "resources": {
    "min_instances": 1,
    "max_instances": 1,
    "instance_type": "mock.instance.type"
  }
}
`
