package auth

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mlab-lattice/lattice/pkg/api/users"
)

type BearerTokenAuthenticator struct {
	authTokenMap map[string]users.User
}

func NewBearerTokenAuthenticator(tokenFile io.ReadCloser) (*BearerTokenAuthenticator, error) {
	authenticator := &BearerTokenAuthenticator{
		authTokenMap: nil,
	}

	authenticator.readAuthTokenMap(tokenFile)

	return authenticator, nil
}

func (authenticator *BearerTokenAuthenticator) AuthenticateRequest(c *gin.Context) (users.User, error) {
	// Check if there is an authorization header
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if auth == "" {
		return nil, nil
	}

	// grab the token
	parts := strings.Split(auth, " ")
	if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, nil
	}

	token := parts[1]

	if len(token) == 0 {
		return nil, fmt.Errorf("")
	}

	user, err := authenticator.authenticateToken(token)

	if user != nil {
		// remove header after succesful auth
		c.Header("Authorization", "")
		return user, nil
	}

	return nil, fmt.Errorf("invalid token: %v", err)
}

func (authenticator *BearerTokenAuthenticator) authenticateToken(token string) (users.User, error) {
	if user, exists := authenticator.authTokenMap[token]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("unauthorized token")
}

func (authenticator *BearerTokenAuthenticator) readAuthTokenMap(tokenFile io.ReadCloser) error {
	defer tokenFile.Close()

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(tokenFile)
	if err != nil {
		return err
	}

	tokensCsvString := buf.String()
	tokenMap := make(map[string]users.User)
	for _, line := range strings.Split(tokensCsvString, "\n") {
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return fmt.Errorf("bad authentication-token-map values")
		}
		token := parts[0]
		username := parts[1]
		tokenMap[token] = &users.DefaultUser{
			Username: username,
		}
	}

	authenticator.authTokenMap = tokenMap
	return nil
}
