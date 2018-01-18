package language

import (
	"encoding/json"
	"fmt"

	"path"
	"regexp"
	"strings"

	"github.com/mlab-lattice/system/pkg/util/git"
)

// resolveUrl reads the template file by resolving the url into a urlResource w
func resolveUrl(url string, env *environment) (*urlResource, error) {
	// if its a git url then return a templateURLInfo for a git url
	if isGitURL(url) {
		return resolveGitURL(url, env)
	} else if isRelativeURL(url) {
		return resolveRelativeURL(url, env)

	}

	return nil, fmt.Errorf("Unsupported url %s", url)
}

// fetchGitFileContents fetches the specified git file contents
func fetchGitFileContents(repoUrl string, fileName string, env *environment) ([]byte, error) {

	gitResolver, _ := git.NewResolver(env.options.WorkDirectory)
	gitOptions := env.options.GitOptions
	if gitOptions == nil {
		gitOptions = &git.Options{}
	}
	ctx := &git.Context{
		URI:     repoUrl,
		Options: gitOptions,
	}

	return gitResolver.FileContents(ctx, fileName)

}

// urlResource artifact for url resolution
type urlResource struct {
	baseUrl string
	data    map[string]interface{}
}

// resolveGitURL resolves a git url
func resolveGitURL(url string, env *environment) (*urlResource, error) {
	if !isGitURL(url) {
		return nil, fmt.Errorf("Invalid git url: '%s'", url)
	}

	parts := gitUrlRegex.FindAllStringSubmatch(url, -1)

	protocol := parts[0][1]
	repoPath := parts[0][3]
	ref := parts[0][5]
	resourcePath := parts[0][8]

	// reconstruct the url minus file path
	baseUrl := fmt.Sprintf("%v://%v", protocol, repoPath)

	// append ref

	if ref != "" {
		baseUrl = baseUrl + "#" + ref
	}

	bytes, err := fetchGitFileContents(baseUrl, resourcePath, env)
	if err != nil {
		return nil, err
	}

	return newUrlResource(baseUrl, resourcePath, bytes)

}

// newUrlResource creates a new urlResource struct
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

// resolveRelativeURL resolves a relative url by creating a full url with the existing url base
func resolveRelativeURL(url string, env *environment) (*urlResource, error) {

	// construct a full url using the existing baseUrl in env
	fullUrl := path.Join(env.currentFrame().baseUrl, url)

	return resolveUrl(fullUrl, env)
}

// regex for matching git file urls
var gitUrlRegex = regexp.MustCompile(`((?:git|file|ssh|https?|git@[-\w.]+)):(//)?(.*.git)(#(([-\d\w._])+)?)?(/(.*))?$`)

// isGitURL
func isGitURL(url string) bool {
	return gitUrlRegex.MatchString(url)
}

// isRelativeURL
func isRelativeURL(url string) bool {
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
