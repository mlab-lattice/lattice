package resolver

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/util/git"
)

// SystemResolver resolves system definitions from different sources such as git
type SystemResolver struct {
	WorkDirectory string
	GitResolver   *git.Resolver
}

// GitResolveOptions allows options for resolution
type GitResolveOptions struct {
	SSHKey []byte
}

type resolveContext struct {
	gitURI            string
	gitResolveOptions *GitResolveOptions
}

func NewSystemResolver(workDirectory string) (*SystemResolver, error) {
	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	gitResolver, err := git.NewResolver(workDirectory + "/git")
	if err != nil {
		return nil, err
	}

	sr := &SystemResolver{
		WorkDirectory: workDirectory,
		GitResolver:   gitResolver,
	}
	return sr, nil
}

// resolves the definition
func (resolver *SystemResolver) ResolveDefinition(uri string, fileName string, gitResolveOptions *GitResolveOptions) (tree.Node, error) {
	if gitResolveOptions == nil {
		gitResolveOptions = &GitResolveOptions{}
	}
	ctx := &resolveContext{
		gitURI:            uri,
		gitResolveOptions: gitResolveOptions,
	}
	return resolver.resolveDefinitionFromGitUri(ctx, fileName)
}

// lists the versions of the specified definition's uri
func (resolver *SystemResolver) ListDefinitionVersions(uri string, gitResolveOptions *GitResolveOptions) ([]string, error) {
	if gitResolveOptions == nil {
		gitResolveOptions = &GitResolveOptions{}
	}
	ctx := &resolveContext{
		gitURI:            uri,
		gitResolveOptions: gitResolveOptions,
	}
	return resolver.listRepoVersionTags(ctx)

}

// resolveDefinitionFromGitUri resolves a definition from a git uri
func (resolver *SystemResolver) resolveDefinitionFromGitUri(ctx *resolveContext, fileName string) (tree.Node, error) {
	return resolver.readNodeFromFile(ctx, fileName)
}

// readNodeFromFile reads a definition node from a file
func (resolver *SystemResolver) readNodeFromFile(ctx *resolveContext, fileName string) (tree.Node, error) {
	jsonMap, err := resolver.readConsolidatedJsonMapFromFile(ctx, fileName)

	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	defInterface, err := definition.NewFromJSON(jsonBytes)

	if err != nil {
		return nil, err
	}

	return tree.NewNode(defInterface, nil)
}

// reads/consolidates json map form a file.
func (resolver *SystemResolver) readConsolidatedJsonMapFromFile(ctx *resolveContext, fileName string) (map[string]interface{}, error) {
	gitResolverContext := &git.Context{
		URI:    ctx.gitURI,
		SSHKey: ctx.gitResolveOptions.SSHKey,
	}
	jsonBytes, err := resolver.GitResolver.FileContents(gitResolverContext, fileName)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	err = json.Unmarshal(jsonBytes, &result)

	if err != nil {
		return nil, err
	}

	// resolve json and bytes
	resolver.resolveJsonMap(result, ctx)

	return result, nil
}

// resolveJsonMap resolves a whole json map
func (resolver *SystemResolver) resolveJsonMap(jsonMap map[string]interface{}, ctx *resolveContext) error {
	for k, v := range jsonMap {
		var err error
		jsonMap[k], err = resolver.resolveJsonValue(k, v, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// resolveJsonValue resolves a single json value. i.e. deals with special values such as $include
func (resolver *SystemResolver) resolveJsonValue(k string, v interface{}, ctx *resolveContext) (interface{}, error) {
	if valMap, valueIsMap := v.(map[string]interface{}); valueIsMap {
		if includeVal, isInclude := valMap["$include"]; isInclude {
			includeFilePath, ok := includeVal.(string)
			if ok {
				return resolver.readConsolidatedJsonMapFromFile(ctx, includeFilePath)
			} else {
				panic("Invalid $include")
			}
		} else {
			resolver.resolveJsonMap(valMap, ctx)
			return valMap, nil
		}
	} else if valArr, valueIsArray := v.([]interface{}); valueIsArray {
		for i, item := range valArr {
			var err error
			valArr[i], err = resolver.resolveJsonValue(k, item, ctx)
			if err != nil {
				return nil, err
			}
		}
		return valArr, nil
	}

	return v, nil
}

// lists the tags in a repo
func (resolver *SystemResolver) listRepoVersionTags(ctx *resolveContext) ([]string, error) {
	gitResolverContext := &git.Context{
		URI:    ctx.gitURI,
		SSHKey: ctx.gitResolveOptions.SSHKey,
	}
	return resolver.GitResolver.GetTagNames(gitResolverContext)
}
