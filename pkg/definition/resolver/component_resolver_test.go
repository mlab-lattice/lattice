package resolver

import (
	"fmt"
	"testing"

	"encoding/json"
	defintionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

const workDir = "/tmp/lattice/test/pkg/definition/resolver/component_resolver"

var (
	remote1Dir = fmt.Sprintf("%v/remote1", workDir)

	oneXTag              = "1.x"
	remote1OneXReference = defintionv1.Reference{
		GitRepository: &defintionv1.GitRepositoryReference{
			GitRepository: &defintionv1.GitRepository{
				URL: fmt.Sprintf("file://%v", remote1Dir),
				Tag: &oneXTag,
			},
			File: "service.json",
		},
	}
	service = defintionv1.Service{
		Container: defintionv1.Container{
			Exec: &defintionv1.ContainerExec{
				Command: []string{"npm", "install"},
			},
		},
	}
)

func TestComponentResolver(t *testing.T) {
	// setup
	setupComponentResolverTest()
	r, err := NewReferenceResolver(workDir)
	if err != nil {
		panic(err)
	}

	c, err := r.ResolveReference(nil, &remote1OneXReference)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", c)

	//defer teardownGitResolverTest()
	//t.Run("TestCloneLocalRepo", testCloneLocalRepo)
	//t.Run("TestCloneGithubRepo", testCloneGithubRepo)
	//t.Run("TestTags", testTags)
	//t.Run("TestFileContents", testFileContents)
	//t.Run("TestInvalidURI", testInvalidURI)
}

func setupComponentResolverTest() {
	if err := git.Init(remote1Dir); err != nil {
		panic(err)
	}

	serviceBytes, err := json.Marshal(&service)
	if err != nil {
		panic(err)
	}

	hash, err := git.WriteAndCommitFile(remote1Dir, "service.json", serviceBytes, 0700, "my commit")
	if err != nil {
		panic(err)
	}

	err = git.Tag(remote1Dir, hash, "1.0.0")
	if err != nil {
		panic(err)
	}

	service.Description = "updated"
	serviceBytes, err = json.Marshal(&service)
	if err != nil {
		panic(err)
	}

	hash, err = git.WriteAndCommitFile(remote1Dir, "service.json", serviceBytes, 0700, "my commit")
	if err != nil {
		panic(err)
	}

	err = git.Tag(remote1Dir, hash, "1.1.0")
	if err != nil {
		panic(err)
	}
}
