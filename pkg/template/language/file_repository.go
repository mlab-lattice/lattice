package language

import (
	"fmt"
	"github.com/mlab-lattice/system/pkg/util/git"
	"regexp"
)

const WORK_DIR = "/tmp/lattice-core/git-file-repository"

// FileRepository interface
type FileRepository interface {
	getFileContents(fileName string) ([]byte, error)
}

// GitRepository FileResolver implementation for git
type GitRepository struct {
	gitResolverContext *git.Context
	gitResolver        *git.Resolver
}

// GitRepository FileResolver implementation for git
type templateURLInfo struct {
	url            string
	fileRepository FileRepository
	filePath       string
}

func (this GitRepository) getFileContents(fileName string) ([]byte, error) {

	return this.gitResolver.FileContents(this.gitResolverContext, fileName)
}

func makeGitRepositoryFor(url string) (FileRepository, error) {

	gitResolver, _ := git.NewResolver(WORK_DIR)
	gitResolverContext := &git.Context{
		URI: url,
	}

	repository := &GitRepository{
		gitResolver:        gitResolver,
		gitResolverContext: gitResolverContext,
	}

	return repository, nil

}

func parseTemplateUrl(url string) (*templateURLInfo, error) {
	if isGitTemplateUrl(url) {
		return parseGitTemplateUrl(url)
	}

	return relativeTemplateUrl(url)
	//return nil, fmt.Errorf("Unsupported url '%s'", url)
}

func parseGitTemplateUrl(url string) (*templateURLInfo, error) {
	if !isGitTemplateUrl(url) {
		return nil, fmt.Errorf("Invalid git url: '%s'", url)
	}

	parts := gitUrlRegex.FindAllStringSubmatch(url, -1)
	repoUri := parts[0][2]
	ref := parts[0][4]
	filePath := parts[0][7]

	cloneUri := repoUri + "#" + ref

	gitRepo, err := makeGitRepositoryFor(cloneUri)

	if err != nil {
		return nil, err
	}

	result := &templateURLInfo{
		url:            url,
		fileRepository: gitRepo,
		filePath:       filePath,
	}

	return result, nil
}

func relativeTemplateUrl(url string) (*templateURLInfo, error) {
	result := &templateURLInfo{
		url:      url,
		filePath: url,
	}

	return result, nil
}

var gitUrlRegex = regexp.MustCompile(`(?:git|file|ssh|https?|git@[-\w.]+):(//)?(.*.git)(#(([-\d\w._])+)?)?(/(.*))?$`)

func isGitTemplateUrl(url string) bool {
	return gitUrlRegex.MatchString(url)
}
