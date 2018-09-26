package resolver

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver/template"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"encoding/json"
	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	gitplumbingobject "gopkg.in/src-d/go-git.v4/plumbing/object"
	"path/filepath"
)

func newV1ComponentResolver(
	resolver *componentResolver,
	gitResolver *git.Resolver,
	templateStore TemplateStore,
	secretStore SecretStore,
) *v1ComponentResolver {
	return &v1ComponentResolver{
		resolver,
		gitResolver,
		templateStore,
		secretStore,
	}
}

type v1ComponentResolver struct {
	resolver *componentResolver

	gitResolver   *git.Resolver
	templateStore TemplateStore
	secretStore   SecretStore
}

func (r *v1ComponentResolver) Resolve(
	c definition.Component,
	id v1.SystemID,
	path tree.Path,
	ctx *resolutionContext,
	depth int,
	result *ResolutionTree,
) error {
	switch v1Component := c.(type) {
	case *definitionv1.Reference:
		return r.resolveReference(v1Component, id, path, ctx, depth, result)

	case *definitionv1.System:
		return r.resolveSystem(v1Component, id, path, ctx, depth, result)

	default:
		info := &ResolutionInfo{
			Component:    c,
			Commit:       ctx.CommitReference,
			SSHKeySecret: ctx.SSHKeySecret,
		}
		result.Insert(path, info)
		return nil
	}
}

func (r *v1ComponentResolver) Versions(
	ref *definitionv1.Reference,
	rng semver.Range,
) ([]string, error) {
	if ref.GitRepository == nil {
		return nil, fmt.Errorf("cannot get versions of %v without git_repository", ref.Type().String())
	}

	ctx := &git.Context{
		RepositoryURL: ref.GitRepository.URL,
		Options:       &git.Options{},
	}

	return r.gitResolver.Versions(ctx, rng)
}

func (r *v1ComponentResolver) resolveReference(
	ref *definitionv1.Reference,
	id v1.SystemID,
	path tree.Path,
	ctx *resolutionContext,
	depth int,
	result *ResolutionTree,
) error {
	if depth == 0 {
		info := &ResolutionInfo{
			Component:    ref,
			Commit:       ctx.CommitReference,
			SSHKeySecret: ctx.SSHKeySecret,
		}
		result.Insert(path, info)
		return nil
	}

	if depth < DepthInfinite {
		return fmt.Errorf("invalid depth: %v", depth)
	}

	nextDepth := DepthInfinite
	if depth > 0 {
		nextDepth = depth - 1
	}

	// retrieve the template and its commit context
	t, resolvedCxt, err := r.resolveTemplate(id, path, ctx, ref)
	if err != nil {
		return err
	}

	p, err := r.hydrateReferenceParameters(path, ref.Parameters)
	if err != nil {
		return err
	}

	// evaluate the template with the reference's parameters
	evaluated, err := t.Evaluate(path, p)
	if err != nil {
		return err
	}

	// create a new component from the evaluated template
	c, err := r.resolver.newComponent(evaluated)
	if err != nil {
		return err
	}

	return r.resolver.resolve(c, id, path, resolvedCxt, nextDepth, result)
}

func (r *v1ComponentResolver) resolveSystem(
	system *definitionv1.System,
	id v1.SystemID,
	path tree.Path,
	ctx *resolutionContext,
	depth int,
	result *ResolutionTree,
) error {
	info := &ResolutionInfo{
		Component:    system,
		Commit:       ctx.CommitReference,
		SSHKeySecret: ctx.SSHKeySecret,
	}
	result.Insert(path, info)

	for name, c := range system.Components {
		err := r.resolver.resolve(c, id, path.Child(name), ctx, depth, result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *v1ComponentResolver) resolveTemplate(
	systemID v1.SystemID,
	path tree.Path,
	ctx *resolutionContext,
	ref *definitionv1.Reference,
) (*template.Template, *resolutionContext, error) {
	gitCtx := &git.Context{
		Options: &git.Options{},
	}

	// Get the proper commit reference and file for the reference, potentially updating
	// the context as well.
	var sshKeySecret *tree.PathSubcomponent
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

		var sshKey []byte
		if ref.GitRepository.SSHKey != nil {
			sshKeySecret = &ref.GitRepository.SSHKey.Value
			sshKeyVal, err := r.secretStore.Get(systemID, ref.GitRepository.SSHKey.Value)
			if err != nil {
				return nil, nil, err
			}

			sshKey = []byte(sshKeyVal)
		}

		commit, err := r.gitReferenceCommit(systemID, ref.GitRepository, sshKey)
		if err != nil {
			return nil, nil, err
		}

		commitHash := commit.Hash.String()
		gitRef = &git.Reference{Commit: &commitHash}
		gitCtx.Options.SSHKey = sshKey

	case ref.File != nil:
		// if the reference is to a file, use the given context as the context, and set the
		// file to file referenced.
		gitCtx.RepositoryURL = ctx.CommitReference.RepositoryURL
		gitCtx.Options.SSHKey = ctx.SSHKey
		gitRef = &git.Reference{Commit: &ctx.CommitReference.Commit}
		file = *ref.File
	}

	commitRef := &git.CommitReference{
		RepositoryURL: gitCtx.RepositoryURL,
		Commit:        *gitRef.Commit,
	}

	fileRef := &git.FileReference{
		CommitReference: *commitRef,
		File:            file,
	}

	resolvedContext := &resolutionContext{
		CommitReference: commitRef,
		SSHKey:          gitCtx.Options.SSHKey,
		SSHKeySecret:    sshKeySecret,
	}

	// Only want to check the cache if no credentials are required.
	// See https://github.com/mlab-lattice/lattice/issues/195 for more info
	checkCache := resolvedContext.SSHKey == nil

	if checkCache {
		// see if we already have this commit from this repository in the template store.
		t, err := r.templateStore.Get(systemID, fileRef)
		if err == nil {
			return t, resolvedContext, nil
		}
	}

	// if there was an error getting the cached version, get the template from the
	// repo
	t, err := r.resolveGitTemplate(gitCtx, gitRef, file)
	if err != nil {
		return nil, nil, err
	}

	if checkCache {
		// put the template into the template store
		if err = r.templateStore.Put(systemID, fileRef, t); err != nil {
			return nil, nil, err
		}
	}

	// return the template that we found either from the store or from the repository
	// as well as the commit reference that was used to find the template
	return t, resolvedContext, nil
}

func (r *v1ComponentResolver) gitReferenceCommit(
	systemID v1.SystemID,
	ref *definitionv1.GitRepositoryReference,
	sshKey []byte,
) (*gitplumbingobject.Commit, error) {
	ctx := &git.Context{
		RepositoryURL: ref.URL,
		Options: &git.Options{
			SSHKey: sshKey,
		},
	}

	gitRef := &git.Reference{
		Commit:  ref.Commit,
		Branch:  ref.Branch,
		Tag:     ref.Tag,
		Version: ref.Version,
	}

	return r.gitResolver.GetCommit(ctx, gitRef)
}

func (r *v1ComponentResolver) resolveGitTemplate(
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
func (r *v1ComponentResolver) hydrateReferenceParameters(
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
