package resolver

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/template"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	gitplumbingobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

const (
	DepthInfinite = int32(-1)

	fileExtensionJSON = ".json"
	fileExtensionYAML = ".yaml"
	fileExtensionYML  = ".yml"
)

// ComponentResolver resolves references into fully hydrated (i.e. the template
// engine has already acted upon it) component.
type ComponentResolver interface {
	Versions(repository string, semverRange semver.Range) ([]string, error)
	// ResolveReference resolves the reference.
	ResolveReference(
		systemID v1.SystemID,
		path tree.NodePath,
		ctx *git.FileReference,
		ref *definitionv1.Reference,
		depth int32,
	) (component.Interface, *git.FileReference, error)
}

// DefaultComponentResolver fulfils the ComponentResolver interface.
type DefaultComponentResolver struct {
	gitResolver *git.Resolver
	store       TemplateStore
}

// NewComponentResolver returns a ComponentResolver that uses workDirectory for
// scratch space such as cloning git repositories.
func NewComponentResolver(workDirectory string, allowLocalRepos bool, store TemplateStore) (ComponentResolver, error) {
	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	gitResolver, err := git.NewResolver(workDirectory+"/git", allowLocalRepos)
	if err != nil {
		return nil, err
	}

	r := &DefaultComponentResolver{
		gitResolver: gitResolver,
		store:       store,
	}
	return r, nil
}

func (r *DefaultComponentResolver) Versions(repository string, semverRange semver.Range) ([]string, error) {
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

// ResolveReference fulfils the ComponentResolver interface.
func (r *DefaultComponentResolver) ResolveReference(
	systemID v1.SystemID,
	path tree.NodePath,
	ctx *git.FileReference,
	ref *definitionv1.Reference,
	depth int32,
) (component.Interface, *git.FileReference, error) {
	if depth == 0 {
		return ref, ctx, nil
	}

	if depth < DepthInfinite {
		return nil, nil, fmt.Errorf("invalid depth: %v", depth)
	}

	nextDepth := DepthInfinite
	if depth > 0 {
		nextDepth = depth - 1
	}

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

	// If the reference resolved to another reference, resolve that reference.
	// FIXME(kevinrosendahl): detect cycles
	if resolvedRef, ok := c.(*definitionv1.Reference); ok {
		return r.ResolveReference(systemID, path, resolvedCxt, resolvedRef, nextDepth)
	}

	// If the reference resolved to a systemID, resolve the system's components.
	if system, ok := c.(*definitionv1.System); ok {
		system, err := r.resolveSystemComponents(systemID, path, resolvedCxt, system, nextDepth)
		return system, resolvedCxt, err
	}

	return c, resolvedCxt, nil
}

func (r *DefaultComponentResolver) resolveTemplate(
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

func (r *DefaultComponentResolver) gitReferenceCommit(ref *definitionv1.GitRepositoryReference) (*gitplumbingobject.Commit, error) {
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

		if len(versions) == 0 {
			return nil, fmt.Errorf("no tags match the requested version")
		}

		tag := versions[len(versions)-1]
		gitRef = &git.Reference{Tag: &tag}

	default:
		return nil, fmt.Errorf("git_repository reference must contain commit, branch, or tag")
	}

	return r.gitResolver.GetCommit(ctx, gitRef)
}

func (r *DefaultComponentResolver) resolveGitTemplate(
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

func (r *DefaultComponentResolver) resolveSystemComponents(
	systemID v1.SystemID,
	path tree.NodePath,
	ctx *git.FileReference,
	system *definitionv1.System,
	depth int32,
) (*definitionv1.System, error) {
	// Loop through each of the components.
	//  - If the component is a system, recursively resolve its components.
	//  - If the component is a reference, resolve it (potentially also recursively resolving
	//    system components if the reference was to a system).
	for name, c := range system.Components {
		childPath := path.Child(name)
		switch typedComponent := c.(type) {

		case *definitionv1.System:
			// If the component is a system, recursively resolve the system and overwrite it in the components map
			subSystem, err := r.resolveSystemComponents(systemID, childPath, ctx, typedComponent, depth)
			if err != nil {
				return nil, err
			}

			system.Components[name] = subSystem

		case *definitionv1.Reference:
			// If the component is a reference, resolve the reference.
			resolved, _, err := r.ResolveReference(systemID, childPath, ctx, typedComponent, depth)
			if err != nil {
				return nil, err
			}

			system.Components[name] = resolved
		}
	}

	return system, nil
}
