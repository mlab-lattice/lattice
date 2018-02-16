package language

import (
	"encoding/json"
	"fmt"

	"regexp"
	"strings"

	"github.com/mlab-lattice/system/pkg/util/git"
)

// resolveURL reads the template file by resolving the url into a urlResource w
func resolveURL(url string, env *environment) (*urlResource, error) {
	// if its a git url then return a templateURLInfo for a git url
	if isGitURL(url) {
		return resolveGitURL(url, env)
	} else if isRelativeURL(url) {
		return resolveRelativeURL(url, env)

	}

	return nil, fmt.Errorf("Unsupported url %s", url)
}

// fetchGitFileContents fetches the specified git file contents
func fetchGitFileContents(repoURL string, fileName string, env *environment) ([]byte, error) {

	gitResolver, _ := git.NewResolver(env.options.WorkDirectory)
	gitOptions := env.options.GitOptions
	if gitOptions == nil {
		gitOptions = &git.Options{}
	}
	ctx := &git.Context{
		URI:     repoURL,
		Options: gitOptions,
	}

	return gitResolver.FileContents(ctx, fileName)

}

// urlResource artifact for url resolution
type urlResource struct {
	url      string
	baseURL  string
	fileName string
	bytes    []byte
	data     map[string]interface{}
}

// resolveGitURL resolves a git url
func resolveGitURL(url string, env *environment) (*urlResource, error) {
	if !isGitURL(url) {
		return nil, fmt.Errorf("Invalid git url: '%s'", url)
	}

	parts := gitURLRegex.FindAllStringSubmatch(url, -1)

	protocol := parts[0][1]
	repoPath := parts[0][3]
	ref := parts[0][5]
	fileName := parts[0][8]

	// reconstruct the url minus file path
	baseURL := fmt.Sprintf("%v://%v", protocol, repoPath)

	// append ref

	if ref != "" {
		baseURL = baseURL + "#" + ref
	}

	bytes, err := fetchGitFileContents(baseURL, fileName, env)
	if err != nil {
		return nil, err
	}

	return newURLResource(url, baseURL, fileName, bytes)

}

// newURLResource creates a new urlResource struct
func newURLResource(url, baseURL string, fileName string, bytes []byte) (*urlResource, error) {
	data, err := unmarshalBytes(bytes, fileName)
	if err != nil {
		return nil, err
	}

	return &urlResource{
		url:     url,
		baseURL: baseURL,
		bytes:   bytes,
		data:    data,
	}, nil
}

// resolveRelativeURL resolves a relative url by creating a full url with the existing url base
func resolveRelativeURL(url string, env *environment) (*urlResource, error) {

	// construct a full url using the existing baseURL in env
	fullURL := fmt.Sprintf("%v/%v", env.currentFrame().resource.baseURL, url)

	return resolveURL(fullURL, env)
}

// regex for matching git file urls
var gitURLRegex = regexp.MustCompile(`((?:git|file|ssh|https?|git@[-\w.]+)):(//)?(.*.git)(#(([-\d\w._])+)?)?(/(.*))?$`)

// isGitURL
func isGitURL(url string) bool {
	return gitURLRegex.MatchString(url)
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
		}

		return result, nil
	}

	return nil, error(fmt.Errorf("Unsupported file %s", fileName))

}
