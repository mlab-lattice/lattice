package authenticator

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/user"
)

// Token interface for a token authenticator
type Token interface {
	// AuthenticateToken returns a user object, ok to indicate success, and error if any
	AuthenticateToken(token string) (user.User, bool, error)
}
