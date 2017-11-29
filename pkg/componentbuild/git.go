package componentbuild

import (
	"fmt"

	gitutil "github.com/mlab-lattice/core/pkg/util/git"
)

func (b *Builder) buildGitRepositoryComponent() error {
	fmt.Println("Cloning git repository...")

	gitRepo := b.ComponentBuildBlock.GitRepository
	if err := gitRepo.Validate(nil); err != nil {
		return newErrorUser("invalid git repository config: " + err.Error())
	}

	gitResolver, err := gitutil.NewResolver(b.WorkingDir + "/git")
	if err != nil {
		return newErrorInternal("failed to create git resolver: " + err.Error())
	}
	b.GitResolver = gitResolver

	uri, err := gitutil.GetGitURIFromComponentBuild(gitRepo)
	if err != nil {
		return newErrorInternal("failed to get git URI from component build: " + err.Error())
	}

	if err = b.checkOutGitRepository(uri); err != nil {
		return newErrorUser("git repository checkout failed: " + err.Error())
	}

	sourceDirectory := b.GitResolver.GetRepositoryPath(b.getGitResolverContext(uri))
	return b.buildDockerImage(sourceDirectory)
}

func (b *Builder) checkOutGitRepository(uri string) error {
	return b.GitResolver.Checkout(b.getGitResolverContext(uri))
}

func (b *Builder) getGitResolverContext(uri string) *gitutil.Context {
	return &gitutil.Context{
		SSHKey: b.GitResolverOptions.SSHKey,
		URI:    uri,
	}
}
