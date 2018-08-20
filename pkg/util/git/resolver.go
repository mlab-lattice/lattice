package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/blang/semver"
	"gopkg.in/src-d/go-git.v4"
	gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
	gitplumbingobject "gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	gitUserGit       = "git"
	remoteNameOrigin = "origin"
)

// Resolver provides utility methods for manipulating git repositories
// under a specific working directory on the filesystem.
type Resolver struct {
	workDirectory   string
	allowLocalRepos bool
}

// Context contains information about the current operation being invoked.
type Context struct {
	RepositoryURL string
	Options       *Options
}

// Reference is a union type containing a reference to a commit, tag, or branch.
type Reference struct {
	Commit  *string
	Branch  *string
	Tag     *string
	Version *string
}

func (r *Reference) Validate() error {
	if r.Commit == nil && r.Branch == nil && r.Tag == nil && r.Version == nil {
		return fmt.Errorf("reference must contain a commit, branch, tag, or version")
	}

	if (r.Commit != nil && (r.Tag != nil || r.Branch != nil || r.Version != nil)) ||
		(r.Tag != nil && (r.Branch != nil || r.Version != nil)) ||
		(r.Branch != nil && r.Version != nil) {
		return fmt.Errorf("reference can only contain a single commit, branch, tag, or version")
	}

	return nil
}

type CommitReference struct {
	RepositoryURL string `json:"repositoryUrl"`
	Commit        string `json:"commit"`
}

type FileReference struct {
	CommitReference
	File string `json:"file"`
}

// Options contains information about how to complete the operation.
type Options struct {
	SSHKey []byte
}

func NewResolver(workDirectory string, allowLocalRepos bool) (*Resolver, error) {
	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	err := os.MkdirAll(workDirectory, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to create git resolver work directory: %v", err)
	}

	sr := &Resolver{
		workDirectory:   workDirectory,
		allowLocalRepos: allowLocalRepos,
	}
	return sr, nil
}

// Clone will  open the repository and return it If the repository specified in the Context has already been cloned,
// otherwise it will attempt to clone the repository and on success return the cloned repository
func (r *Resolver) Clone(ctx *Context) (*git.Repository, error) {
	// validate repo url
	if !r.IsValidRepositoryURI(ctx.RepositoryURL) {
		return nil, fmt.Errorf("bad git uri '%v'", ctx.RepositoryURL)
	}
	repoDir := r.RepositoryPath(ctx.RepositoryURL)

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

// Commit will parse the Ref (i.e. #<ref>) from the git uri and determine if its a branch/tag/commit.
// Returns the actual commit object for that ref. Defaults to HEAD.
// GetCommit will first fetch from origin.
func (r *Resolver) GetCommit(ctx *Context, ref *Reference) (*gitplumbingobject.Commit, error) {
	if err := ref.Validate(); err != nil {
		return nil, err
	}

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
		refName := gitplumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", *ref.Branch))
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

	case ref.Version != nil:
		rng, err := semver.ParseRange(*ref.Version)

		// If the tag is not a semver range, just use the tag
		if err != nil {
			return nil, fmt.Errorf("version is not a valid semver range")
		}

		versions, err := r.Versions(ctx, rng)
		if err != nil {
			return nil, err
		}

		if len(versions) == 0 {
			return nil, fmt.Errorf("no tags match the requested version")
		}

		tag := versions[len(versions)-1]
		ref = &Reference{Tag: &tag}
		return r.GetCommit(ctx, ref)
	}

	return repository.CommitObject(hash)
}

func (r *Resolver) Versions(ctx *Context, semverRange semver.Range) ([]string, error) {
	tags, err := r.Tags(ctx)
	if err != nil {
		return nil, err
	}

	var versions []semver.Version
	for _, tag := range tags {
		v, err := semver.Parse(tag)
		if err != nil {
			continue
		}

		// If a semver range was passed in, check to see if the version
		// matches the range.
		if semverRange != nil && !semverRange(v) {
			continue
		}
		versions = append(versions, v)
	}

	semver.Sort(versions)
	var v []string
	for _, version := range versions {
		v = append(v, version.String())
	}
	return v, nil
}

// Checkout will clone and fetch, then attempt to check out the ref specified in the context.
func (r *Resolver) Checkout(ctx *Context, ref *Reference) error {
	if err := ref.Validate(); err != nil {
		return err
	}

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
	if err := ref.Validate(); err != nil {
		return nil, err
	}

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

func (r *Resolver) RepositoryPath(url string) string {
	return path.Join(r.workDirectory, stripProtocol(url))
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

// IsValidRepositoryURI returns a boolean indicating whether or not the
// uri is a valid git repository.
func (r *Resolver) IsValidRepositoryURI(uri string) bool {
	e, err := transport.NewEndpoint(uri)
	if err != nil {
		return false
	}

	if !r.allowLocalRepos {
		return e.Protocol() != "file"
	}

	return true
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
