package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"github.com/blang/semver"
)

type ReferenceResolver interface {
	Resolve(*Reference) (component.Interface, error)
}

type DefaultReferenceResolver struct {
	gitResolver *git.Resolver
}

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

func (r *DefaultReferenceResolver) Resolve(ref *Reference) (component.Interface, error) {
	if ref.GitRepository != nil {
		return r.resolveGitReference(ref.GitRepository)
	}

	if ref.File != nil {
		return r.resolveFileReference(*ref.File)
	}

	return nil, fmt.Errorf("reference must contain either git_repository or file")
}

func (r *DefaultReferenceResolver) resolveGitReference(repository *GitRepositoryReference) (component.Interface, error) {
	ctx := &git.Context{
		RepositoryURL: repository.URL,
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

	return NewComponentFromJSON(data)
}

func (r *DefaultReferenceResolver) resolveFileReference(file string) (component.Interface, error) {
	return nil, nil
}
