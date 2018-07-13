package resolver

import (
	"fmt"
	"testing"

	"github.com/mlab-lattice/lattice/pkg/util/git"
)

const workDir = "/tmp/lattice/test/pkg/definition/resolver/component_resolver"

var remoteDir = fmt.Sprintf("%v/remote", workDir)

func TestComponentResolver(t *testing.T) {
	// setup
	setupComponentResolverTest()

	//defer teardownGitResolverTest()
	//t.Run("TestCloneLocalRepo", testCloneLocalRepo)
	//t.Run("TestCloneGithubRepo", testCloneGithubRepo)
	//t.Run("TestTags", testTags)
	//t.Run("TestFileContents", testFileContents)
	//t.Run("TestInvalidURI", testInvalidURI)
}

func setupComponentResolverTest() {
	if err := git.Init(remoteDir); err != nil {
		panic(err)
	}

	hash, err := git.WriteAndCommitFile(remoteDir, "foo", []byte("hello world"), 0700, "my commit")
	if err != nil {
		panic(err)
	}

	err = git.Tag(remoteDir, hash, "1.0.0")
	if err != nil {
		panic(err)
	}
}
