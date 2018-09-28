package authentication

import (
	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/server/user"
)

// Request interface for an authenticator that authenticates requests
type Request interface {
	// AuthenticateRequest returns user, success, error if any
	AuthenticateRequest(c *gin.Context) (user.User, bool, error)
}
