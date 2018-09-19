package bearertoken

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication"
	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication/user"
)

type Authenticator struct {
	token authentication.Token
}

func New(token authentication.Token) (*Authenticator, error) {
	authenticator := &Authenticator{
		token: token,
	}

	return authenticator, nil
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

	if ok {
		// remove header after successful auth
		c.Header("Authorization", "")
	}

	if !ok && err == nil {
		err = fmt.Errorf("invalid token: %v", err)
	}

	return u, ok, err
}
