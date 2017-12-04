package git

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func GetGitURIFromComponentBuild(gitRepo *block.GitRepository) (string, error) {
	if gitRepo == nil {
		return "", fmt.Errorf("cannot get git URI from nil component build")
	}

	if err := gitRepo.Validate(nil); err != nil {
		return "", fmt.Errorf("invalid component build git_repository: %v", err.Error())
	}

	uri := gitRepo.URL
	if gitRepo.Commit != nil {
		uri += "#" + *gitRepo.Commit
	} else {
		uri += "#" + *gitRepo.Tag
	}
	return uri, nil
}
