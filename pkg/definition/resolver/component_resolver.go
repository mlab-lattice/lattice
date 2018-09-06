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
	// Component contains the hydrated but unresolved version of the component.
	// That is, if the component is a system, it may contain unresolved references.
	Component    component.Interface
	Commit       git.CommitReference
	SSHKeySecret *tree.PathSubcomponent
}

// TODO(kevindrosendahl): seems a little leaky that this is here
func (r *ResolutionNodeInfo) MarshalJSON() ([]byte, error) {
	componentBytes, err := json.Marshal(r.Component)
	if err != nil {
		return nil, err
	}

	d := resolutionNodeInfoDecoder{
		Component:    componentBytes,
		Commit:       r.Commit,
		SSHKeySecret: r.SSHKeySecret,
	}
	return json.Marshal(&d)
}

func (r *ResolutionNodeInfo) UnmarshalJSON(data []byte) error {
	var d resolutionNodeInfoDecoder
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	c, err := definitionv1.NewComponentFromJSON(d.Component)
	if err != nil {
		return err
	}

	rni := ResolutionNodeInfo{
		Component:    c,
		Commit:       d.Commit,
		SSHKeySecret: d.SSHKeySecret,
	}
	*r = rni

	return nil
}

type resolutionNodeInfoDecoder struct {
	Component    json.RawMessage        `json:"component"`
	Commit       git.CommitReference    `json:"commit"`
	SSHKeySecret *tree.PathSubcomponent `json:"sshKeySecret,omitempty"`
}

type resolutionContext struct {
	FileReference *git.FileReference
	SSHKeySecret  *tree.PathSubcomponent
	SSHKey        []byte
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
	ResolveComponentReferences(
		c component.Interface,
		systemID v1.SystemID,
		path tree.Path,
		ctx *git.FileReference,
		depth int32,
	) (*ResolutionResult, error)
}

// DefaultComponentResolver fulfils the ComponentResolver interface.
type DefaultComponentResolver struct {
	gitResolver   *git.Resolver
	templateStore TemplateStore
	secretStore   SecretStore
}

// NewComponentResolver returns a ComponentResolver that uses workDirectory for
// scratch space such as cloning git repositories.
func NewComponentResolver(
	workDirectory string,
	allowLocalRepos bool,
	templateStore TemplateStore,
	secretStore SecretStore,
) (ComponentResolver, error) {
	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	gitResolver, err := git.NewResolver(workDirectory+"/git", allowLocalRepos)
	if err != nil {
		return nil, err
	}

	r := &DefaultComponentResolver{
		gitResolver:   gitResolver,
		templateStore: templateStore,
		secretStore:   secretStore,
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
	// TODO(kevindrosendahl): this here is why private system definitions aren't supported
	rctx := &resolutionContext{FileReference: ctx}
	info := make(ResolutionInfo)
	c, err := r.resolveReference(systemID, path, rctx, ref, depth, info)
	if err != nil {
		return nil, err
	}

	rr := &ResolutionResult{
		Component: c,
		Info:      info,
	}
	return rr, nil
}

// ResolveComponentReferences fulfils the ComponentResolver interface.
func (r *DefaultComponentResolver) ResolveComponentReferences(
	c component.Interface,
	systemID v1.SystemID,
	path tree.Path,
	ctx *git.FileReference,
	depth int32,
) (*ResolutionResult, error) {
	// TODO(kevindrosendahl): this here is why private system definitions aren't supported
	rctx := &resolutionContext{FileReference: ctx}
	info := make(ResolutionInfo)
	c, err := r.resolveComponentReferences(c, systemID, path, rctx, depth, info)
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
	ctx *resolutionContext,
	ref *definitionv1.Reference,
	depth int32,
	info ResolutionInfo,
) (component.Interface, error) {
	if depth == 0 {
		info[path] = ResolutionNodeInfo{
			Commit:       ctx.FileReference.CommitReference,
			SSHKeySecret: ctx.SSHKeySecret,
		}
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

	return r.resolveComponentReferences(c, systemID, path, resolvedCxt, nextDepth, info)
}

func (r *DefaultComponentResolver) resolveComponentReferences(
	c component.Interface,
	systemID v1.SystemID,
	path tree.Path,
	ctx *resolutionContext,
	depth int32,
	info ResolutionInfo,
) (component.Interface, error) {
	// If the reference resolved to another reference, resolve that reference.
	// FIXME(kevinrosendahl): detect cycles
	if resolvedRef, ok := c.(*definitionv1.Reference); ok {
		return r.resolveReference(systemID, path, ctx, resolvedRef, depth, info)
	}

	info[path] = ResolutionNodeInfo{
		Component:    c,
		Commit:       ctx.FileReference.CommitReference,
		SSHKeySecret: ctx.SSHKeySecret,
	}

	// If the reference is not a system, then there's nothing more to do
	system, ok := c.(*definitionv1.System)
	if !ok {
		return c, nil
	}

	return r.resolveSystemComponents(systemID, path, ctx, system, depth, info)
}

func (r *DefaultComponentResolver) resolveSystemComponents(
	systemID v1.SystemID,
	path tree.Path,
	ctx *resolutionContext,
	system *definitionv1.System,
	depth int32,
	info ResolutionInfo,
) (*definitionv1.System, error) {
	for name, c := range system.Components {
		r.resolveComponentReferences(c, systemID, path.Child(name), ctx, depth, info)
	}

	return system, nil
}

func (r *DefaultComponentResolver) resolveTemplate(
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
		gitCtx.RepositoryURL = ctx.FileReference.RepositoryURL
		gitCtx.Options.SSHKey = ctx.SSHKey
		gitRef = &git.Reference{Commit: &ctx.FileReference.Commit}
		file = *ref.File
	}

	fileRef := &git.FileReference{
		CommitReference: git.CommitReference{
			RepositoryURL: gitCtx.RepositoryURL,
			Commit:        *gitRef.Commit,
		},
		File: file,
	}

	resolvedContext := &resolutionContext{
		FileReference: fileRef,
		SSHKey:        gitCtx.Options.SSHKey,
		SSHKeySecret:  sshKeySecret,
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

func (r *DefaultComponentResolver) gitReferenceCommit(
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
