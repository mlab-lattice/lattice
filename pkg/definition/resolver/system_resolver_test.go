package resolver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const RESOLVER_TEST_DIR = "/tmp/lattice-core/test/resolver"
const TEST_REPO_DIR = "/tmp/lattice-core/test/resolver/my-repo"
const TEST_REPO_GIT_URI_V1 = "/tmp/lattice-core/test/resolver/my-repo#v1"
const TEST_REPO_GIT_URI_V2 = "/tmp/lattice-core/test/resolver/my-repo#v2"
const TEST_WORK_DIR = "/tmp/lattice-core/test/resolver/work"
const SYSTEM_FILE_NAME = "system.json"
const SERVICE_FILE_NAME = "service.json"

func TestMain(m *testing.M) {
	fmt.Println("Running resolvers tests...")
	setupTest()
	retCode := m.Run()
	teardownTest()
	os.Exit(retCode)
}

func TestValidateSystemResolver(t *testing.T) {

	testV1(t)
	testV2(t)
	testListVersions(t)
}

func testV1(t *testing.T) {
	fmt.Println("--------------- Testing ResolveDefinition V1")

	res, err := NewSystemResolver(TEST_WORK_DIR)
	if err != nil {
		t.Fatalf("Got error calling NewSystemResolver: %v", err)
	}

	defNode, err := res.ResolveDefinition(TEST_REPO_GIT_URI_V1, "system.json", &GitResolveOptions{})
	if err != nil {
		t.Fatalf("Error is not nil: %v", err)
	}

	if defNode.Name() != "my-system-v1" {
		t.Error("Wrong system name")
	}

	if len(defNode.Subsystems()) != 2 {
		t.Error("Wrong # of subsystems")
	}

	if len(defNode.Services()) != 2 {
		t.Error("Wrong # of services")
	}

	if defNode.Subsystems()["/my-system-v1/my-service"].Name() != "my-service" {
		t.Error("Invalid Subsystem map")
	}

}

func testV2(t *testing.T) {

	fmt.Println("--------------- Testing ResolveDefinition V2")

	res, err := NewSystemResolver(TEST_WORK_DIR)
	if err != nil {
		t.Fatalf("Got error calling NewSystemResolver: %v", err)
	}

	defNode, err := res.ResolveDefinition(TEST_REPO_GIT_URI_V2, "system.json", &GitResolveOptions{})
	if err != nil {
		t.Error("Error is not nil: ", err)
	}

	if defNode.Name() != "my-system-v2" {
		t.Error("Wrong system name")
	}
}

func testListVersions(t *testing.T) {
	fmt.Println("--------------- Testing ListDefinitionVersions")

	res, err := NewSystemResolver(TEST_WORK_DIR)
	if err != nil {
		t.Fatalf("Got error calling NewSystemResolver: %v", err)
	}

	versions, err := res.ListDefinitionVersions(TEST_REPO_GIT_URI_V2, &GitResolveOptions{})
	if err != nil {
		t.Fatal("Error is not nil: ", err)
	}

	if len(versions) != 2 {
		t.Error("Wrong # of versions")
	}

	if versions[0] != "v1" || versions[1] != "v2" {
		t.Error("Wrong version")
	}

}

func setupTest() {

	fmt.Println("Setting up resolver test")
	// ensure work directory
	os.Mkdir(TEST_REPO_DIR, 0700)

	git.PlainInit(TEST_REPO_DIR, false)

	commitTestFiles(SYSTEM_JSON, SERVICE_JSON, "v1")
	commitTestFiles(SYSTEM_JSON_V2, SERVICE_JSON, "v2")

}

func commitTestFiles(systemJson string, serviceJson string, tag string) {

	systemFileContents := []byte(systemJson)
	ioutil.WriteFile(path.Join(TEST_REPO_DIR, SYSTEM_FILE_NAME), systemFileContents, 0644)

	serviceFileContents := []byte(serviceJson)
	ioutil.WriteFile(path.Join(TEST_REPO_DIR, SERVICE_FILE_NAME), serviceFileContents, 0644)

	repo, _ := git.PlainOpen(TEST_REPO_DIR)

	workTree, _ := repo.Worktree()

	workTree.Add(SYSTEM_FILE_NAME)

	workTree.Add(SERVICE_FILE_NAME)

	// commit
	hash, _ := workTree.Commit("test", &git.CommitOptions{
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

func teardownTest() {
	fmt.Println("Tearing down resolver test")
	os.RemoveAll(RESOLVER_TEST_DIR)
}

const SYSTEM_JSON = `
{
  "name": "my-system-v1",
  "type": "system",
  "subsystems": [
    {"$include": "service.json"},
    {
      "name": "my-service",
      "type": "service",
      "components": [
        {
          "name": "service",
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

const SERVICE_JSON = `
{
  "name": "my-service-2",
  "type": "service",
  "components": [
    {
      "name": "service",
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

const SYSTEM_JSON_V2 = `
{
  "name": "my-system-v2",
  "type": "system",
  "subsystems": [
    {"$include": "service.json"},
    {
	  "name": "my-service",
      "type": "service",
      "components": [
        {
          "name": "service",
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
