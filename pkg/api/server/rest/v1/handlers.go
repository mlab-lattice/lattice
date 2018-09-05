package v1

import (
	backendv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"

	"github.com/gin-gonic/gin"
)

func MountHandlers(router *gin.RouterGroup, backend backendv1.Backend, resolver resolver.ComponentResolver) {
	mountSystemHandlers(router, backend, resolver)
}
