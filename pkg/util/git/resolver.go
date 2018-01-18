package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"golang.org/x/crypto/ssh"

	git "gopkg.in/src-d/go-git.v4"
	gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
	gitplumbingobject "gopkg.in/src-d/go-git.v4/plumbing/object"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	gitUserGit       = "git"
	remoteNameOrigin = "origin"
)

// Resolver provides utility methods for manipulating git repositories
// under a specific working directory on the filesystem.
type Resolver struct {
	WorkDirectory string
}

// Context contains information about the current operation being invoked.
type Options struct {
	SSHKey []byte
}

// Context contains information about the current operation being invoked.
type Context struct {
	URI     string
	Options *Options
}

func NewResolver(workDirectory string) (*Resolver, error) {
	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	err := os.MkdirAll(workDirectory, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to create git resolver work directory: %v", err)
	}

	sr := &Resolver{
		WorkDirectory: workDirectory,
	}
	return sr, nil
}

// Clone will  open the repository and return it If the repository specified in the Context has already been cloned,
// otherwise it will attempt to clone the repository and on success return the cloned repository
func (r *Resolver) Clone(ctx *Context) (*git.Repository, error) {
	repoDir := r.GetRepositoryPath(ctx)

	// If the repository already exists, simply open it.
	repoExists, err := pathExists(repoDir)
	if err != nil {
		return nil, err
	}

	if repoExists {
		repo, err := git.PlainOpen(repoDir)
		return repo, err
	}

	// Otherwise try to clone the repository.
	cloneOptions := git.CloneOptions{
		URL:      ctx.URI,
		Progress: os.Stdout,
	}

	// If an SSH key was supplied, try to use it.
	if ctx.Options.SSHKey != nil {
		signer, err := ssh.ParsePrivateKey([]byte(ctx.Options.SSHKey))
		if err != nil {
			return nil, err
		}

		auth := &gitssh.PublicKeys{User: gitUserGit, Signer: signer}
		cloneOptions.Auth = auth
	}

	// Do a plain clone of the repo
	repo, err := git.PlainClone(repoDir, false, &cloneOptions)
	return repo, newCloneError(err)
}

// Fetch will clone a repository if necessary, then attempt to fetch
// the repository from origin
func (r *Resolver) Fetch(ctx *Context) error {
	repository, err := r.Clone(ctx)
	if err != nil {
		return err
	}

	fetchOptions := &git.FetchOptions{
		RemoteName: remoteNameOrigin,
	}
	// If an SSH key was supplied, try to use it.
	if ctx.Options.SSHKey != nil {
		signer, err := ssh.ParsePrivateKey([]byte(ctx.Options.SSHKey))
		if err != nil {
			return newFetchError(err)
		}

		auth := &gitssh.PublicKeys{User: gitUserGit, Signer: signer}
		fetchOptions.Auth = auth
	}

	err = repository.Fetch(fetchOptions)

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return newFetchError(err)
	}
	return nil
}

// GetCommit will parse the Ref (i.e. #<ref>) from the git uri and determine if its a branch/tag/commit.
// Returns the actual commit object for that ref. Defaults to HEAD.
// GetCommit will first fetch from origin.
func (r *Resolver) GetCommit(ctx *Context) (*gitplumbingobject.Commit, error) {
	err := r.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	repository, err := r.Clone(ctx)
	if err != nil {
		return nil, err
	}

	// If no ref is specified, default to HEAD.
	uriInfo := parseGitURI(ctx.URI)
	if uriInfo.Ref == "" {
		head, err := repository.Head()
		if err != nil {
			return nil, err
		}

		return repository.CommitObject(head.Hash())
	}

	// Otherwise figure out what type of reference it is.
	// First check if its a branch
	refName := gitplumbing.ReferenceName(fmt.Sprintf("%s:refs/remotes/origin", uriInfo.Ref))
	ref, _ := repository.Reference(refName, false)
	if ref != nil {
		return repository.CommitObject(ref.Hash())
	}

	// Next check if it's a tag
	refName = gitplumbing.ReferenceName(fmt.Sprintf("refs/tags/%s", uriInfo.Ref))
	ref, _ = repository.Reference(refName, false)
	if ref != nil {
		return repository.CommitObject(ref.Hash())
	}

	// Finally, if it's not a branch or a tag, just take it as a hash
	refHash := gitplumbing.NewHash(uriInfo.Ref)
	return repository.CommitObject(refHash)
}

// Checkout will clone and fetch, then attempt to check out the ref specified in the context.
func (r *Resolver) Checkout(ctx *Context) error {
	repository, err := r.Clone(ctx)
	if err != nil {
		return err
	}

	commit, err := r.GetCommit(ctx)
	if err != nil {
		return err
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}

	checkoutOpts := &git.CheckoutOptions{
		Hash: commit.Hash,
	}
	return worktree.Checkout(checkoutOpts)
}

// FileContents will clone, fetch, and checkout the proper reference, and if successful
// will attempt to return the contents of the file at fileName.
func (r *Resolver) FileContents(ctx *Context, fileName string) ([]byte, error) {
	commit, err := r.GetCommit(ctx)
	if err != nil {
		return nil, err
	}

	file, err := commit.File(fileName)
	if err != nil {
		return nil, err
	}

	reader, err := file.Reader()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(reader)
}

func (r *Resolver) GetRepositoryPath(ctx *Context) string {
	uriInfo := parseGitURI(ctx.URI)
	return path.Join(r.WorkDirectory, stripProtocol(uriInfo.CloneURI))
}

// GetTagNames will clone and fetch, and if successful will return the repository's tags.
func (r *Resolver) GetTagNames(ctx *Context) ([]string, error) {
	err := r.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	repository, err := r.Clone(ctx)
	if err != nil {
		return nil, err
	}

	tagRefs, err := repository.Tags()
	if err != nil {
		return nil, err
	}

	tags := []string{}
	err = tagRefs.ForEach(func(t *gitplumbing.Reference) error {
		tagNameParts := strings.Split(t.Name().String(), "/")
		tags = append(tags, tagNameParts[len(tagNameParts)-1])
		return nil
	})

	return tags, nil
}

type uriInfo struct {
	FullURI  string
	CloneURI string
	Ref      string
	RepoName string
}

func parseGitURI(gitURI string) uriInfo {
	partByRef := strings.Split(gitURI, "#")
	cloneURI := partByRef[0]

	// strip /.git from local repositories references
	if strings.HasSuffix(cloneURI, "/.git") {
		cloneURI = strings.Replace(cloneURI, "/.git", "", -1)
	}

	repoNameParts := strings.Split(cloneURI, "/")
	repoName := strings.Replace(repoNameParts[len(repoNameParts)-1], ".git", "", 1)

	var ref string
	if len(partByRef) == 2 {
		ref = partByRef[1]
	}
	return uriInfo{
		FullURI:  gitURI,
		CloneURI: cloneURI,
		RepoName: repoName,
		Ref:      ref,
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// stripProtocol strips out the protocol:// from uri
func stripProtocol(uri string) string {
	protocolParts := strings.Split(uri, "://")

	if len(protocolParts) > 1 {
		return protocolParts[1]
	} else {
		return uri
	}
}
