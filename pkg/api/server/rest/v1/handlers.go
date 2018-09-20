package v1

import (
	backendv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"

	"github.com/gin-gonic/gin"
)

func MountHandlers(router *gin.RouterGroup, backend backendv1.Interface, resolver resolver.Interface) {
	api := newLatticeAPI(router, backend, resolver)
	api.setupAPI()
}
