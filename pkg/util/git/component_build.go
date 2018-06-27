package git

import (
	"fmt"

	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

func GetGitURIFromDefinition(gitRepo *definitionv1.GitRepository) (string, error) {
	if gitRepo == nil {
		return "", fmt.Errorf("cannot get git URI from nil component build")
	}

	uri := gitRepo.URL
	if gitRepo.Commit != nil {
		uri += "#" + *gitRepo.Commit
	} else {
		uri += "#" + *gitRepo.Tag
	}
	return uri, nil
}
