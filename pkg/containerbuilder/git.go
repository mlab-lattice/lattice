package containerbuilder

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"github.com/fatih/color"
)

func (b *Builder) retrieveGitRepository(repository *definitionv1.GitRepository) (string, error) {
	color.Blue("Cloning git repository...")

	if b.StatusUpdater != nil {
		// For now ignore status update errors, don't need to fail a build because the status could
		// not be updated.
		b.StatusUpdater.UpdateProgress(b.BuildID, b.SystemID, v1.ContainerBuildPhasePullingGitRepository)
	}

	gitResolver, err := git.NewResolver(b.WorkingDir+"/git", false)
	if err != nil {
		return "", newErrorInternal("failed to create git resolver: " + err.Error())
	}

	ctx := &git.Context{
		RepositoryURL: repository.URL,
		Options:       b.GitOptions,
	}

	ref := &git.Reference{Commit: repository.Commit}

	if err := gitResolver.Checkout(ctx, ref); err != nil {
		return "", newErrorUser("git repository checkout failed: " + err.Error())
	}

	color.Green("✓ Success!")
	fmt.Println()

	return gitResolver.RepositoryPath(repository.URL), nil
}
