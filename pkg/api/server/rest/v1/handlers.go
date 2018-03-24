package v1

import (
	"github.com/mlab-lattice/system/pkg/api/server/rest/v1/system"
	"github.com/mlab-lattice/system/pkg/api/server/v1"
	"github.com/mlab-lattice/system/pkg/definition/resolver"

	"github.com/gin-gonic/gin"
)

func MountHandlers(router *gin.Engine, backend v1.Interface, sysResolver *resolver.SystemResolver) {
	system.MountHandlers(router, backend, sysResolver)
}
