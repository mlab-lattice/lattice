package resolver

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/template/language"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

type ComponentResolver interface {
	Resolve(*Context) (component.Interface, error)
}

type Context struct {
	Git *git.Context
}

// DefaultComponentResolver resolves system definitions from different sources such as git
type DefaultComponentResolver struct {
	gitResolver *git.Resolver
}

type resolveContext struct {
	gitURI            string
	gitResolveOptions *git.Options
}

func NewComponentResolver(workDirectory string) (ComponentResolver, error) {
	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	gitResolver, err := git.NewResolver(workDirectory + "/git")
	if err != nil {
		return nil, err
	}

	r := &DefaultComponentResolver{
		gitResolver: gitResolver,
	}
	return r, nil
}

func (r *DefaultComponentResolver) Resolve(ctx *Context) (component.Interface, error) {

	return nil, nil
}

// resolves the definition
func (resolver *DefaultComponentResolver) ResolveDefinition(uri string, gitResolveOptions *git.Options) (tree.Node, error) {

	if gitResolveOptions == nil {
		gitResolveOptions = &git.Options{}
	}
	ctx := &resolveContext{
		gitURI:            uri,
		gitResolveOptions: gitResolveOptions,
	}

	return resolver.readNodeFromFile(ctx)
}

// lists the versions of the specified definition's
func (r *DefaultComponentResolver) ListDefinitionVersions(ctx *Context) ([]string, error) {
	return r.gitResolver.GetTagNames(ctx.Git)
}

// readNodeFromFile reads a definition node from a file
func (resolver *DefaultComponentResolver) readNodeFromFile(ctx *resolveContext) (tree.Node, error) {
	engine := language.NewEngine()

	options, err := language.CreateOptions(resolver.gitResolver.WorkDirectory, ctx.gitResolveOptions)
	if err != nil {
		return nil, err
	}

	result, err := engine.EvalFromURL(ctx.gitURI, make(map[string]interface{}), options)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(result.ValueAsMap())
	if err != nil {
		return nil, err
	}

	def, err := definitionv1.NewComponentFromJSON(jsonBytes)
	if err != nil {
		return nil, err
	}

	return definitionv1.NewNode(def, "", nil)
}
