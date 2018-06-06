package language

import (
	"encoding/json"
	"fmt"

	"regexp"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/util/git"
)

// readTemplateFromURL reads the template file by resolving the url into a Template w
func readTemplateFromURL(url string, env *environment) (*Template, error) {
	// if its a git url then return a templateURLInfo for a git url
	if isGitTemplateURL(url) {
		return readTemplateFromGitURL(url, env)
	}

	return nil, fmt.Errorf("unsupported url %s", url)
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

// Template artifact for url resolution
type Template struct {
	url      string
	baseURL  string
	fileName string
	bytes    []byte
	data     map[string]interface{}
}

// readTemplateFromGitURL resolves a git url
func readTemplateFromGitURL(url string, env *environment) (*Template, error) {
	if !isGitTemplateURL(url) {
		return nil, fmt.Errorf("invalid git url: '%s'", url)
	}

	parts := gitTemplateURLRegex.FindAllStringSubmatch(url, -1)

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

	return newTemplate(url, baseURL, fileName, bytes)

}

// newTemplate creates a new Template struct
func newTemplate(url, baseURL string, fileName string, bytes []byte) (*Template, error) {
	data, err := unmarshalBytes(bytes, fileName)
	if err != nil {
		return nil, err
	}

	return &Template{
		url:     url,
		baseURL: baseURL,
		bytes:   bytes,
		data:    data,
	}, nil
}

// regex for matching git file urls
var gitTemplateURLRegex = regexp.MustCompile(`((?:git|file|ssh|https?|git@[-\w.]+)):(//)?(.*.git)(#(([-\d\w._])+)?)?(/(.*))?$`)

// isGitTemplateURL
func isGitTemplateURL(url string) bool {
	return gitTemplateURLRegex.MatchString(url)
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

	return nil, error(fmt.Errorf("unsupported file %s", fileName))

}
