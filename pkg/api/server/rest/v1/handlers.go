package v1

import (
	serverv1 "github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"

	"github.com/gin-gonic/gin"
)

func MountHandlers(router *gin.RouterGroup, backend serverv1.Interface, sysResolver resolver.SystemResolver) {
	api := newLatticeAPI(router, backend, sysResolver)
	api.setupAPI()
}
