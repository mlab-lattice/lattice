package authentication

import (
	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication/user"
)

type Request interface {
	AuthenticateRequest(c *gin.Context) (user.User, bool, error)
}

type Token interface {
	AuthenticateToken(token string) (user.User, bool, error)
}
