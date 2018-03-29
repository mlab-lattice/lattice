package v1

import (
	serverv1 "github.com/mlab-lattice/system/pkg/api/server/v1"
	"github.com/mlab-lattice/system/pkg/definition/resolver"

	"github.com/gin-gonic/gin"
)

func MountHandlers(router *gin.Engine, backend serverv1.Interface, sysResolver *resolver.SystemResolver) {
	mountSystemHandlers(router, backend, sysResolver)
}