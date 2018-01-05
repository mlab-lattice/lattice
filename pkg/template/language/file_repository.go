package language

import (
	"fmt"
	"regexp"

	"github.com/mlab-lattice/system/pkg/util/git"
)

// WORK_DIR
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

func (gitRepository GitRepository) getFileContents(fileName string) ([]byte, error) {

	return gitRepository.gitResolver.FileContents(gitRepository.gitResolverContext, fileName)
}

// makeGitRepositoryFor constructs a git file repository for the specified url
func makeGitRepositoryFor(url string, env *environment) (FileRepository, error) {

	gitResolver, _ := git.NewResolver(WORK_DIR)
	gitOptions := env.gitOptions
	if gitOptions == nil {
		gitOptions = &git.Options{}
	}
	gitResolverContext := &git.Context{
		URI:     url,
		Options: gitOptions,
	}

	repository := &GitRepository{
		gitResolver:        gitResolver,
		gitResolverContext: gitResolverContext,
	}

	return repository, nil

}

// templateURLInfo represents info needed when parsing a template url
type templateURLInfo struct {
	url            string
	fileRepository FileRepository
	filePath       string
}

// parseTemplateUrl parses the url and returns a templateURLInfo
func parseTemplateUrl(url string, env *environment) (*templateURLInfo, error) {
	// if its a git url then return a templateURLInfo for a git url
	if isGitTemplateUrl(url) {
		return parseGitTemplateUrl(url, env)
	}

	// otherwhise, always assume its a relative url for the current repository
	return relativeTemplateUrl(url)
}

// parseGitTemplateUrl
func parseGitTemplateUrl(url string, env *environment) (*templateURLInfo, error) {
	if !isGitTemplateUrl(url) {
		return nil, fmt.Errorf("Invalid git url: '%s'", url)
	}

	parts := gitUrlRegex.FindAllStringSubmatch(url, -1)
	repoUri := parts[0][2]
	ref := parts[0][4]
	filePath := parts[0][7]

	cloneUri := repoUri + "#" + ref

	gitRepo, err := makeGitRepositoryFor(cloneUri, env)

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

// relativeTemplateUrl returns a relative templateURLInfo
func relativeTemplateUrl(url string) (*templateURLInfo, error) {
	result := &templateURLInfo{
		url:      url,
		filePath: url,
	}

	return result, nil
}

// regex for matching git file urls
var gitUrlRegex = regexp.MustCompile(`(?:git|file|ssh|https?|git@[-\w.]+):(//)?(.*.git)(#(([-\d\w._])+)?)?(/(.*))?$`)

// isGitTemplateUrl
func isGitTemplateUrl(url string) bool {
	return gitUrlRegex.MatchString(url)
}
