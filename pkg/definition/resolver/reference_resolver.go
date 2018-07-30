package resolver

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/template"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	gitplumbingobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

const (
	fileExtensionJSON = ".json"
	fileExtensionYAML = ".yaml"
	fileExtensionYML  = ".yml"
)

// ReferenceResolver resolves references into fully hydrated (i.e. the template
// engine has already acted upon it) component.
type ReferenceResolver interface {
	Versions(repository string, semverRange semver.Range) ([]string, error)
	// ResolveReference resolves the reference.
	ResolveReference(
		systemID v1.SystemID,
		path tree.NodePath,
		ctx *git.FileReference,
		ref *definitionv1.Reference,
	) (component.Interface, *git.FileReference, error)
}

// DefaultReferenceResolver fulfils the ReferenceResolver interface.
type DefaultReferenceResolver struct {
	gitResolver *git.Resolver
	store       TemplateStore
}

// NewReferenceResolver returns a ReferenceResolver that uses workDirectory for
// scratch space such as cloning git repositories.
func NewReferenceResolver(workDirectory string, store TemplateStore) (ReferenceResolver, error) {
	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	gitResolver, err := git.NewResolver(workDirectory + "/git")
	if err != nil {
		return nil, err
	}

	r := &DefaultReferenceResolver{
		gitResolver: gitResolver,
		store:       store,
	}
	return r, nil
}

func (r *DefaultReferenceResolver) Versions(repository string, semverRange semver.Range) ([]string, error) {
	ctx := &git.Context{
		RepositoryURL: repository,
		Options:       &git.Options{},
	}
	tags, err := r.gitResolver.Tags(ctx)
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

// ResolveReference fulfils the ReferenceResolver interface.
func (r *DefaultReferenceResolver) ResolveReference(
	systemID v1.SystemID,
	path tree.NodePath,
	ctx *git.FileReference,
	ref *definitionv1.Reference,
) (component.Interface, *git.FileReference, error) {
	// retrieve the template and its commit context
	t, resolvedCxt, err := r.resolveTemplate(systemID, path, ctx, ref)
	if err != nil {
		return nil, nil, err
	}

	// evaluate the template with the reference's parameters
	result, err := t.Evaluate(path, ref.Parameters)
	if err != nil {
		return nil, nil, err
	}

	// create a new component from the evaluated template
	c, err := NewComponent(result)
	if err != nil {
		return nil, nil, err
	}

	return c, resolvedCxt, nil
}

func (r *DefaultReferenceResolver) resolveTemplate(
	systemID v1.SystemID,
	path tree.NodePath,
	ctx *git.FileReference,
	ref *definitionv1.Reference,
) (*template.Template, *git.FileReference, error) {
	gitCtx := &git.Context{
		Options: &git.Options{},
	}

	// Get the proper commit reference and file for the reference, potentially updating
	// the context as well.
	var gitRef *git.Reference
	var file string
	switch {
	case ref.GitRepository != nil && ref.File != nil:
		return nil, nil, fmt.Errorf("reference cannot have both git_repository and file")

	case ref.GitRepository != nil:
		// if the reference is to a git_repository, resolve the commit for the reference,
		// and set the context to the repo referenced
		gitCtx.RepositoryURL = ref.GitRepository.URL
		file = ref.GitRepository.File

		commit, err := r.gitReferenceCommit(ref.GitRepository)
		if err != nil {
			return nil, nil, err
		}

		commitHash := commit.Hash.String()
		gitRef = &git.Reference{Commit: &commitHash}

	case ref.File != nil:
		// if the reference is to a file, use the given context as the context, and set the
		// file to file referenced.
		gitCtx.RepositoryURL = ctx.RepositoryURL
		gitRef = &git.Reference{Commit: &ctx.Commit}
		file = *ref.File
	}

	fileRef := &git.FileReference{
		RepositoryURL: gitCtx.RepositoryURL,
		Commit:        *gitRef.Commit,
		File:          file,
	}

	// see if we already have this commit from this repository in the template store.
	t, err := r.store.Get(systemID, fileRef)
	if err != nil {
		// if there was an error getting the cached version, get the template from the
		// repo
		t, err = r.resolveGitTemplate(gitCtx, gitRef, file)
		if err != nil {
			return nil, nil, err
		}

		// put the template into the template store
		if err = r.store.Put(systemID, fileRef, t); err != nil {
			return nil, nil, err
		}
	}

	// return the template that we found either from the store or from the repository
	// as well as the commit reference that was used to find the template
	return t, fileRef, nil
}

func (r *DefaultReferenceResolver) gitReferenceCommit(ref *definitionv1.GitRepositoryReference) (*gitplumbingobject.Commit, error) {
	ctx := &git.Context{
		RepositoryURL: ref.URL,
		Options:       &git.Options{},
	}

	var gitRef *git.Reference
	switch {
	case ref.Commit != nil:
		gitRef = &git.Reference{Commit: ref.Commit}

	case ref.Branch != nil:
		gitRef = &git.Reference{Branch: ref.Branch}

	case ref.Tag != nil:
		rng, err := semver.ParseRange(*ref.Tag)

		// If the tag is not a semver range, just use the tag
		if err != nil {
			gitRef = &git.Reference{Tag: ref.Tag}
			break
		}

		versions, err := r.Versions(ref.URL, rng)
		if err != nil {
			return nil, err
		}

		tag := versions[len(versions)-1]
		gitRef = &git.Reference{Tag: &tag}

	default:
		return nil, fmt.Errorf("git_repository reference must contain commit, branch, or tag")
	}

	return r.gitResolver.GetCommit(ctx, gitRef)
}

func (r *DefaultReferenceResolver) resolveGitTemplate(
	ctx *git.Context,
	ref *git.Reference,
	filePath string,
) (*template.Template, error) {
	data, err := r.gitResolver.FileContents(ctx, ref, filePath)
	if err != nil {
		return nil, err
	}

	var t template.Template
	switch e := filepath.Ext(filePath); e {
	case fileExtensionJSON:
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, err
		}

	case fileExtensionYAML, fileExtensionYML:
		if err := yaml.Unmarshal(data, &t); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("invalid file extension %v", e)
	}

	return &t, nil
}
