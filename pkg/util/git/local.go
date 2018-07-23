package git

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"fmt"
	"gopkg.in/src-d/go-git.v4"
	gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
	gitplumbingobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

func Init(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}

	_, err := git.PlainInit(path, false)
	return err
}

// WriteFile will write the contents to the file in the repository.
// Currently this does not check to ensure that the path does not escape the
// directory, so it could be used to write files not in the repository.
func WriteFile(repositoryPath, path string, contents []byte, perm os.FileMode) error {
	filePath := filepath.Join(repositoryPath, path)
	return ioutil.WriteFile(filePath, contents, perm)
}

// AddFile will add the the path to the staged index.
func AddFile(repositoryPath, path string) error {
	worktree, err := worktree(repositoryPath)
	if err != nil {
		return err
	}

	_, err = worktree.Add(path)
	return err
}

// Commit attempts to commit the repository.
func Commit(repositoryPath, message string) (gitplumbing.Hash, error) {
	worktree, err := worktree(repositoryPath)
	if err != nil {
		return gitplumbing.ZeroHash, err
	}

	hash, err := worktree.Commit(message, &git.CommitOptions{
		Author: &gitplumbingobject.Signature{
			Name: "lattice",
		},
	})
	if err != nil {
		return gitplumbing.ZeroHash, err
	}

	return hash, nil
}

func WriteAndAddFile(repositoryPath, path string, contents []byte, perm os.FileMode) error {
	if err := WriteFile(repositoryPath, path, contents, perm); err != nil {
		return err
	}

	return AddFile(repositoryPath, path)
}

func WriteAndCommitFile(
	repositoryPath string,
	path string,
	contents []byte,
	perm os.FileMode,
	message string,
) (gitplumbing.Hash, error) {
	if err := WriteAndAddFile(repositoryPath, path, contents, perm); err != nil {
		return gitplumbing.ZeroHash, err
	}

	return Commit(repositoryPath, message)
}

// Tag attempts to create a lightweight tag referencing the hash.
func Tag(repositoryPath string, commit gitplumbing.Hash, name string) error {
	repository, err := git.PlainOpen(repositoryPath)
	if err != nil {
		return err
	}

	refName := gitplumbing.ReferenceName("refs/tags/" + name)
	t := gitplumbing.NewHashReference(refName, commit)
	return repository.Storer.SetReference(t)
}

func GetBranchHeadCommit(repositoryPath, branch string) (gitplumbing.Hash, error) {
	repository, err := git.PlainOpen(repositoryPath)
	if err != nil {
		return gitplumbing.ZeroHash, err
	}

	refName := gitplumbing.ReferenceName(fmt.Sprintf("%s:refs/remotes/origin", branch))
	gitRef, err := repository.Reference(refName, false)
	if err != nil || gitRef == nil {
		return gitplumbing.ZeroHash, fmt.Errorf("invalid branch name %v", branch)
	}

	return gitRef.Hash(), nil
}

func CheckoutCommit(repositoryPath string, commit gitplumbing.Hash) error {
	wt, err := worktree(repositoryPath)
	if err != nil {
		return err
	}

	opts := &git.CheckoutOptions{Hash: commit}
	return wt.Checkout(opts)
}

func CheckoutBranch(repositoryPath, branch string) error {
	wt, err := worktree(repositoryPath)
	if err != nil {
		return err
	}

	opts := &git.CheckoutOptions{
		Branch: gitplumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
	}
	return wt.Checkout(opts)
}

func CreateBranch(repositoryPath, branch string, commit gitplumbing.Hash) error {
	wt, err := worktree(repositoryPath)
	if err != nil {
		return err
	}

	opts := &git.CheckoutOptions{
		Hash:   commit,
		Branch: gitplumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		Create: true,
	}
	return wt.Checkout(opts)
}

func worktree(repositoryPath string) (*git.Worktree, error) {
	repository, err := git.PlainOpen(repositoryPath)
	if err != nil {
		return nil, err
	}

	return repository.Worktree()
}
