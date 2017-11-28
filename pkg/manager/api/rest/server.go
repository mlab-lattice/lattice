package rest

import (
	"fmt"
	"net/http"

	systemresolver "github.com/mlab-lattice/core/pkg/system/resolver"

	"github.com/mlab-lattice/system/pkg/manager/backend"

	"github.com/gin-gonic/gin"
)

type restServer struct {
	router   *gin.Engine
	backend  backend.Interface
	resolver systemresolver.SystemResolver
}

func RunNewRestServer(b backend.Interface, port int32, workingDirectory string) {
	s := restServer{
		router:  gin.Default(),
		backend: b,
		resolver: systemresolver.SystemResolver{
			WorkDirectory: workingDirectory + "/resolver",
		},
	}

	s.mountHandlers()
	s.router.Run(fmt.Sprintf(":%v", port))
}

func (r *restServer) mountHandlers() {
	// Status
	r.router.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	r.mountNamespaceHandlers()
	r.mountAdminHandlers()
}
