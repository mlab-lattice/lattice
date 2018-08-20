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

	DefaultFile = "lattice.yaml"
)

// ResolutionResult contains the component as well as information about the resolution
// of the component and its subcomponents.
type ResolutionResult struct {
	Component component.Interface
	Info      ResolutionInfo
}

// ResolutionInfo maps paths to information about their resolution.
type ResolutionInfo map[tree.Path]ResolutionNodeInfo

// ResolutionNodeInfo contains information about the resolution of a subcomponent.
type ResolutionNodeInfo struct {
	Commit git.CommitReference `json:"commit"`
}

// ComponentResolver resolves references into fully hydrated (i.e. the template
// engine has already acted upon it) component.
type ComponentResolver interface {
	Versions(repository string, semverRange semver.Range) ([]string, error)
	// ResolveReference resolves the reference.
	ResolveReference(
		systemID v1.SystemID,
		path tree.Path,
		ctx *git.FileReference,
		ref *definitionv1.Reference,
		depth int32,
	) (*ResolutionResult, error)
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
	return r.gitResolver.Versions(ctx, semverRange)
}

// ResolveReference fulfils the ComponentResolver interface.
func (r *DefaultComponentResolver) ResolveReference(
	systemID v1.SystemID,
	path tree.Path,
	ctx *git.FileReference,
	ref *definitionv1.Reference,
	depth int32,
) (*ResolutionResult, error) {
	info := make(ResolutionInfo)
	c, err := r.resolveReference(systemID, path, ctx, ref, depth, info)
	if err != nil {
		return nil, err
	}

	rr := &ResolutionResult{
		Component: c,
		Info:      info,
	}
	return rr, nil
}

func (r *DefaultComponentResolver) resolveReference(
	systemID v1.SystemID,
	path tree.Path,
	ctx *git.FileReference,
	ref *definitionv1.Reference,
	depth int32,
	info ResolutionInfo,
) (component.Interface, error) {
	if depth == 0 {
		info[path] = ResolutionNodeInfo{Commit: ctx.CommitReference}
		return ref, nil
	}

	if depth < DepthInfinite {
		return nil, fmt.Errorf("invalid depth: %v", depth)
	}

	nextDepth := DepthInfinite
	if depth > 0 {
		nextDepth = depth - 1
	}

	// retrieve the template and its commit context
	t, resolvedCxt, err := r.resolveTemplate(systemID, path, ctx, ref)
	if err != nil {
		return nil, err
	}

	p, err := r.hydrateReferenceParameters(path, ref.Parameters)
	if err != nil {
		return nil, err
	}

	// evaluate the template with the reference's parameters
	result, err := t.Evaluate(path, p)
	if err != nil {
		return nil, err
	}

	// create a new component from the evaluated template
	c, err := NewComponent(result)
	if err != nil {
		return nil, err
	}

	// If the reference resolved to another reference, resolve that reference.
	// FIXME(kevinrosendahl): detect cycles
	if resolvedRef, ok := c.(*definitionv1.Reference); ok {
		return r.resolveReference(systemID, path, resolvedCxt, resolvedRef, nextDepth, info)
	}

	// If the reference resolved to a systemID, resolve the system's components.
	if system, ok := c.(*definitionv1.System); ok {
		c, err = r.resolveSystemComponents(systemID, path, resolvedCxt, system, nextDepth, info)
		if err != nil {
			return nil, err
		}
	}

	info[path] = ResolutionNodeInfo{Commit: resolvedCxt.CommitReference}
	return c, nil
}

func (r *DefaultComponentResolver) resolveTemplate(
	systemID v1.SystemID,
	path tree.Path,
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
		file = DefaultFile
		if ref.GitRepository.File != nil {
			file = *ref.GitRepository.File
		}

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
		CommitReference: git.CommitReference{
			RepositoryURL: gitCtx.RepositoryURL,
			Commit:        *gitRef.Commit,
		},
		File: file,
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

	gitRef := &git.Reference{
		Commit:  ref.Commit,
		Branch:  ref.Branch,
		Tag:     ref.Tag,
		Version: ref.Version,
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

// FIXME(kevindrosendahl): this is pretty gross
func (r *DefaultComponentResolver) hydrateReferenceParameters(
	path tree.Path,
	parameters map[string]interface{},
) (map[string]interface{}, error) {
	p := make(map[string]interface{})

	// look for any secret parameters
	for k, v := range parameters {
		p[k] = v

		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		s, ok := m[template.SecretParameterLVal]
		if !ok {
			continue
		}

		ss, ok := s.(string)
		if !ok {
			return nil, fmt.Errorf("expected secret value to be a string for parameter %v", k)
		}

		// path is the tree.Path of the reference being resolved. if there's a secret being passed
		// down as a parameter, that means that it is the secret of the component which has the
		// secret, i.e. the parent of the the path that is passed in
		parent, err := path.Parent()
		if err != nil {
			return nil, fmt.Errorf("got error creating secret reference for parameter %v: %v", k, err)
		}

		sp, err := tree.NewPathSubcomponentFromParts(parent, ss)
		if err != nil {
			return nil, fmt.Errorf("got error creating secret reference for parameter %v: %v", k, err)
		}

		p[k] = &definitionv1.SecretRef{Value: sp}
	}

	return p, nil
}

func (r *DefaultComponentResolver) resolveSystemComponents(
	systemID v1.SystemID,
	path tree.Path,
	ctx *git.FileReference,
	system *definitionv1.System,
	depth int32,
	info ResolutionInfo,
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
			subSystem, err := r.resolveSystemComponents(systemID, childPath, ctx, typedComponent, depth, info)
			if err != nil {
				return nil, err
			}

			system.Components[name] = subSystem

		case *definitionv1.Reference:
			// If the component is a reference, resolve the reference.
			resolved, err := r.resolveReference(systemID, childPath, ctx, typedComponent, depth, info)
			if err != nil {
				return nil, err
			}

			system.Components[name] = resolved

		default:
			info[childPath] = ResolutionNodeInfo{Commit: ctx.CommitReference}
		}
	}

	return system, nil
}
