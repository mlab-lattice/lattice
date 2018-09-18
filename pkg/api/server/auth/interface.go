package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/users"
)

type Authenticator interface {
	AuthenticateRequest(c *gin.Context) (users.User, error)
}
