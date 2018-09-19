package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/users"
)

const (
	apiKeyHeader = "API_KEY"
)

// Global legacy api key user
var legacyApiUser = &users.DefaultUser{Username: "legacyApiUser"}

type LegacyApiKeyAuthenticator struct {
	apiKey string
}

func NewLegacyApiKeyAuthenticator(apiKey string) *LegacyApiKeyAuthenticator {
	return &LegacyApiKeyAuthenticator{
		apiKey: apiKey,
	}
}

func (authenticator *LegacyApiKeyAuthenticator) AuthenticateRequest(c *gin.Context) (users.User, error) {
	// grab request API key from header
	requestAPIKey := c.Request.Header.Get(apiKeyHeader)
	if requestAPIKey == "" {
		return nil, nil
	} else if requestAPIKey != authenticator.apiKey {
		return nil, fmt.Errorf("invalid '%v'", apiKeyHeader)
	}

	// Authentication success!!
	return legacyApiUser, nil
}
