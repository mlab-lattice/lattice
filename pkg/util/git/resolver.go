package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
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
type Context struct {
	RepositoryURL string
	Options       *Options
}

type Reference struct {
	Commit *string
	Tag    *string
	Branch *string
}

// Options contains information about how to complete the operation.
type Options struct {
	SSHKey []byte
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
	// validate repo url
	if !IsValidRepositoryURI(ctx.RepositoryURL) {
		return nil, fmt.Errorf("bad git uri '%v'", ctx.RepositoryURL)
	}
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
		URL:      ctx.RepositoryURL,
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
func (r *Resolver) GetCommit(ctx *Context, ref *Reference) (*gitplumbingobject.Commit, error) {
	err := r.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	repository, err := r.Clone(ctx)
	if err != nil {
		return nil, err
	}

	var hash gitplumbing.Hash
	switch {
	case ref.Commit != nil:
		hash = gitplumbing.NewHash(*ref.Commit)

	case ref.Branch != nil:
		refName := gitplumbing.ReferenceName(fmt.Sprintf("%s:refs/remotes/origin", *ref.Branch))
		gitRef, _ := repository.Reference(refName, false)
		if gitRef == nil {
			return nil, fmt.Errorf("invalid branch name %v", *ref.Branch)
		}

		hash = gitRef.Hash()

	case ref.Tag != nil:
		refName := gitplumbing.ReferenceName(fmt.Sprintf("refs/tags/%s", *ref.Tag))
		gitRef, _ := repository.Reference(refName, false)
		if gitRef == nil {
			return nil, fmt.Errorf("invalid tag name %v", *ref.Tag)
		}

		hash = gitRef.Hash()

	default:
		head, err := repository.Head()
		if err != nil {
			return nil, err
		}

		hash = head.Hash()
	}

	return repository.CommitObject(hash)
}

// Checkout will clone and fetch, then attempt to check out the ref specified in the context.
func (r *Resolver) Checkout(ctx *Context, ref *Reference) error {
	repository, err := r.Clone(ctx)
	if err != nil {
		return err
	}

	commit, err := r.GetCommit(ctx, ref)
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
func (r *Resolver) FileContents(ctx *Context, ref *Reference, fileName string) ([]byte, error) {
	commit, err := r.GetCommit(ctx, ref)
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
	return path.Join(r.WorkDirectory, stripProtocol(ctx.RepositoryURL))
}

// Tags will clone and fetch, and if successful will return the repository's tags (annotated + light-weight).
func (r *Resolver) Tags(ctx *Context) ([]string, error) {
	err := r.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	repository, err := r.Clone(ctx)
	if err != nil {
		return nil, err
	}

	// list all tags! (annotated + light-weight)
	tagRefs, err := repository.Tags()
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0)
	err = tagRefs.ForEach(func(t *gitplumbing.Reference) error {
		tagNameParts := strings.Split(t.Name().String(), "/")
		tags = append(tags, tagNameParts[len(tagNameParts)-1])
		return nil
	})

	return tags, nil
}

// regex for matching git repo urls
var gitRepositoryURIRegex = regexp.MustCompile(`((?:git|file|ssh|https?|git@[-\w.]+)):(//)?(.*.git)(#(([-\d\w._])+)?)?$`)

// IsValidRepositoryURI
func IsValidRepositoryURI(uri string) bool {
	return gitRepositoryURIRegex.MatchString(uri)
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
	}

	return uri
}
