package git

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
	testRepoDir    = "/tmp/lattice-core/test/git/my-repo"
	testWorkDir    = "/tmp/lattice-core/test/git/work"
	testFile       = "hello.txt"
	localRepoURI1  = "/tmp/lattice-core/test/git/my-repo"
	localRepoURI2  = "/tmp/lattice-core/test/git/my-repo/.git"
	localRepoURI3  = "file:///tmp/lattice-core/test/git/my-repo"
	localRepoURI4  = "file:///tmp/lattice-core/test/git/my-repo/.git"
	localRepoURI5  = "git://tmp/lattice-core/test/git/my-repo"
	localRepoURI6  = "git://tmp/lattice-core/test/git/my-repo/.git"
	remoteRepoURI1 = "https://github.com/mlab-lattice/testing__system"
	remoteRepoURI2 = "https://github.com/mlab-lattice/testing__system.git"
)

func TestGitResolver(t *testing.T) {

	fmt.Println("Running git resolver tests...")
	// setup
	setupGitResolverTest()
	// defer teardown
	defer teardownGitResolverTest()
	t.Run("TestCloneLocalRepo", testCloneLocalRepo)
	t.Run("TestCloneGithubRepo", testCloneGithubRepo)
	t.Run("TestTags", testTags)
	t.Run("TestFileContents", testFileContents)
}

func setupGitResolverTest() {
	fmt.Println("Setting up test...")
	// delete temp dirs in case there was previous tests had some left overs
	deleteTempDirs()

	createTestGitRepo()
	commitTestFile("hello", "v1")
	commitTestFile("hello there", "v2")

}

func teardownGitResolverTest() {
	fmt.Println("Tearing down git resolver test")
	deleteTempDirs()
}

func deleteTempDirs() {
	// remove the test repo
	os.RemoveAll(testRepoDir)
	// remove work dir
	os.RemoveAll(testWorkDir)
}

func testCloneLocalRepo(t *testing.T) {
	testCloneURI(t, localRepoURI1)
	testCloneURI(t, localRepoURI2)
	testCloneURI(t, localRepoURI3)
	testCloneURI(t, localRepoURI4)
	testCloneURI(t, localRepoURI5)
	testCloneURI(t, localRepoURI6)
}

func testCloneGithubRepo(t *testing.T) {
	testCloneURI(t, remoteRepoURI1)
	testCloneURI(t, remoteRepoURI2)
}

func testTags(t *testing.T) {
	resolver, err := NewResolver(testWorkDir)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	ctx := &Context{URI: testRepoDir,
		Options: &Options{},
	}

	fmt.Println("Testing tags")
	// test tags
	tags, err := resolver.GetTagNames(ctx)

	if err != nil {
		t.Fatalf("Failed to get tags: %s", err)
	}

	fmt.Printf("Got tags: %v\n", tags)

	if len(tags) != 2 || tags[0] != "v1" || tags[1] != "v2" {
		t.Fatalf("bad tags: %v. Must be [v1 v2]", tags)
	}

}

func testFileContents(t *testing.T) {
	testFileContent(t, testRepoDir+"#v1", "hello.txt", "hello")
	testFileContent(t, testRepoDir+"#v2", "hello.txt", "hello there")
}

func testFileContent(t *testing.T, uri string, filename string, contents string) {
	resolver, err := NewResolver(testWorkDir)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	ctx := &Context{URI: uri,
		Options: &Options{},
	}

	fmt.Printf("Testing file contents for uri '%v', file '%v' against '%v'\n",
		uri, filename, contents)
	// test tags
	bytes, err := resolver.FileContents(ctx, filename)

	if err != nil {
		t.Fatalf("Got error getting file contents for uri '%v', file '%v'. Error: %v",
			uri, filename, err)
	}

	actualContents := string(bytes)
	if actualContents != contents {
		t.Fatalf("Unexpected contents for uri '%v', file '%v'. Expected '%v' but got '%s'",
			uri, filename, contents, actualContents)
	}
}

func testCloneURI(t *testing.T, uri string) {
	fmt.Printf("Test clone %s\n", uri)

	resolver, err := NewResolver(testWorkDir)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	// test clone
	ctx := &Context{URI: uri,
		Options: &Options{},
	}

	_, err = resolver.Clone(ctx)

	if err != nil {
		t.Fatalf("Failed to clone uri '%v'. Error: %s", uri, err)
	}

}

func createTestGitRepo() {

	// ensure work directory
	os.Mkdir(testRepoDir, 0700)

	// init git repo
	gogit.PlainInit(testRepoDir, false)

}

func commitTestFile(contents string, tag string) {
	ioutil.WriteFile(path.Join(testRepoDir, testFile), []byte(contents), 0644)

	repo, _ := gogit.PlainOpen(testRepoDir)

	workTree, _ := repo.Worktree()

	workTree.Add(testFile)

	// commit
	hash, _ := workTree.Commit("test", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@mlab-lattice.com",
			When:  time.Now(),
		},
	})

	// create the tag
	n := plumbing.ReferenceName(fmt.Sprintf("refs/tags/%s", tag))
	t := plumbing.NewHashReference(n, hash)
	repo.Storer.SetReference(t)
}
