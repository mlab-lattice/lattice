package language

import (
	"encoding/json"
	"fmt"

	"path"
	"regexp"
	"strings"

	"github.com/mlab-lattice/system/pkg/util/git"
)

// WORK_DIR
const WORK_DIR = "/tmp/lattice-core/git-file-repository"

func resolveUrl(url string, env *environment) (*urlResource, error) {
	// if its a git url then return a templateURLInfo for a git url
	if isGitUrl(url) {
		return resolveGitUrl(url, env)
	} else if isRelativeUrl(url) {
		return resolveRelativeUrl(url, env)

	}

	return nil, fmt.Errorf("Unsupported url %s", url)
}

// makeGitRepositoryFor constructs a git file repository for the specified url
func fetchGitFileContents(repoUrl string, fileName string, env *environment) ([]byte, error) {

	gitResolver, _ := git.NewResolver(WORK_DIR)
	gitOptions := env.options.gitOptions
	if gitOptions == nil {
		gitOptions = &git.Options{}
	}
	ctx := &git.Context{
		URI:     repoUrl,
		Options: gitOptions,
	}

	return gitResolver.FileContents(ctx, fileName)

}

// urlResource represents info needed when parsing a template url
type urlResource struct {
	baseUrl string
	data    map[string]interface{}
}

// parseGitTemplateUrl
func resolveGitUrl(url string, env *environment) (*urlResource, error) {
	if !isGitUrl(url) {
		return nil, fmt.Errorf("Invalid git url: '%s'", url)
	}

	parts := gitUrlRegex.FindAllStringSubmatch(url, -1)
	repoUri := parts[0][2]
	ref := parts[0][4]
	resourcePath := parts[0][7]

	baseUrl := path.Dir(url)

	cloneUri := repoUri + "#" + ref

	bytes, err := fetchGitFileContents(cloneUri, resourcePath, env)
	if err != nil {
		return nil, err
	}

	return newUrlResource(baseUrl, resourcePath, bytes)

}

func newUrlResource(baseUrl string, resourcePath string, bytes []byte) (*urlResource, error) {
	data, err := unmarshalBytes(bytes, resourcePath)
	if err != nil {
		return nil, err
	}

	return &urlResource{
		baseUrl: baseUrl,
		data:    data,
	}, nil
}
func resolveRelativeUrl(url string, env *environment) (*urlResource, error) {

	fullUrl := path.Join(env.currentFrame().baseUrl, url)

	return resolveUrl(fullUrl, env)
}

// regex for matching git file urls
var gitUrlRegex = regexp.MustCompile(`(?:git|file|ssh|https?|git@[-\w.]+):(//)?(.*.git)(#(([-\d\w._])+)?)?(/(.*))?$`)

// isGitTemplateUrl
func isGitUrl(url string) bool {
	return gitUrlRegex.MatchString(url)
}

func isRelativeUrl(url string) bool {
	return !strings.Contains(url, "://")
}

// unmarshalBytes unmarshal the bytes specified based on the the file name
func unmarshalBytes(bytes []byte, fileName string) (map[string]interface{}, error) {

	// unmarshal file contents based on file type. Only .json is supported atm

	result := make(map[string]interface{})

	if strings.HasSuffix(fileName, ".json") {
		err := json.Unmarshal(bytes, &result)

		if err != nil {
			return nil, err
		} else {
			return result, nil
		}
	} else {
		return nil, error(fmt.Errorf("Unsupported file %s", fileName))
	}

}
