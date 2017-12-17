package language

import (
	"github.com/mlab-lattice/system/pkg/util/git"
)

// FileResolver interface
type FileResolver interface {
	FileContents(fileName string) ([]byte, error)
}

// GitResolverWrapper FileResolver implementation for git
type GitResolverWrapper struct {
	GitResolverContext *git.Context
	GitResolver        *git.Resolver
}

func (gitWrapper GitResolverWrapper) FileContents(fileName string) ([]byte, error) {
	return gitWrapper.GitResolver.FileContents(gitWrapper.GitResolverContext, fileName)
}
