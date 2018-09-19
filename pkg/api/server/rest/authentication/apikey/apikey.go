package apikey

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication/user"
)

const (
	apiKeyHeader = "API_KEY"
)

// Global legacy api key user
var legacyApiKeyUser = &user.DefaultUser{Username: "legacyApiKeyUser"}

type Authenticator struct {
	apiKey string
}

func New(apiKey string) *Authenticator {
	return &Authenticator{
		apiKey: apiKey,
	}
}

func (authenticator *Authenticator) AuthenticateRequest(c *gin.Context) (user.User, bool, error) {
	// grab request API key from header
	requestAPIKey := c.Request.Header.Get(apiKeyHeader)
	if requestAPIKey == "" {
		return nil, false, nil
	} else if requestAPIKey != authenticator.apiKey {
		return nil, false, fmt.Errorf("invalid '%v'", apiKeyHeader)
	}

	// Authentication success!!
	return legacyApiKeyUser, true, nil
}
