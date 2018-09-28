package bearertoken

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication/authenticator"
	"github.com/mlab-lattice/lattice/pkg/api/server/user"
)

// Authenticator implementation for authentication.Request which authenticates requests based on bearer tokens
type Authenticator struct {
	token authenticator.Token
}

func New(token authenticator.Token) (*Authenticator, error) {
	return &Authenticator{token: token}, nil
}

func (authenticator *Authenticator) AuthenticateRequest(c *gin.Context) (user.User, bool, error) {
	// Check if there is an authorization header
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if auth == "" {
		return nil, false, nil
	}

	// grab the token
	parts := strings.Split(auth, " ")
	if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, false, nil
	}

	token := parts[1]

	if len(token) == 0 {
		return nil, false, fmt.Errorf("")
	}

	u, ok, err := authenticator.token.AuthenticateToken(token)

	if err != nil {
		return nil, false, err
	}

	if ok {
		// remove header after successful auth
		c.Header("Authorization", "")
	}
	return u, ok, nil
}
