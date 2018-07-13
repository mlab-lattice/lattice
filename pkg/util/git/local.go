package git

import (
	"io/ioutil"
	"os"
	"path/filepath"

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
func Commit(repositoryPath, message string) (Hash, error) {
	worktree, err := worktree(repositoryPath)
	if err != nil {
		return Hash{}, err
	}

	hash, err := worktree.Commit(message, &git.CommitOptions{
		Author: &gitplumbingobject.Signature{
			Name: "lattice",
		},
	})
	if err != nil {
		return Hash{}, err
	}

	h := Hash{hash}
	return h, nil
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
) (Hash, error) {
	if err := WriteAndAddFile(repositoryPath, path, contents, perm); err != nil {
		return Hash{}, err
	}

	return Commit(repositoryPath, message)
}

// Tag attempts to create a lightweight tag referencing the hash.
func Tag(repositoryPath string, hash Hash, name string) error {
	repository, err := git.PlainOpen(repositoryPath)
	if err != nil {
		return err
	}

	refName := gitplumbing.ReferenceName("refs/tags/" + name)
	t := gitplumbing.NewHashReference(refName, hash.Hash)
	return repository.Storer.SetReference(t)
}

func worktree(repositoryPath string) (*git.Worktree, error) {
	repository, err := git.PlainOpen(repositoryPath)
	if err != nil {
		return nil, err
	}

	return repository.Worktree()
}
