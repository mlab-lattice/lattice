package resolver

import (
	"fmt"
	"path/filepath"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"github.com/blang/semver"
)

// ReferenceResolver resolves references into fully hydrated (i.e. the template
// engine has already acted upon it) component.
type ReferenceResolver interface {
	// ResolveReference resolves the reference.
	ResolveReference(ctx *definitionv1.GitRepositoryReference, ref *definitionv1.Reference) (component.Interface, error)
}

// DefaultReferenceResolver fulfils the ReferenceResolver interface.
type DefaultReferenceResolver struct {
	gitResolver *git.Resolver
}

// NewReferenceResolver returns a ReferenceResolver that uses workDirectory for
// scratch space such as cloning git repositories.
func NewReferenceResolver(workDirectory string) (ReferenceResolver, error) {
	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	gitResolver, err := git.NewResolver(workDirectory + "/git")
	if err != nil {
		return nil, err
	}

	r := &DefaultReferenceResolver{
		gitResolver: gitResolver,
	}
	return r, nil
}

// ResolveReference fulfils the ReferenceResolver interface.
func (r *DefaultReferenceResolver) ResolveReference(
	ctx *definitionv1.GitRepositoryReference,
	ref *definitionv1.Reference,
) (component.Interface, error) {
	if ref.GitRepository != nil {
		return r.resolveGitReference(ref.GitRepository)
	}

	if ref.File != nil {
		return r.resolveFileReference(ctx, *ref.File)
	}

	return nil, fmt.Errorf("reference must contain either git_repository or file")
}

func (r *DefaultReferenceResolver) resolveGitReference(repository *definitionv1.GitRepositoryReference) (component.Interface, error) {
	ctx := &git.Context{
		RepositoryURL: repository.URL,
		Options:       &git.Options{},
	}

	var ref *git.Reference
	switch {
	case repository.Commit != nil:
		ref = &git.Reference{Commit: repository.Commit}

	case repository.Branch != nil:
		ref = &git.Reference{Branch: repository.Branch}

	case repository.Tag != nil:
		rng, err := semver.ParseRange(*repository.Tag)

		// If the tag is not a semver range, just use the tag
		if err != nil {
			ref = &git.Reference{Tag: repository.Tag}
			break
		}

		// Otherwise, get the tags and find the largest tag that satisfies the range constraint.
		tags, err := r.gitResolver.Tags(ctx)
		if err != nil {
			return nil, err
		}

		// We allow tags such as v1.2.3 which technically don't adhere to the semver
		// spec (which expects 1.2.3), but when we parse it into a semver.Version
		// we'll lose information about what the git tag actually is, so map the
		// semver.Versions to their git tags.
		gitTags := make(map[string]string)
		var versions []semver.Version
		for _, tag := range tags {
			v, err := semver.ParseTolerant(tag)
			if err != nil {
				continue
			}

			if rng(v) {
				gitTags[v.String()] = tag
				versions = append(versions, v)
			}
		}

		if len(versions) == 0 {
			return nil, fmt.Errorf("no tags match %v", *repository.Tag)
		}

		semver.Sort(versions)
		largestVersion := versions[len(versions)-1]
		tag := gitTags[largestVersion.String()]
		ref = &git.Reference{Tag: &tag}

	default:
		return nil, fmt.Errorf("git_repository reference must contain commit, branch, or tag")
	}

	data, err := r.gitResolver.FileContents(ctx, ref, repository.File)
	if err != nil {
		return nil, err
	}

	return definitionv1.NewComponentFromJSON(data)
}

func (r *DefaultReferenceResolver) resolveFileReference(
	ctx *definitionv1.GitRepositoryReference,
	file string,
) (component.Interface, error) {
	if ctx.GitRepository.Commit == nil {
		return nil, fmt.Errorf("file reference context git_repository must have a commit")
	}

	filePath := filepath.Join(filepath.Dir(ctx.File), file)

	gitCtx := &git.Context{
		RepositoryURL: ctx.URL,
		Options:       &git.Options{},
	}
	ref := &git.Reference{Commit: ctx.GitRepository.Commit}
	data, err := r.gitResolver.FileContents(gitCtx, ref, filePath)
	if err != nil {
		return nil, err
	}

	return definitionv1.NewComponentFromJSON(data)
}
