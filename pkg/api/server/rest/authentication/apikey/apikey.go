package apikey

import (
	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/server/user"
)

const (
	apiKeyHeader = "API_KEY"
)

// Global legacy api key user
var legacyAPIKeyUser = user.NewDefaultUser("legacyApiKeyUser")

// Authenticator implementation for authentication.Request which authenticates requests based on API_KEY header
type Authenticator struct {
	// apiKey to authenticate against
	apiKey string
}

// New
func New(apiKey string) *Authenticator {
	return &Authenticator{
		apiKey: apiKey,
	}
}

func (a *Authenticator) AuthenticateRequest(c *gin.Context) (user.User, bool, error) {
	// grab request API key from header
	requestAPIKey := c.Request.Header.Get(apiKeyHeader)
	if requestAPIKey == "" {
		return nil, false, nil
	} else if requestAPIKey != a.apiKey {
		return nil, false, nil
	}

	// Authentication success!!
	return legacyAPIKeyUser, true, nil
}
