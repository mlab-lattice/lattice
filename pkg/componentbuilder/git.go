package componentbuilder

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/git"

	"github.com/fatih/color"
)

func (b *Builder) buildGitRepositoryComponent() error {
	color.Blue("Cloning git repository...")

	if b.StatusUpdater != nil {
		// For now ignore status update errors, don't need to fail a build because the status could
		// not be updated.
		b.StatusUpdater.UpdateProgress(b.BuildID, b.SystemID, types.ComponentBuildPhasePullingGitRepository)
	}

	gitRepo := b.ComponentBuildBlock.GitRepository
	if err := gitRepo.Validate(nil); err != nil {
		return newErrorUser("invalid git repository config: " + err.Error())
	}

	gitResolver, err := git.NewResolver(b.WorkingDir + "/git")
	if err != nil {
		return newErrorInternal("failed to create git resolver: " + err.Error())
	}
	b.GitResolver = gitResolver

	uri, err := git.GetGitURIFromComponentBuild(gitRepo)
	if err != nil {
		return newErrorInternal("failed to get git URI from component build: " + err.Error())
	}

	if err = b.checkOutGitRepository(uri); err != nil {
		return newErrorUser("git repository checkout failed: " + err.Error())
	}

	color.Green("✓ Success!")
	fmt.Println()

	sourceDirectory := b.GitResolver.GetRepositoryPath(b.getGitResolverContext(uri))
	return b.buildDockerImage(sourceDirectory)
}

func (b *Builder) checkOutGitRepository(uri string) error {
	return b.GitResolver.Checkout(b.getGitResolverContext(uri))
}

func (b *Builder) getGitResolverContext(uri string) *git.Context {
	return &git.Context{
		Options: &git.Options{SSHKey: b.GitResolverOptions.SSHKey},
		URI:     uri,
	}
}
