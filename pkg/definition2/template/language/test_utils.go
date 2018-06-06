package language

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const (
	testRepoDir = "/tmp/lattice-core/test/template-engine/my-repo"
	testWorkDir = "/tmp/lattice-core/test/engine"

	baseFileURL = "file:///tmp/lattice-core/test/template-engine/my-repo/.git"
)

func initTestRepo() {
	// ensure work directory
	os.Mkdir(testRepoDir, 0700)

	gogit.PlainInit(testRepoDir, false)
}

func deleteTestRepo() {
	// remove the test repo
	os.RemoveAll(testRepoDir)
	// remove work dir
	os.RemoveAll(testWorkDir)
}
func commitTestFile(fileName string, jsonStr string) {

	ioutil.WriteFile(path.Join(testRepoDir, fileName), []byte(jsonStr), 0644)

	repo, _ := gogit.PlainOpen(testRepoDir)

	workTree, _ := repo.Worktree()

	workTree.Add(fileName)

	// commit
	workTree.Commit("test", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@mlab-lattice.com",
			When:  time.Now(),
		},
	})

}

func getTestFileURL(fileName string) string {
	return fmt.Sprintf("%v/%v", baseFileURL, fileName)
}
