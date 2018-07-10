package containerbuilder

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"github.com/fatih/color"
)

func (b *Builder) buildGitRepositoryComponent() error {
	color.Blue("Cloning git repository...")

	if b.StatusUpdater != nil {
		// For now ignore status update errors, don't need to fail a build because the status could
		// not be updated.
		b.StatusUpdater.UpdateProgress(b.BuildID, b.SystemID, v1.ContainerBuildPhasePullingGitRepository)
	}

	gitResolver, err := git.NewResolver(b.WorkingDir + "/git")
	if err != nil {
		return newErrorInternal("failed to create git resolver: " + err.Error())
	}

	ctx := &git.Context{
		Resource: git.Resource{
			RepositoryURL: b.ContainerBuild.GitRepository.URL,
			Commit:        b.ContainerBuild.GitRepository.Commit,
		},
		Options: b.GitOptions,
	}
	if err := gitResolver.Checkout(ctx); err != nil {
		return newErrorUser("git repository checkout failed: " + err.Error())
	}

	color.Green("✓ Success!")
	fmt.Println()

	sourceDirectory := gitResolver.GetRepositoryPath(ctx)
	return b.buildDockerImage(sourceDirectory)
}
